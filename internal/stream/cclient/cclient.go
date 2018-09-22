package cclient

import (
	"log"
	"os"

	"github.com/lomocoin/lws/internal/coreclient"
)

var primaryClient *coreclient.Client

func StartCoreClient() *coreclient.Client {
	addr := os.Getenv("CORECLIENT_URL")

	log.Printf("Connect to core client [%s]", addr)
	primaryClient = coreclient.NewTCPClient(addr)

	primaryClient.Start()

	return primaryClient
}

func GetPrimaryClient() *coreclient.Client {
	return primaryClient
}

// make test easy
func SetPrimaryClient(cli *coreclient.Client) {
	primaryClient = cli
}
