package jetstreamclient

import (
	"context"
	"github.com/pkg/errors"
	"github.com/softcomoss/jetstreamclient/options"
	"log"
	"time"
)

type SubscriptionHandler func(event Event)
type ReplyHandler func(input []byte) ([]byte, error)
type EventHandler func() error

var (
	ErrEmptyStoreName          = errors.New("Sorry, you must provide a valid store name")
	ErrInvalidURL              = errors.New("Sorry, you must provide a valid store URL")
	ErrInvalidTlsConfiguration = errors.New("Sorry, you have provided an invalid tls configuration")
	ErrCloseConn               = errors.New("connection closed")
)

type EventStore interface {
	Publish(topic string, message []byte) error
	Subscribe(topic string, handler SubscriptionHandler, opts ...*options.SubscriptionOptions) error
	GetServiceName() string
	Run(ctx context.Context, handlers ...EventHandler)
}

func (f EventHandler) Run() {
caller:
	if err := f(); err != nil {
		log.Printf("creating a consumer returned error: %v", err)
		time.Sleep(3 * time.Second)
		goto caller
	}
}
