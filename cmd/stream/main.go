package main

import (
	"log"
	"os"

	"github.com/FissionAndFusion/lws/internal/stream"
	"github.com/hashicorp/logutils"
)

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "DEBUG"
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(logLevel),
		Writer:   os.Stdout,
	}
	log.SetOutput(filter)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	server := new(stream.Server)
	defer server.Start()
}
