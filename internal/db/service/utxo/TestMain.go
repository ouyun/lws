package utxo

import (
	"github.com/lomocoin/lws/internal/db"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	connection := db.GetConnection()
	connection.LogMode(true)

	exitCode := m.Run()

	connection.Close()
	os.Exit(exitCode)
}
