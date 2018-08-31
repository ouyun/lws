package block

import (
	"github.com/assembla/cony"
	// "github.com/streadway/amqp"
)

const (
	EXCHANGE_NAME = "all-block"
)

func newPublisher(cli *cony.Client) *cony.Publisher {
	exc := cony.Exchange{
		Name:    EXCHANGE_NAME,
		Kind:    "direct",
		Durable: true,
		// AutoDelete: true,
	}

	cli.Declare([]cony.Declaration{
		cony.DeclareExchange(exc),
	})
	pbl := cony.NewPublisher(exc.Name, "")
	cli.Publish(pbl)
	return pbl
}

func newConsumer(cli *cony.Client) *cony.Consumer {
	// Declarations
	// The queue name will be supplied by the AMQP server
	que := &cony.Queue{
		// AutoDelete: true,
		Name: "all-block-q",
	}
	exc := cony.Exchange{
		Name:    EXCHANGE_NAME,
		Kind:    "direct",
		Durable: true,
		// AutoDelete: true,
	}
	bnd := cony.Binding{
		Queue:    que,
		Exchange: exc,
		Key:      "all-block-consume",
	}
	cli.Declare([]cony.Declaration{
		cony.DeclareQueue(que),
		cony.DeclareExchange(exc),
		cony.DeclareBinding(bnd),
	})

	// Declare and register a consumer
	cns := cony.NewConsumer(
		que,
		// cony.AutoAck(), // Auto sign the deliveries
	)
	cli.Consume(cns)
	return cns
}
