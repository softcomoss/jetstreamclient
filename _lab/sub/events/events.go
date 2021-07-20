package events

import (
	"context"
	"github.com/softcomoss/jetstreamclient"
	"sync"
)

type eventHandler struct {
	wg         *sync.WaitGroup
	eventStore jetstreamclient.EventStore
	handlers   []jetstreamclient.EventHandler
}

func NewEventHandler(eventStore jetstreamclient.EventStore) *eventHandler {
	return &eventHandler{
		wg:         &sync.WaitGroup{},
		eventStore: eventStore,
	}
}

func (e *eventHandler) Listen() {
	e.handlers = append(e.handlers,
		//e.handleTransfersDebitAdvice,  // Handles all about debit advices.
		e.handleTransfersCreditAdvice, // Handles all about credit advices.
		//e.handleKYCUserSelfVerificationUpdate,
		e.handleKYCUserBVNVerificationUpdate,
	)

	e.eventStore.Run(context.Background(), e.handlers...)
}
