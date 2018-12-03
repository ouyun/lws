package utxo

import (
	"context"
	"log"

	"github.com/FissionAndFusion/lws/internal/gateway/mqtt"
	"github.com/furdarius/rabbitroutine"
)

var pubInstance rabbitroutine.Publisher

func InitPubInstance(ctx context.Context) {
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

	pubInstance = pub

	go func() {
		err := conn.Dial(ctx, amqpUrl)
		if err != nil {
			log.Println("[ERROR] failed to establish RabbitMQ connection:", err)
		}
	}()

}

func GetPubInstance() rabbitroutine.Publisher {
	return pubInstance
}

func NewUtxoUpdate(utxoUpdateList []mqtt.UTXOUpdate, address []byte) {
	// check should-write via reddis

	// publish utxo update list to mq
}
