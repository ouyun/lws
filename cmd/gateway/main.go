package main

import (
	"flag"
	"github.com/FissionAndFusion/lws/internal/gateway"
	"log"
	"os"
)

func main() {
	var lwsId string

	flag.StringVar(&lwsId, "id", "", "lwsId")
	flag.Parse()
	if lwsId == "" {
		lwsId = os.Getenv("LWS_ID")
	}
	log.Printf(lwsId)
	server := gateway.Server{
		Id: lwsId,
	}
	defer server.Start()
}
