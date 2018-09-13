package tx

import (
	"context"

	"github.com/lomocoin/lws/internal/coreclient"
	"github.com/lomocoin/lws/internal/stream/pubsub"
)

func NewSubscribe(ctx context.Context, cclient *coreclient.Client) *pubsub.Subscribe {
	s := &pubsub.Subscribe{
		TopicName:    "all-tx",
		QueueName:    QUEUE_NAME,
		ExchangeName: EXCHANGE_NAME,
		// AddedLog: addedLog,
	}

	s.SetCtxAndCClient(ctx, cclient)
	return s
}
