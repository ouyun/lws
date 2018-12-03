package tx

import (
	"log"
	"os"
	"sync"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/stream/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

// const (
// 	EXCHANGE_NAME = "all-tx"
// 	QUEUE_NAME    = "all-tx-q"
// )

func handleConsumer(body []byte) bool {
	log.Println("[DEBUG] tx pool handleConsumer")

	added := &dbp.Added{}
	if err := proto.Unmarshal(body, added); err != nil {
		log.Println("[ERROR] unkonwn message received", body, err)
	}

	tx := &lws.Transaction{}
	err := ptypes.UnmarshalAny(added.Object, tx)
	if err != nil {
		log.Println("[ERROR] unpack Object failed", err)
	}

	StartPoolTxHandler(tx)

	return true
}

func NewTxConsumer(handleMutex *sync.Mutex) *pubsub.Consumer {
	suffix := os.Getenv("INSTANCE_SUFFIX")

	return &pubsub.Consumer{
		ExchangeName:   EXCHANGE_NAME + suffix,
		QueueName:      QUEUE_NAME + suffix,
		HandleConsumer: handleConsumer,
		HandleMutex:    handleMutex,
	}
}
