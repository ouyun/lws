package utxo

import (
	"bytes"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/test/helper"
	"testing"
)

func TestGetListByInputs(t *testing.T) {
	helper.ResetDb()

	connection := db.GetConnection()

	testUtxo := model.Utxo{
		TxHash:      []byte{1, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		Destination: []byte{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
		Amount:      10000,
		BlockHeight: 0xFFFFFFFF,
		Out:         0,
	}

	testUtxo2 := model.Utxo{
		TxHash:      []byte{1, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		Destination: []byte{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3},
		Amount:      100,
		BlockHeight: 0xFFFFFFFF,
		Out:         1,
	}

	result := connection.Create(&testUtxo)

	if result.Error != nil {
		t.Errorf("create temp utxo record err: [%s]", result.Error)
	}

	result = connection.Create(&testUtxo2)

	if result.Error != nil {
		t.Errorf("create temp utxo2 record err: [%s]", result.Error)
	}

	list, err := GetListByInputs([]*model.Utxo{
		&model.Utxo{
			TxHash: []byte{1, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			Out:    0,
		},
	}, connection)

	if err != nil {
		t.Errorf("get utxo summary failed: %s", err)
	}

	if len(list) != 1 {
		t.Errorf("expect results list length to be %d, but got %d", 1, len(list))
	}

	item := list[0]

	if !bytes.Equal(item.TxHash, testUtxo.TxHash) || item.Out != testUtxo.Out {
		t.Errorf("expect result item to match to created one, but not match")
	}
}
