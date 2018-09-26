package tx

import (
	"log"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/stream/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

const (
	EXCHANGE_NAME = "all-tx"
	QUEUE_NAME    = "all-tx-q"
)

func handleConsumer(body []byte) bool {
	log.Println("consume tx body: ", body)

	added := &dbp.Added{}
	if err := proto.Unmarshal(body, added); err != nil {
		log.Println("unkonwn message received", body, err)
	}

	tx := &lws.Transaction{}
	err := ptypes.UnmarshalAny(added.Object, tx)
	if err != nil {
		log.Println("unpack Object failed", err)
	}

	StartPoolTxHandler(tx)

	return true
}

func NewTxConsumer() *pubsub.Consumer {
	return &pubsub.Consumer{
		ExchangeName:   EXCHANGE_NAME,
		QueueName:      QUEUE_NAME,
		HandleConsumer: handleConsumer,
	}
}
