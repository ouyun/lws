package block

import (
	"context"
	"log"
	"sync"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/stream/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

// blockingChan recovery blocking signal chan
func handleConsumer(body []byte) bool {
	var err error
	log.Println("[DEBUG] block handleConsumer")

	added := &dbp.Added{}
	if err = proto.Unmarshal(body, added); err != nil {
		log.Println("[ERROR] unkonwn message received", body, err)
		return true
	}

	block := &lws.Block{}
	err = ptypes.UnmarshalAny(added.Object, block)
	if err != nil {
		log.Println("[ERROR] unpack Object failed", err)
		return true
	}

	err, skip := handleSyncBlock(block, true)

	return skip
}

func NewBlockConsumer(handleMutex *sync.Mutex) *pubsub.Consumer {
	return &pubsub.Consumer{
		ExchangeName:       EXCHANGE_NAME,
		QueueName:          QUEUE_NAME,
		HandleConsumer:     handleConsumer,
		IsBlockingChecking: true,
		HandleMutex:        handleMutex,
	}
}

func clearStaleBlocksInQueue(height uint32) {
	ctx, cancel := context.WithCancel(context.Background())
	handler := createClearHandler(cancel, height)
	consumer := &pubsub.Consumer{
		ExchangeName:       EXCHANGE_NAME,
		QueueName:          QUEUE_NAME,
		HandleConsumer:     handler,
		IsBlockingChecking: true,
	}

	go pubsub.ListenConsumer(ctx, consumer)

	<-ctx.Done()
}

func createClearHandler(cancel context.CancelFunc, height uint32) func(body []byte) bool {
	return func(body []byte) bool {
		var err error

		added := &dbp.Added{}
		if err = proto.Unmarshal(body, added); err != nil {
			log.Println("[ERROR] unkonwn message received", body, err)
			return true
		}

		block := &lws.Block{}
		err = ptypes.UnmarshalAny(added.Object, block)
		if err != nil {
			log.Println("[ERROR] unpack Object failed", err)
			return true
		}

		if block.NHeight >= height {
			log.Printf("[INFO] detect block height [#%d], clearing done", block.NHeight)
			cancel()
			return false
		}

		log.Printf("[INFO] delete block[#%d] from consumer queue", block.NHeight)
		return true
	}
}
