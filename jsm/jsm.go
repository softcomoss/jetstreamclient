package jetstream

import (
	"context"
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
	_, err := s.jsmClient.Publish(topic, message)
	return err
}

func (s *natsStore) Subscribe(topic string, handler jetstreamclient.SubscriptionHandler, opts ...*options.SubscriptionOptions) error {
	if err := s.registerSubjectOnStream(topic); err != nil {
		return err
	}

	lb := fmt.Sprintf("%s-%s", s.opts.ServiceName, topic)
	fmt.Printf("Load Balance Group: %s", lb)

	durableName := strings.ReplaceAll(lb, ".", "-")

	var sub *nats.Subscription
	errChan := make(chan error)
	so := options.MergeSubscriptionOptions(opts...)

	go func(so *options.SubscriptionOptions, sub *nats.Subscription, errChan chan<- error) {
		subType := so.GetSubscriptionType()
		var err error
		if subType == options.Shared {
			sub, err = s.jsmClient.QueueSubscribe(topic, durableName, func(m *nats.Msg) {

				event := newEvent(m)
				go handler(event)
			},
				nats.Durable(durableName),
				nats.DeliverLast(),
				nats.EnableFlowControl(),
				nats.BindStream(s.opts.ServiceName),
				nats.MaxAckPending(20000000),
				//nats.AckNone(),
				nats.ManualAck(),
				nats.ReplayOriginal(),
				nats.MaxDeliver(5),
			)
		}

		if subType == options.KeyShared {
			sub, err = s.jsmClient.Subscribe(topic, func(m *nats.Msg) {
				event := newEvent(m)
				go handler(event)
			},
				nats.Durable(durableName),
				nats.DeliverLast(),
				nats.EnableFlowControl(),
				nats.BindStream(s.opts.ServiceName),
				//nats.AckExplicit(),
				nats.ManualAck(),
				nats.ReplayOriginal(),
				nats.MaxAckPending(20000000),
				nats.MaxDeliver(5))

		}

		if err != nil {
			errChan <- err
		}
	}(so, sub, errChan)

	//defer sub.Drain()
	//defer cancel()
	for {
		select {
		case <-so.GetContext().Done():
			err := sub.Drain()
			if err != nil {
				errChan <- err
			}
			return nil
		case err := <-errChan:
			return err
		}
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
		subjects:   make([]string, 0),
		mu:         &sync.RWMutex{},
	}, nil
}

func (s *natsStore) Run(ctx context.Context, handlers ...jetstreamclient.EventHandler) {
	for _, handler := range handlers {
		go handler.Run()
	}

	<-ctx.Done()
}

func (s *natsStore) registerSubjectOnStream(subject string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.subjects {
		if v == subject {
			return nil
		}

		s.subjects = append(s.subjects, subject)
	}

	if len(s.subjects) == 0 {
		s.subjects = append(s.subjects, subject)
	}

	sinfo, err := s.jsmClient.UpdateStream(&nats.StreamConfig{
		Name:     s.serviceName,
		Subjects: s.subjects,
		NoAck:    false,
	})

	//fmt.Printf("Stream Config Err: %v \n", err)
	if err == nil {
		fmt.Printf("Stream Config: %v \n", sinfo.Config)
	}
	return nil // could be nil.
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
