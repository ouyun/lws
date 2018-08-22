package gateway

import (
	"fmt"
)

type Server struct {
	Status int
}

func (s *Server) Start() {
	s.Status = 1
	fmt.Printf("gateway server started (status: %d)", 3)
}
