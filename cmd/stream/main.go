package main

import (
	"github.com/FissionAndFusion/lws/internal/stream"
)

func main() {
	server := new(stream.Server)
	defer server.Start()
}
