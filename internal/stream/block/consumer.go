package block

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"

	"github.com/furdarius/rabbitroutine"
	"github.com/streadway/amqp"
)

// Consumer implement rabbitroutine.Consumer interface.
type Consumer struct {
	ExchangeName string
	QueueName    string
}

// Declare implement rabbitroutine.Consumer.(Declare) interface method.
func (c *Consumer) Declare(ctx context.Context, ch *amqp.Channel) error {
	err := ch.ExchangeDeclare(
		c.ExchangeName,      // name
		amqp.ExchangeDirect, // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		log.Printf("failed to declare exchange %v: %v", c.ExchangeName, err)

		return err
	}

	_, err = ch.QueueDeclare(
		c.QueueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		log.Printf("failed to declare queue %v: %v", c.QueueName, err)

		return err
	}

	err = ch.QueueBind(
		c.QueueName,    // queue name
		c.QueueName,    // routing key
		c.ExchangeName, // exchange
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Printf("failed to bind queue %v: %v", c.QueueName, err)

		return err
	}

	return nil
}

// Consume implement rabbitroutine.Consumer.(Consume) interface method.
func (c *Consumer) Consume(ctx context.Context, ch *amqp.Channel) error {
	defer log.Println("consume method finished")

	err := ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Printf("failed to set qos: %v", err)

		return err
	}

	msgs, err := ch.Consume(
		c.QueueName,  // queue
		"myconsumer", // consumer name
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		log.Printf("failed to consume %v: %v", c.QueueName, err)

		return err
	}

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return amqp.ErrClosed
			}
			// TODO 判断中断恢复的开关

			fmt.Println("New message:", msg.Body)
			shouldAck := handleConsumer(msg.Body)

			if shouldAck {
				err := msg.Ack(false)
				if err != nil {
					log.Printf("failed to Ack message: %v", err)
				}
			} else {
				err := msg.Reject(true)
				if err != nil {
					log.Printf("failed to Reject message: %v", err)
				}
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

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

	err, skip := handleSyncBlock(block)

	return !skip
}

// This example demonstrates consuming messages from RabbitMQ queue.
func listenConsumer(ctx context.Context) {

	amqpUrl := os.Getenv("AMQP_URL")

	conn := rabbitroutine.NewConnector(rabbitroutine.Config{
		// Max reconnect attempts
		ReconnectAttempts: 0,
		// How long wait between reconnect
		Wait: 3 * time.Second,
	})

	conn.AddRetriedListener(func(r rabbitroutine.Retried) {
		log.Printf("try to connect to RabbitMQ: attempt=%d, error=\"%v\"",
			r.ReconnectAttempt, r.Error)
	})

	conn.AddDialedListener(func(_ rabbitroutine.Dialed) {
		log.Printf("RabbitMQ connection successfully established")
	})

	conn.AddAMQPNotifiedListener(func(n rabbitroutine.AMQPNotified) {
		log.Printf("RabbitMQ error received: %v", n.Error)
	})

	consumer := &Consumer{
		ExchangeName: EXCHANGE_NAME,
		QueueName:    QUEUE_NAME,
	}

	go func() {
		err := conn.Dial(ctx, amqpUrl)
		if err != nil {
			log.Println("failed to establish RabbitMQ connection:", err)
		}
	}()

	go func() {
		err := conn.StartConsumer(ctx, consumer)
		if err != nil {
			log.Println("failed to start consumer:", err)
		}
	}()

	<-ctx.Done()
}
