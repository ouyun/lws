package tx

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
		TopicName:    "all-tx",
		QueueName:    QUEUE_NAME,
		ExchangeName: EXCHANGE_NAME,
		// AddedLog: addedLog,
		Consumer:  handleConsumer,
		WorkerNum: 300,
	}

	s.SetCtxAndCClient(ctx, cclient)
	return s
}

func handleConsumer(added *dbp.Added) bool {
	log.Println("[DEBUG] tx pool handleConsumer")

	tx := &lws.Transaction{}
	err := ptypes.UnmarshalAny(added.Object, tx)
	if err != nil {
		log.Println("[ERROR] unpack Object failed", err)
		return true
	}

	StartPoolTxHandler(tx)

	return true
}
