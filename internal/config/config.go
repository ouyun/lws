package config

import (
	"fmt"
	"os"
)

type Configs struct {
	UTXO_UPDATE_QUEUE_NAME string
}

var Config *Configs

func InitSettings() *Configs {

	identifier := os.Getenv("INSTANCE_ID")

	Config := &Configs{
		UTXO_UPDATE_QUEUE_NAME: fmt.Sprintf("LWS%s.utxoupdate.", identifier),
	}

	return Config
}
