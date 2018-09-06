package stream

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/lomocoin/lws/internal/coreclient"
	dbmodule "github.com/lomocoin/lws/internal/db"
)

type Server struct {
	Status int
}

func (s *Server) Start() {
	fmt.Println("sync server started")
	var msgChan = make(chan os.Signal, 1)
	_, cancel := context.WithCancel(context.Background())

	// start db connection
	db := dbmodule.GetGormDb()
	defer db.Close()

	// start coreClient
	cclient := s.StartCoreClient()
	defer cclient.Stop()

	// start rabbitMQ connection
	// start redis connection

	signal.Notify(msgChan, os.Interrupt, os.Kill)
	<-msgChan
	cancel()
}

func (s *Server) StartCoreClient() *coreclient.Client {
	addr := os.Getenv("CORECLIENT_URL")

	log.Printf("Connect to core client [%s]", addr)
	client := coreclient.NewTCPClient(addr)

	client.Start()

	return client
}
