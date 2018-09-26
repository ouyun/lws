package main

import (
	"github.com/FissionAndFusion/lws/internal/gateway"
)

func main() {
	server := gateway.Server{}
	defer server.Start()
}
