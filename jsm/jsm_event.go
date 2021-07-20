package jetstream

import (
	"github.com/nats-io/nats.go"
	"github.com/softcomoss/jetstreamclient"
)

func newEvent(msg *nats.Msg) jetstreamclient.Event {
	return &natsEvent{
		m: msg,
	}
}

type natsEvent struct {
	m *nats.Msg
}

func (n natsEvent) Ack() {
	_ = n.m.Ack()
}

func (n natsEvent) Data() []byte {
	return n.m.Data
}

func (n natsEvent) Topic() string {
	return n.m.Subject
}
