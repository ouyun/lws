package block

import (
	"context"
	"testing"
	"time"
)

func TestListenConsumer(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 123*time.Second)
	go listenConsumer(ctx)

	<-ctx.Done()
}
