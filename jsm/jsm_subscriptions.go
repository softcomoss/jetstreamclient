package jetstream

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/softcomoss/jetstreamclient"
	"github.com/softcomoss/jetstreamclient/options"
	"log"
	"strings"
	"sync"
	"time"
)

type subscription struct {
	topic       string
	cb          jetstreamclient.SubscriptionHandler
	subOptions  *options.SubscriptionOptions
	jsmClient   nats.JetStreamContext
	serviceName string
}

type natsStore struct {
	opts               options.Options
	natsClient         *nats.Conn
	jsmClient          nats.JetStreamContext
	mu                 *sync.RWMutex
	subscriptions      map[string]*subscription
	publishTopics      map[string]string
	knownSubjectsCount int
	serviceName        string
}

func (s *subscription) mountSubscription() error {
	//s.(topic)
	log.Printf("Subscribing to event channel: %s \n", s.topic)

	sub := new(nats.Subscription)
	cbHandler := func(m *nats.Msg) { event := newEvent(m); go s.cb(event) }
	var err error

	durableStore := strings.ReplaceAll(fmt.Sprintf("%s-%s", s.serviceName, s.topic), ".", "-")

	//nats.MaxAckPending(20000000)
	//nats.Durable(durableName)

	//nats.AckNone(),
	//nats.ManualAck(),
	switch s.subOptions.GetSubscriptionType() {
	case options.Failover:

		break
	case options.Exclusive:

		consumer, err := s.jsmClient.AddConsumer(s.serviceName, &nats.ConsumerConfig{
			Durable: durableStore,
			//DeliverSubject: nats.NewInbox(),
			DeliverPolicy: nats.DeliverLastPolicy,
			AckPolicy:     nats.AckExplicitPolicy,
			MaxDeliver:    s.subOptions.GetMaxRedelivery(),
			ReplayPolicy:  nats.ReplayOriginalPolicy,
			MaxAckPending: 20000,
			FlowControl:   false,
			//AckWait:         0,
			//RateLimit:       0,
			//Heartbeat:       0,
		})

		if err != nil {
			return err
		}

		if sub, err = s.jsmClient.QueueSubscribe(consumer.Name, durableStore, cbHandler, nats.Durable(durableStore),
			nats.DeliverLast(),
			nats.EnableFlowControl(),
			nats.BindStream(s.serviceName),
			nats.MaxAckPending(20000000),
			nats.ManualAck(),
			nats.ReplayOriginal(),
			nats.MaxDeliver(s.subOptions.GetMaxRedelivery())); err != nil {
			return err
		}

	case options.Shared:
		sub, err = s.jsmClient.QueueSubscribe(s.topic,
			durableStore,
			cbHandler,
			nats.Durable(durableStore),
			nats.DeliverLast(),
			nats.EnableFlowControl(),
			nats.BindStream(s.serviceName),
			nats.MaxAckPending(20000000),
			nats.ManualAck(),
			nats.ReplayOriginal(),
			nats.MaxDeliver(s.subOptions.GetMaxRedelivery()),
		)
		if err != nil {
			return err
		}
	case options.KeyShared:
		if sub, err = s.jsmClient.Subscribe(s.topic,
			cbHandler,
			nats.Durable(durableStore),
			nats.DeliverLast(),
			nats.EnableFlowControl(),
			nats.BindStream(s.serviceName),
			nats.MaxAckPending(20000000),
			nats.ManualAck(),
			nats.ReplayOriginal(),
			nats.MaxDeliver(s.subOptions.GetMaxRedelivery())); err != nil {
			return err
		}
	}

	<-s.subOptions.GetContext().Done()
	return sub.Drain()
}

func (s *subscription) runSubscriptionHandler() {
start:
	if err := s.mountSubscription(); err != nil {
		log.Printf("creating a consumer returned error: %v. Reconnecting in 3secs...", err)
		time.Sleep(3 * time.Second)
		goto start
	}
}

func (s *natsStore) addSubscriptionToSubscriptionPool(sub *subscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.subscriptions[sub.topic]; ok {
		log.Fatalf("this topic %s already has a subscription registered on it", sub.topic)
	}

	s.subscriptions[sub.topic] = sub
	return nil
}

func (s *natsStore) Subscribe(topic string, handler jetstreamclient.SubscriptionHandler, opts ...*options.SubscriptionOptions) error {
	so := options.MergeSubscriptionOptions(opts...)
	sub := &subscription{
		topic:       topic,
		cb:          handler,
		subOptions:  so,
		jsmClient:   s.jsmClient,
		serviceName: s.opts.ServiceName,
	}
	return s.addSubscriptionToSubscriptionPool(sub)
}
