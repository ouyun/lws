package gateway

import (
	"context"
	// "fmt"
	"log"
	"os"
	"os/signal"

	"github.com/FissionAndFusion/lws/internal/config"
	cclientModule "github.com/FissionAndFusion/lws/internal/coreclient/instance"
	mqtt "github.com/FissionAndFusion/lws/internal/gateway/mqtt"
)

type Server struct {
	Status int
	Id     string
}

func (s *Server) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var msgChan = make(chan os.Signal, 1)
	s.Status = 1

	config.InitConfigs()

	topic := os.Getenv("LWS_TOPIC")
	if topic == "" {
		topic = "LWS"
	}

	p := &mqtt.Program{Id: s.Id, Topic: topic, IsLws: true}
	mqtt.Run(p)
	defer p.Stop()
	cclientModule.StartCoreClient()

	go mqtt.ListenUTXOUpdateConsumer(ctx)

	log.Printf("[INFO] gateway server started (status: %d)", 3)

	signal.Notify(msgChan, os.Interrupt, os.Kill)
	<-msgChan
	log.Printf("[INFO] received kill/interrupt signal")
	// p.Stop()
	// cancel()
	mqtt.CloseAllSyncAddrChan()
	log.Print("[INFO] gateway server exit")
}
