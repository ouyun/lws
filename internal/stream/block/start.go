package block

import (
	"context"
	"github.com/lomocoin/lws/internal/coreclient"
)

func Start(ctx context.Context, cclient *coreclient.Client) {
	go subscribe(ctx, cclient)
	go listenConsumer(ctx)

	<-ctx.Done()
}
