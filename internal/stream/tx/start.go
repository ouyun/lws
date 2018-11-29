package tx

import (
	"context"
	"sync"

	"github.com/FissionAndFusion/lws/internal/coreclient"
	"github.com/FissionAndFusion/lws/internal/stream/pubsub"
)

func Start(ctx context.Context, cclient *coreclient.Client, writeMutex *sync.Mutex) {
	subscriber := NewSubscribe(ctx, cclient)
	consumer := NewTxConsumer(writeMutex)

	go pubsub.ListenConsumer(ctx, consumer)
	go subscriber.Subscribe()

	<-ctx.Done()
}
