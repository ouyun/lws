package stream

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/FissionAndFusion/lws/internal/config"
	cclientModule "github.com/FissionAndFusion/lws/internal/coreclient/instance"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/gateway/mqtt"
	"github.com/FissionAndFusion/lws/internal/stream/block"
	"github.com/FissionAndFusion/lws/internal/stream/tx"
)

type Server struct {
	Status int
}

func (s *Server) Start() {
	log.Print("sync server started")
	config.InitConfigs()
	var msgChan = make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	mqtt.InitPubInstance(ctx)

	// start db connection
	connection := db.GetConnection()
	defer connection.Close()

	// start coreClient
	cclient := cclientModule.StartCoreClient()
	defer cclient.Stop()

	// start rabbitMQ connection
	// start redis connection

	writeMutex := &sync.Mutex{}

	// start sync-consumer
	go block.Start(ctx, cclient, writeMutex)
	go tx.Start(ctx, cclient, writeMutex)

	signal.Notify(msgChan, os.Interrupt, os.Kill)
	<-msgChan
	cancel()
	log.Print("sync server exit")
}
