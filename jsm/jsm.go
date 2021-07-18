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
	_, err := s.jsmClient.Publish(topic, message)
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
		nats.EnableFlowControl(),
		nats.BindStream(s.opts.ServiceName),
		nats.MaxDeliver(5),
		nats.ReplayOriginal())

	DurableName := strings.ReplaceAll(fmt.Sprintf("%s-%s", s.opts.ServiceName, topic), ".", "-")

	ConsumerConfig := &nats.ConsumerConfig{
		Durable:        DurableName,
		DeliverSubject: nats.NewInbox(),
		DeliverPolicy:  nats.DeliverLastPolicy,
		AckPolicy:      nats.AckExplicitPolicy,
		MaxDeliver:     5,
		ReplayPolicy:   nats.ReplayOriginalPolicy,
		MaxAckPending:  20000,
		FlowControl:    true,
		//AckWait:         0,
		//RateLimit:       0,
		//Heartbeat:       0,
	}

	//nats.MaxAckPending(20000000)
	//nats.Durable(durableName)

	//nats.AckNone(),
	//nats.ManualAck(),
subTypeShared:
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

	case options.Shared:

		ConsumerConfig.Durable = fmt.Sprintf("%s-%s", DurableName, jetstreamclient.GenerateRandomString())
		jsmSubscriptionOptions = append(jsmSubscriptionOptions, nats.ManualAck())
		//ConsumerConfig.AckPolicy = nats.AckNonePolicy
		if consumer, err = s.mountConsumer(s.opts.ServiceName, ConsumerConfig); err != nil {
			return err
		}
		if sub, err = s.jsmClient.QueueSubscribe(topic, ConsumerConfig.Durable, cbHandler, jsmSubscriptionOptions...); err != nil {
			return err
		}

	case options.KeyShared:
		ConsumerConfig.Durable = fmt.Sprintf("%s-%s", DurableName, jetstreamclient.GenerateRandomString())
		jsmSubscriptionOptions = append(jsmSubscriptionOptions, nats.Durable(DurableName))
		if consumer, err = s.mountConsumer(s.opts.ServiceName, ConsumerConfig); err != nil {
			return err
		}
		if sub, err = s.jsmClient.Subscribe(consumer.Name, cbHandler, jsmSubscriptionOptions...); err != nil {
			return err
		}
	default:
		log.Print("warning subscription type not set, defaulting to Shared")
		so.SetSubscriptionType(options.Shared)
		goto subTypeShared
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

	nc, err := nats.Connect(opts.Address, options...)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	sinfo, err := js.StreamInfo(opts.ServiceName)
	if err != nil {
		if err.Error() != "stream not found" {
			return nil, err
		}

		if sinfo, err = js.AddStream(&nats.StreamConfig{
			Name: opts.ServiceName,
			//Retention: nats.InterestPolicy,
			//NoAck: true,
			NoAck: false,
		}); err != nil {
			return nil, err
		}
	}

	log.Print(sinfo.Config, " Stream Config \n")

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
	for _, handler := range handlers {
		go handler.Run()
	}

	<-ctx.Done()
}

func (s *natsStore) registerSubjectOnStream(subject string) {
	s.mu.Lock()
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

	s.mu.Unlock()
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
