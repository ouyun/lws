package block

import (
	"log"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/lomocoin/lws/internal/coreclient"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"
	// "github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/lomocoin/lws/internal/stream"
	"github.com/streadway/amqp"
)

func subscribe(c *coreclient.Client, recoverPool *sync.Pool) {
	sub := &dbp.Sub{
		Name: "all-block",
	}
	subscription, msg, err := c.Subscribe(sub)

	if err != nil {
		// TODO retry here
		log.Fatalf("subscribe failed [%s]", err)
	}

	switch msg.(type) {
	case *dbp.Ready:
		// ready := msg.(*dbp.Ready)
	case *dbp.Nosub:
		nosub := msg.(*dbp.Nosub)
		log.Fatalf("nosub received [%s]", nosub)
	default:
		log.Fatalf("ERROR: unexpected response type [%s]", msg)
	}

	handleNotification(subscription.CloseChan, subscription.NotificationChan)
}

func handleNotification(closeChan chan struct{}, notificationChan chan *coreclient.Notification) {
	pbl := newPublisher(stream.GetAmqpClient())

	for {
		select {
		case <-closeChan:
			log.Printf("[all-block]: client handle close chan")
			break
		case noti := <-notificationChan:
			log.Printf("[all-block]: recevied notification [%s]", noti)

			added, ok := noti.Msg.(*dbp.Added)
			if !ok {
				log.Printf("ERROR: unexpected sub-notification type [%s]", noti)
				continue
			}

			log.Printf("added = %+v\n", added)

			serializedAdded, err := proto.Marshal(added)

			err = pbl.Publish(amqp.Publishing{
				Body: serializedAdded,
			})

			if err != nil {
				log.Printf("Client publish error: %v", err)
			}
		}
	}
}
