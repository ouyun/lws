package utxo

import (
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/test/helper"
	"testing"
)

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

	sum, count, err := GetSummary([]*model.Utxo{
		&model.Utxo{
			TxHash: []byte("fffffffffffffffffffffffffffffff3"),
			Out:    0,
		},
	}, connection)
	if err != nil {
		t.Errorf("get utxo summary failed: %s", err)
	}

	if sum != amountExpected || count != countExpected {
		t.Errorf("expect (sum, count) to be (%d, %d), but got (%d, %d)", amountExpected, countExpected, sum, count)
	}
}
