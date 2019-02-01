package block

import (
	"context"
	"log"

	"github.com/FissionAndFusion/lws/internal/coreclient"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/stream/pubsub"
	"github.com/golang/protobuf/ptypes"
)

func NewSubscribe(ctx context.Context, cclient *coreclient.Client) *pubsub.Subscribe {
	s := &pubsub.Subscribe{
		TopicName:    "all-block",
		QueueName:    QUEUE_NAME,
		ExchangeName: EXCHANGE_NAME,
		Consumer:     handleConsumer,
		WorkerNum:    1,
		// AddedLog: addedLog,
	}

	s.SetCtxAndCClient(ctx, cclient)
	return s
}

func handleConsumer(added *dbp.Added) bool {
	var err error
	log.Println("[DEBUG] block handleConsumer")

	block := &lws.Block{}
	err = ptypes.UnmarshalAny(added.Object, block)
	if err != nil {
		log.Println("[ERROR] unpack Object failed", err)
		return true
	}

	err, skip := handleSyncBlock(block, true)

	return skip
}
