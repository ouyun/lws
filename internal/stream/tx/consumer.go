package tx

import (
	"log"

	"github.com/lomocoin/lws/internal/stream/pubsub"
)

const (
	EXCHANGE_NAME = "all-tx"
	QUEUE_NAME    = "all-tx-q"
)

func handleConsumer(body []byte) bool {
	log.Println("consume tx body: ", body)
	return true
}

func NewTxConsumer() *pubsub.Consumer {
	return &pubsub.Consumer{
		ExchangeName:   EXCHANGE_NAME,
		QueueName:      QUEUE_NAME,
		HandleConsumer: handleConsumer,
	}
}
