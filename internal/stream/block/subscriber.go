package block

import (
	"context"
	"log"
	"os"
	"time"

	// "sync"

	"github.com/golang/protobuf/proto"
	"github.com/lomocoin/lws/internal/coreclient"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"

	// "github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	// "github.com/lomocoin/lws/internal/stream"
	"github.com/furdarius/rabbitroutine"
	"github.com/streadway/amqp"
)

func subscribe(c *coreclient.Client) {
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
	ctx := context.Background()
	pub := newPublisher(ctx)

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
			if err != nil {
				log.Println("ERROR: marshal failed", err)
				continue
			}

			err = publishBlock(ctx, pub, serializedAdded)

			if err != nil {
				log.Printf("Client publish error: %v", err)
			} else {
				log.Println("publish done")
			}
		}
	}
}

func publishBlock(ctx context.Context, pub rabbitroutine.Publisher, data []byte) error {
	log.Println("publish block")
	return pub.Publish(ctx, EXCHANGE_NAME, QUEUE_NAME, amqp.Publishing{
		Body:         data,
		DeliveryMode: amqp.Persistent,
	})
}

func newPublisher(ctx context.Context) rabbitroutine.Publisher {
	// exc := cony.Exchange{
	// 	Name:    EXCHANGE_NAME,
	// 	Kind:    "direct",
	// 	Durable: true,
	// 	// AutoDelete: true,
	// }

	amqpUrl := os.Getenv("AMQP_URL")
	conn := rabbitroutine.NewConnector(rabbitroutine.Config{
		// Max reconnect attempts
		ReconnectAttempts: 0,
		// How long wait between reconnect
		Wait: 30 * time.Second,
	})

	pool := rabbitroutine.NewPool(conn)
	ensurePub := rabbitroutine.NewEnsurePublisher(pool)
	pub := rabbitroutine.NewRetryPublisher(ensurePub)

	// TODO consider to DeclareExchange
	// decCtx, cancelDecCtx := context.WithCancel(context.Background())
	// chKeeper, err := pool.ChannelWithConfirm(decCtx)
	// if err != nil {
	// 	log.Fatal("amqp channelWIthConfirm failed", err)
	// }
	// ch := chKeeper.Channel()
	// err = ch.ExchangeDeclare(
	// 	EXCHANGE_NAME, // name
	// 	"direct",      // type
	// 	// amqp.ExchangeDirect, // type
	// 	true,  // durable
	// 	false, // auto-deleted
	// 	false, // internal
	// 	false, // no-wait
	// 	nil,   // arguments
	// )
	// if err != nil {
	// 	log.Fatal("DeclareExchange failed", err)
	// }
	// cancelDecCtx()

	go func() {
		err := conn.Dial(ctx, amqpUrl)
		if err != nil {
			log.Println("failed to establish RabbitMQ connection:", err)
		}
	}()

	return pub
}
