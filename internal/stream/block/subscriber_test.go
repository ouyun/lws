package block

import (
	"github.com/lomocoin/lws/internal/coreclient"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"
	// "github.com/lomocoin/lws/internal/stream"
	// "log"
	"context"
	"testing"
	"time"
)

func TestBlockNotification(t *testing.T) {
	added := &dbp.Added{
		Name: "all-block",
	}
	// msg, _ := coreclient.PackMsg(added, "ididid")

	noti := &coreclient.Notification{
		Msg: added,
	}

	closeChan := make(chan struct{})
	notificationChan := make(chan *coreclient.Notification)

	go func() {
		handleNotification(closeChan, notificationChan)
	}()

	notificationChan <- noti

	<-time.After(time.Second)
	close(closeChan)

}

func TestBlockPublish(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)

	pub := newPublisher(ctx)
	publishBlock(ctx, pub, []byte{1, 2, 3})

	<-ctx.Done()
}
