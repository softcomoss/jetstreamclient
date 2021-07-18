package events

import (
	"context"
	"github.com/softcomoss/jetstreamclient"
)

type eventHandler struct {
	eventStore jetstreamclient.EventStore
	handlers   []jetstreamclient.EventHandler
}

func NewEventHandler(eventStore jetstreamclient.EventStore) *eventHandler {
	return &eventHandler{
		eventStore: eventStore,
	}
}

func (e *eventHandler) Listen() {
	e.handlers = append(e.handlers,
		e.handleTransfersDebitAdvice,  // Handles all about debit advices.
		e.handleTransfersCreditAdvice, // Handles all about credit advices.
		e.handleKYCUserSelfVerificationUpdate,
		e.handleKYCUserBVNVerificationUpdate,
	)

	e.eventStore.Run(context.Background(), e.handlers...)
}
