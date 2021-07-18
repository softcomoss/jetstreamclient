package jetstream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/softcomoss/jetstreamclient"
	"github.com/softcomoss/jetstreamclient/options"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type natsStore struct {
	opts        options.Options
	natsClient  *nats.Conn
	jsmClient   nats.JetStreamContext
	mu          *sync.RWMutex
	subjects    []string
	serviceName string
}

func (s *natsStore) Publish(topic string, message []byte) error {
	s.registerSubjectOnStream(topic)
	ak, err := s.jsmClient.Publish(topic, message)

	PrettyJson(ak)
	return err
}

func (s *natsStore) Subscribe(topic string, handler jetstreamclient.SubscriptionHandler, opts ...*options.SubscriptionOptions) error {
	s.registerSubjectOnStream(topic)

	sub := new(nats.Subscription)
	consumer := new(nats.ConsumerInfo)
	so := options.MergeSubscriptionOptions(opts...)
	cbHandler := func(m *nats.Msg) { event := newEvent(m); go handler(event) }
	var err error

	jsmSubscriptionOptions := make([]nats.SubOpt, 0)
	jsmSubscriptionOptions = append(jsmSubscriptionOptions,
		nats.DeliverLast(),
		//nats.EnableFlowControl(),
		nats.BindStream(s.opts.ServiceName),
		nats.MaxDeliver(5),
		nats.ReplayOriginal())

	DurableName := strings.ReplaceAll(fmt.Sprintf("%s-%s", s.opts.ServiceName, topic), ".", "-")

	ConsumerConfig := &nats.ConsumerConfig{
		//DeliverSubject: nats.NewInbox(),
		DeliverPolicy: nats.DeliverLastPolicy,
		AckPolicy:     nats.AckExplicitPolicy,
		MaxDeliver:    5,
		ReplayPolicy:  nats.ReplayOriginalPolicy,
		MaxAckPending: 20000,
		FlowControl:   true,
		//AckWait:         0,
		//RateLimit:       0,
		//Heartbeat:       0,
	}

	//nats.MaxAckPending(20000000)
	//nats.Durable(durableName)

	//nats.AckNone(),
	//nats.ManualAck(),
	switch so.GetSubscriptionType() {
	case options.Failover:

		break
	case options.Exclusive:
		ConsumerConfig.Durable = DurableName
		jsmSubscriptionOptions = append(jsmSubscriptionOptions, nats.Durable(DurableName))
		if consumer, err = s.mountConsumer(s.opts.ServiceName, ConsumerConfig); err != nil {
			return err
		}
		if sub, err = s.jsmClient.Subscribe(consumer.Name, cbHandler, jsmSubscriptionOptions...); err != nil {
			return err
		}
		break

	case options.Shared:
		//jsmSubscriptionOptions = append(jsmSubscriptionOptions,
		//	nats.ManualAck(),
		//	nats.AckExplicit(),
		//	//nats.EnableFlowControl(),
		//	nats.Durable(DurableName),
		//)

		//ConsumerConfig.Durable = fmt.Sprintf("%s-%s", DurableName, jetstreamclient.GenerateRandomString())
		////ConsumerConfig.AckPolicy = nats.AckNonePolicy
		////ConsumerConfig.DeliverSubject = nats.NewInbox()
		////ConsumerConfig.MaxAckPending = 0
		////ConsumerConfig.AckPolicy = nats.AckNonePolicy
		//if consumer, err = s.mountConsumer(s.opts.ServiceName, ConsumerConfig); err != nil {
		//	return err
		//}

		if sub, err = s.jsmClient.QueueSubscribe(topic, DurableName, cbHandler,
			nats.Durable(DurableName),
			nats.DeliverLast(),
			nats.EnableFlowControl(),
			nats.BindStream(s.opts.ServiceName),
			nats.MaxAckPending(20000000),
			nats.ManualAck(),
			nats.ReplayOriginal(),
			nats.MaxDeliver(5)); err != nil {
			return err
		}

		break
	case options.KeyShared:
		if sub, err = s.jsmClient.Subscribe(topic, cbHandler,
			nats.Durable(DurableName),
			nats.DeliverLast(),
			nats.EnableFlowControl(),
			nats.BindStream(s.opts.ServiceName),
			nats.MaxAckPending(20000000),
			nats.ManualAck(),
			nats.ReplayOriginal(),
			nats.MaxDeliver(5)); err != nil {
			return err
		}
	}

	//keySub := func(sub *nats.Subscription, errChan chan<- error) {
	//	var err error
	//	,
	//		//nats.Durable(durableName),
	//		nats.DeliverLast(),
	//		nats.EnableFlowControl(),
	//		nats.BindStream(s.opts.ServiceName),
	//		//nats.AckExplicit(),
	//		nats.ManualAck(),
	//		nats.ReplayOriginal(),
	//		nats.MaxAckPending(20000000),
	//		nats.MaxDeliver(5))
	//
	//	if err != nil {
	//		errChan <- err
	//	}
	//}

	select {
	case <-so.GetContext().Done():
		fmt.Print("Done draining...")
		return sub.Drain()
	}
}

