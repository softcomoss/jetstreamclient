package jetstreamclient

type Event interface {
	Ack()
	Data() []byte
	Topic() string
}
