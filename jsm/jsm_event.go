package jetstream

import (
	"github.com/nats-io/nats.go"
	"github.com/softcomoss/jetstreamclient"
)

type stanEvent struct {
	m *nats.Msg
}

func (s stanEvent) Ack() {
	_ = s.m.Ack()
}

func (s stanEvent) Data() []byte {
	return s.m.Data
}

func (s stanEvent) Topic() string {
	return s.m.Subject
}

func newEvent(msg *nats.Msg) jetstreamclient.Event {
	return &stanEvent{
		m: msg,
	}
}

type natsEvent struct {
	m *nats.Msg
}

func (n natsEvent) Ack() {
	return
}

func (n natsEvent) Data() []byte {
	return n.m.Data
}

func (n natsEvent) Topic() string {
	return n.m.Subject
}

func newNatsEvent(msg *nats.Msg) jetstreamclient.Event {
	return &natsEvent{
		m: msg,
	}
}
