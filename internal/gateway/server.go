package gateway

import (
	"fmt"
	"github.com/lomocoin/lws/internal/gateway/mqtt"
)

type Server struct {
	Status int
}

func (s *Server) Start() {
	s.Status = 1
	p := &mqtt.Program{Id: "LWS"}
	mqtt.Run(p)
	fmt.Printf("gateway server started (status: %d)", 3)
}
