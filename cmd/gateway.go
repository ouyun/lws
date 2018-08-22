package main

import (
	"github.com/lomocoin/lws/internal/gateway"
)

func main() {
	server := gateway.Server{}
	defer server.Start()
}
