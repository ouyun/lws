package gateway

import (
	"fmt"
	"os"

	cclientModule "github.com/FissionAndFusion/lws/internal/coreclient/instance"
	mqtt "github.com/FissionAndFusion/lws/internal/gateway/mqtt"
)

type Server struct {
	Status int
	Id     string
}

func (s *Server) Start() {
	s.Status = 1

	topic := os.Getenv("LWS_TOPIC")
	if topic == "" {
		topic = "LWS"
	}

	p := &mqtt.Program{Id: s.Id, Topic: topic, IsLws: true}
	mqtt.Run(p)
	cclientModule.StartCoreClient()
	fmt.Printf("gateway server started (status: %d)", 3)
}
