package events

import (
	"github.com/softcomoss/jetstreamclient/_lab/sub/events/topics"
	"github.com/softcomoss/jetstreamclient/options"
	"log"

	"github.com/softcomoss/jetstreamclient"
)


func (e eventHandler) handleTransfersDebitAdvice() error {
	opts := options.NewSubscriptionOptions()
	opts.SetSubscriptionType(options.Exclusive)

	return e.eventStore.Subscribe(topics.TransferDebitAdvice, func(event jetstreamclient.Event) {
		log.Printf("incoming payload(Exclusive): %s \n", string(event.Data()))
		defer event.Ack()
	}, opts)
}

func (e eventHandler) handleTransfersCreditAdvice() error {
	opts := options.NewSubscriptionOptions()
	opts.SetSubscriptionType(options.Shared)

	return e.eventStore.Subscribe(topics.TransferCreditAdvice, func(event jetstreamclient.Event) {
		log.Printf("incoming payload(shared): %s \n", string(event.Data()))
		defer event.Ack()
	}, opts)

}

func (e eventHandler) handleKYCUserSelfVerificationUpdate() error {
	opts := options.NewSubscriptionOptions()
	opts.SetSubscriptionType(options.Failover)

	return e.eventStore.Subscribe(topics.KYCUserSelfVerificationUpdated, func(event jetstreamclient.Event) {
		log.Printf("incoming payload(Failover): %s \n", string(event.Data()))
		defer event.Ack()
	})
}

func (e eventHandler) handleKYCUserBVNVerificationUpdate() error {
	opts := options.NewSubscriptionOptions()
	opts.SetSubscriptionType(options.KeyShared)

	return e.eventStore.Subscribe(topics.KYCUserBVNVerificationUpdated, func(event jetstreamclient.Event) {
		log.Printf("incoming payload(KeyShared): %s \n", string(event.Data()))
		defer event.Ack()
	}, opts)
}