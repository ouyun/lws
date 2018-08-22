package main

import (
	"github.com/lomocoin/lws/internal/sync"
)

func main() {
	server := new(sync.Server)
	defer server.Start()
}
