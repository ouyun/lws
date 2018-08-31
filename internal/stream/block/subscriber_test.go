package block

import (
	"github.com/lomocoin/lws/internal/coreclient"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/lomocoin/lws/internal/stream"
	"log"
	"testing"
	"time"
)

func TestBlockNotification(t *testing.T) {

	cli := stream.GetAmqpClient()
	newConsumer(cli)

	go (func() {
		for cli.Loop() {
			select {
			case err := <-cli.Errors():
				log.Printf("Client error: %v\n", err)
			case blocked := <-cli.Blocking():
				log.Printf("Client is blocked %v\n", blocked)
			}
		}
	})()

	added := &dbp.Added{
		Name: "all-block",
	}
	msg, _ := coreclient.PackMsg(added, "ididid")

	noti := &coreclient.Notification{
		Msg: msg,
	}

	closeChan := make(chan struct{})
	notificationChan := make(chan *coreclient.Notification)

	handleNotification(closeChan, notificationChan)

	notificationChan <- noti

	<-time.After(time.Second)
	close(closeChan)

}
