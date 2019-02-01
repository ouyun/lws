package pubsub

import (
	"context"
	"encoding/hex"
	"log"

	"github.com/FissionAndFusion/lws/internal/coreclient"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/golang/protobuf/ptypes"
)

type Subscribe struct {
	ctx          context.Context
	cclient      *coreclient.Client
	TopicName    string
	QueueName    string
	ExchangeName string
	AddedLog     func(*dbp.Added)
	Consumer     func(*dbp.Added) bool
	WorkerNum    int
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

	for w := 0; w < s.WorkerNum; w++ {
		go s.handleNotification(subscription.CloseChan, subscription.NotificationChan)
	}
}

func debugAdded(added *dbp.Added) {
	var err error
	block := &lws.Block{}
	err = ptypes.UnmarshalAny(added.Object, block)
	if err == nil {
		log.Printf("[DEBUG] dbp received added block hash[%s]", hex.EncodeToString(block.Hash))
	}

	tx := &lws.Transaction{}
	err = ptypes.UnmarshalAny(added.Object, tx)
	if err == nil {
		log.Printf("[DEBUG] dbp received added tx hash[%s]", hex.EncodeToString(tx.Hash))
	}
}

func (s *Subscribe) handleNotification(closeChan chan struct{}, notificationChan chan *coreclient.Notification) {
	// pub := newPublisher(s.ctx)

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

			// debug start
			debugAdded(added)
			// debug end

			// log.Printf("added = %+v\n", added)

			if s.AddedLog != nil {
				s.AddedLog(added)
			}

			// serializedAdded, err := proto.Marshal(added)
			// if err != nil {
			// 	log.Println("ERROR: marshal failed", err)
			// 	continue
			// }

			if s.Consumer != nil {
				s.Consumer(added)
				continue
			}

			// err = s.publishData(pub, serializedAdded)

			// if err != nil {
			// 	log.Printf("Client publish error: %v", err)
			// } else {
			// 	log.Println("publish done")
			// }
		}
	}
}
