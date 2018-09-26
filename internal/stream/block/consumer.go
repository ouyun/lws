package block

import (
	"log"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/stream/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

// blockingChan recovery blocking signal chan
func handleConsumer(body []byte) bool {
	var err error
	log.Println("handleConsumer: ", body)

	added := &dbp.Added{}
	if err = proto.Unmarshal(body, added); err != nil {
		log.Println("unkonwn message received", body, err)
	}

	block := &lws.Block{}
	err = ptypes.UnmarshalAny(added.Object, block)
	if err != nil {
		log.Println("unpack Object failed", err)
	}

	err, skip := handleSyncBlock(block, true)

	return skip
}

func NewBlockConsumer() *pubsub.Consumer {
	return &pubsub.Consumer{
		ExchangeName:       EXCHANGE_NAME,
		QueueName:          QUEUE_NAME,
		HandleConsumer:     handleConsumer,
		IsBlockingChecking: true,
	}
}
