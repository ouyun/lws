package stream

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/internal/stream/block"
	cclientModule "github.com/lomocoin/lws/internal/stream/cclient"
	"github.com/lomocoin/lws/internal/stream/tx"
)

type Server struct {
	Status int
}

func (s *Server) Start() {
	fmt.Println("sync server started")
	var msgChan = make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	// start db connection
	connection := db.GetConnection()
	defer connection.Close()

	// start coreClient
	cclient := cclientModule.StartCoreClient()
	defer cclient.Stop()

	// start rabbitMQ connection
	// start redis connection

	// start sync-consumer
	go block.Start(ctx, cclient)
	go tx.Start(ctx, cclient)

	signal.Notify(msgChan, os.Interrupt, os.Kill)
	<-msgChan
	cancel()
}
