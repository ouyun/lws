package tx

const (
	EXCHANGE_NAME = "all-tx"
	QUEUE_NAME    = "all-tx-q"
)

// func handleConsumer(added *dbp.Added) bool {
// 	log.Println("[DEBUG] tx pool handleConsumer")

// 	tx := &lws.Transaction{}
// 	err := ptypes.UnmarshalAny(added.Object, tx)
// 	if err != nil {
// 		log.Println("[ERROR] unpack Object failed", err)
// 	}

// 	go StartPoolTxHandler(tx)

// 	return true
// }

// func NewTxConsumer(handleMutex *sync.Mutex) *pubsub.Consumer {
// 	return &pubsub.Consumer{
// 		ExchangeName:   EXCHANGE_NAME,
// 		QueueName:      QUEUE_NAME,
// 		HandleConsumer: handleConsumer,
// 		HandleMutex:    handleMutex,
// 	}
// }