func (s *natsStore) GetServiceName() string {
	return s.serviceName
}

func Init(opts options.Options, options ...nats.Option) (jetstreamclient.EventStore, error) {

	addr := strings.TrimSpace(opts.Address)
	if addr == "" {
		return nil, jetstreamclient.ErrInvalidURL
	}

	name := strings.TrimSpace(opts.ServiceName)
	if name == "" {
		return nil, jetstreamclient.ErrEmptyStoreName
	}
	opts.ServiceName = strings.TrimSpace(name)
	options = append(options, nats.Name(name))
	if opts.AuthenticationToken != "" {
		options = append(options, nats.Token(opts.AuthenticationToken))
	}

	//if opts.CertContent != "" {
	//	certPath, err := initCert(opts.CertContent)
	//	if err != nil {
	//		return nil, err
	//	}
	//	options = append(options, nats.ClientCert(certPath, ""))
	//}

	nc, js, err := connect(opts.ServiceName, opts.Address, options)

	if err != nil {
		return nil, err
	}

	//log.Print(sinfo.Config, " Stream Config \n")

	return &natsStore{
		opts:        opts,
		jsmClient:   js,
		natsClient:  nc,
		serviceName: name,
		subjects:    make([]string, 0),
		mu:          &sync.RWMutex{},
	}, nil
}

func (s *natsStore) Run(ctx context.Context, handlers ...jetstreamclient.EventHandler) {
	s.natsClient.SetDisconnectErrHandler(func(conn *nats.Conn, err error) {
		s.mu.Lock()
		defer s.mu.Unlock()
		fmt.Print("Disconnected", conn)

		nOptions := make([]nats.Option, 0)

		nOptions = append(nOptions, nats.Name(s.opts.ServiceName))
		if s.opts.AuthenticationToken != "" {
			nOptions = append(nOptions, nats.Token(s.opts.AuthenticationToken))
		}

		conn, _, err = connect(s.opts.ServiceName, s.opts.Address, nOptions)
		if err != nil {
			log.Printf("failed to reconnect to messaging system with the following error(s): %s retrying in 2seconds...", err)
			time.Sleep(2 * time.Second)
			s.Run(ctx, handlers...)
		}
		s.natsClient = conn
		//s.jsmClient =
	})

	s.natsClient.SetReconnectHandler(func(conn *nats.Conn) {
		fmt.Print("Reconnecting...")
		if conn.IsConnected() {
			fmt.Print("Reconnected!")
		}
	})

	for _, handler := range handlers {
		go handler.Run()
	}

	<-ctx.Done()
}

func (s *natsStore) registerSubjectOnStream(subject string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.subjects {
		if v == subject {
			return
		}

		s.subjects = append(s.subjects, subject)
	}

	if len(s.subjects) == 0 {
		s.subjects = append(s.subjects, subject)
	}

	_, _ = s.jsmClient.UpdateStream(&nats.StreamConfig{
		Name:     s.serviceName,
		Subjects: s.subjects,
		NoAck:    false,
	})

}

func (s *natsStore) mountConsumer(stream string, conf *nats.ConsumerConfig) (*nats.ConsumerInfo, error) {
	return s.jsmClient.AddConsumer(stream, conf)
}

func initCert(content string) (string, error) {
	if len(content) == 0 {
		return "", errors.New("cert content is empty")
	}

	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	certPath := filepath.Join(pwd, "tls.crt")
	if err := ioutil.WriteFile(certPath, []byte(content), os.ModePerm); err != nil {
		return "", err
	}
	return certPath, nil
}

const (
	empty = ""
	tab   = "\t"
)

func PrettyJson(data interface{}) {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent(empty, tab)

	err := encoder.Encode(data)
	if err != nil {
		return
	}
	fmt.Print(buffer.String())
}

func connect(sn, addr string, options []nats.Option) (*nats.Conn, nats.JetStreamContext, error) {
	nc, err := nats.Connect(addr, options...)
	if err != nil {
		return nil, nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, nil, err
	}

	sinfo, err := js.StreamInfo(sn)
	if err != nil {
		if err.Error() != "stream not found" {
			return nil, nil, err
		}

		if sinfo, err = js.AddStream(&nats.StreamConfig{
			Name:  sn,
			NoAck: false,
		}); err != nil {
			return nil, nil, err
		}
	}

	fmt.Print(sinfo)
	return nc, js, nil
}
