package main

import (
	"flag"
	"github.com/FissionAndFusion/lws/internal/gateway"
	"github.com/hashicorp/logutils"
	"log"
	"os"
)

func main() {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "VERB", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("DEBUG"),
		Writer:   os.Stdout,
	}
	log.SetOutput(filter)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

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
	server.Start()
}
