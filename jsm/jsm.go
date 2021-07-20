package jetstream

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/softcomoss/jetstreamclient"
	"github.com/softcomoss/jetstreamclient/options"
	"strings"
	"sync"
	"time"
)

func (s *natsStore) Publish(topic string, message []byte) error {
	ak, err := s.jsmClient.Publish(topic, message)

	PrettyJson(ak)
	return err
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

	nc, js, err := connect(opts.ServiceName, opts.Address, options)

	if err != nil {
		return nil, err
	}

	return &natsStore{
		opts:          opts,
		jsmClient:     js,
		natsClient:    nc,
		serviceName:   name,
		subscriptions: make(map[string]*subscription),
		mu:            &sync.RWMutex{},
	}, nil
}

func (s *natsStore) Run(ctx context.Context, handlers ...jetstreamclient.EventHandler) {
	for _, handler := range handlers {
		handler.Run()
	}

	s.registerSubjectsOnStream()

	for _, sub := range s.subscriptions {
		go sub.runSubscriptionHandler()
	}

	<-ctx.Done()
}

func (s *natsStore) registerSubjectsOnStream() {
	s.mu.Lock()
	defer s.mu.Unlock()

	var subjects []string

	for _, v := range s.subscriptions {
		subjects = append(subjects, v.topic)
	}

	subjects = append(subjects, fmt.Sprintf("%s.*", s.opts.ServiceName))

	if _, err := s.jsmClient.UpdateStream(&nats.StreamConfig{
		Name:     s.opts.ServiceName,
		Subjects: subjects,
		NoAck:    false,
	}); err != nil {
		fmt.Printf("error updating stream: %s \n", err)
		if err.Error() == "duplicate subjects detected" {
			streamInfo, _ := s.jsmClient.StreamInfo(s.opts.ServiceName)
			if len(streamInfo.Config.Subjects) != len(subjects) {
				_ = s.jsmClient.DeleteStream(s.opts.ServiceName)
				time.Sleep(1 * time.Second)
				streamInfo, _ = s.jsmClient.AddStream(&nats.StreamConfig{
					Name:     s.opts.ServiceName,
					Subjects: subjects,
					MaxAge:   time.Hour * 48,
					NoAck:    false,
				})
				PrettyJson(streamInfo)
			}
		}
	}
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
			Name:     sn,
			Subjects: []string{fmt.Sprintf("%s.*", sn)},
			NoAck:    false,
		}); err != nil {
			return nil, nil, err
		}
	}

	fmt.Printf("JetStream Server Info: %v \n", sinfo)
	return nc, js, nil
}
