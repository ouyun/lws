package config

import (
	"fmt"
	"log"
	"os"
)

type Configs struct {
	UTXO_UPDATE_QUEUE_NAME string
	INSTANCE_ID            string
}

var config *Configs

func InitConfigs() *Configs {

	identifier := os.Getenv("INSTANCE_ID")

	config = &Configs{
		UTXO_UPDATE_QUEUE_NAME: fmt.Sprintf("LWS%s.utxoupdate.", identifier),
		INSTANCE_ID:            identifier,
	}

	log.Printf("[INFO] init config %v", config)

	return config
}

func GetConfig() *Configs {
	if config == nil {
		return InitConfigs()
	}
	return config
}
