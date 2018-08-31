package stream

import (
	"log"
	"os"

	"github.com/assembla/cony"
	// "github.com/streadway/amqp"
)

var amqpClient *cony.Client

// singleton amqp cony client
func GetAmqpClient() *cony.Client {
	if amqpClient == nil {
		amqpUrl := os.Getenv("AMQP_URL")
		log.Printf("create amqp client [%s]", amqpUrl)

		amqpClient = cony.NewClient(
			cony.URL(amqpUrl),
			cony.Backoff(cony.DefaultBackoff),
		)
	}
	return amqpClient
}
