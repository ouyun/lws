package stream

import (
	"fmt"
)

type Server struct {
	Status int
}

func (s *Server) Start() {
	fmt.Println("sync server started")
}
