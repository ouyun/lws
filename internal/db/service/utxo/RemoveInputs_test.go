package utxo

import (
	"github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/internal/db/model"
	"github.com/lomocoin/lws/test/helper"
	"testing"
)

func TestRemoveInputs(t *testing.T) {
	helper.ResetDb()

	connection := db.GetConnection()

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

	err := RemoveInputs([]*model.Utxo{
		&model.Utxo{
			TxHash: []byte("fffffffffffffffffffffffffffffff3"),
			Out:    0,
		},
	}, connection)

	if err != nil {
		t.Errorf("get utxo summary failed: %s", err)
	}
}
