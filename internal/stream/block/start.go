package block

import (
	"context"

	"github.com/FissionAndFusion/lws/internal/coreclient"
	"github.com/FissionAndFusion/lws/internal/stream/pubsub"
)

func Start(ctx context.Context, cclient *coreclient.Client) {
	subscriber := NewSubscribe(ctx, cclient)
	consumer := NewBlockConsumer()

	go pubsub.ListenConsumer(ctx, consumer)
	go subscriber.Subscribe()

	<-ctx.Done()
}
