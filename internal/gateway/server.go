package gateway

import (
	"fmt"
	mqtt "github.com/FissionAndFusion/lws/internal/gateway/mqtt"
)

type Server struct {
	Status int
}

func (s *Server) Start() {
	s.Status = 1
	p := &mqtt.Program{Id: "LWS000010", Topic: "LWS01", IsLws: true}
	mqtt.Run(p)
	fmt.Printf("gateway server started (status: %d)", 3)
}
