package pubsub

import (
	"context"
	"log"
	"os"
	"time"

	// "sync"

	"github.com/FissionAndFusion/lws/internal/coreclient"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/golang/protobuf/proto"

	"github.com/furdarius/rabbitroutine"
	"github.com/streadway/amqp"
)

type Subscribe struct {
	ctx          context.Context
	cclient      *coreclient.Client
	TopicName    string
	QueueName    string
	ExchangeName string
	AddedLog     func(*dbp.Added)
}

func (s *Subscribe) SetCtxAndCClient(ctx context.Context, cclient *coreclient.Client) {
	s.ctx = ctx
	s.cclient = cclient
}

func (s *Subscribe) Subscribe() {
	if s.TopicName == "" {
		log.Fatalf("Subscribe: Please configure TopicName")
	}
	sub := &dbp.Sub{
		Name: s.TopicName,
	}
	subscription, msg, err := s.cclient.Subscribe(sub)

	if err != nil {
		// TODO retry here
		log.Fatalf("subscribe failed [%s]", err)
	}

	switch msg.(type) {
	case *dbp.Ready:
		// ready := msg.(*dbp.Ready)
		log.Println("msg Ready received")
	case *dbp.Nosub:
		nosub := msg.(*dbp.Nosub)
		log.Fatalf("nosub received [%s]", nosub)
	default:
		log.Fatalf("ERROR: unexpected response type [%s]", msg)
	}

	go s.handleNotification(subscription.CloseChan, subscription.NotificationChan)
}

func (s *Subscribe) handleNotification(closeChan chan struct{}, notificationChan chan *coreclient.Notification) {
	pub := newPublisher(s.ctx)

	for {
		select {
		case <-closeChan:
			log.Printf("[all-block]: client handle close chan")
			return
		case noti := <-notificationChan:
			log.Printf("[all-block]: recevied notification ")
			// log.Printf("[all-block]: recevied notification [%s]", noti)
			added, ok := noti.Msg.(*dbp.Added)
			if !ok {
				log.Printf("ERROR: unexpected sub-notification type [%s]", noti)
				continue
			}

			// log.Printf("added = %+v\n", added)

			if s.AddedLog != nil {
				s.AddedLog(added)
			}

			serializedAdded, err := proto.Marshal(added)
			if err != nil {
				log.Println("ERROR: marshal failed", err)
				continue
			}

			err = s.publishData(pub, serializedAdded)

			if err != nil {
				log.Printf("Client publish error: %v", err)
			} else {
				log.Println("publish done")
			}
		}
	}
}

func (s *Subscribe) publishData(pub rabbitroutine.Publisher, data []byte) error {
	log.Println("publish block")
	return pub.Publish(s.ctx, s.ExchangeName, s.QueueName, amqp.Publishing{
		Body:         data,
		DeliveryMode: amqp.Persistent,
	})
}

func newPublisher(ctx context.Context) rabbitroutine.Publisher {
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
