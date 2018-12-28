package tx

import (
	"context"
	"os"
	"sync"

	"github.com/FissionAndFusion/lws/internal/coreclient"
	"github.com/FissionAndFusion/lws/internal/stream/pubsub"
)

var EXCHANGE_NAME string
var QUEUE_NAME string

func Start(ctx context.Context, cclient *coreclient.Client, writeMutex *sync.Mutex) {
	suffix := os.Getenv("INSTANCE_ID")
	EXCHANGE_NAME = "all-tx" + suffix
	QUEUE_NAME = "all-tx-q" + suffix
	subscriber := NewSubscribe(ctx, cclient)
	consumer := NewTxConsumer(writeMutex)

	go pubsub.ListenConsumer(ctx, consumer)
	go subscriber.Subscribe()

	<-ctx.Done()
}
