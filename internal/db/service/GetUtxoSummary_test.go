package service

import (
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/internal/db/model"
	"github.com/lomocoin/lws/test/helper"
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

func TestGetSummary(t *testing.T) {
	helper.ResetDb()

	connection := db.GetConnection()

	var amountExpected int64 = 10000
	var countExpected int = 1

	result := connection.Create(&model.Utxo{
		TxHash:      []byte("fffffffffffffffffffffffffffffff3"),
		Destination: []byte("1ffffff78901234567890123456789013"),
		Amount:      10000,
		BlockHeight: 0xFFFFFFFF,
		Out:         0,
	})

	if result.Error != nil {
		t.Errorf("create temp utxo record err: [%s]", result.Error)
	}

	sum, count, err := GetUtxoSummary([]*lws.Transaction_CTxIn{
		&lws.Transaction_CTxIn{
			Hash: []byte("fffffffffffffffffffffffffffffff3"),
			N:    0,
		},
	}, connection)
	if err != nil {
		t.Errorf("get utxo summary failed: %s", err)
	}

	if sum != amountExpected || count != countExpected {
		t.Errorf("expect (sum, count) to be (%d, %d), but got (%d, %d)", amountExpected, countExpected, sum, count)
	}
}
