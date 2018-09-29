package tx

import (
	"testing"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/test/helper"
)

func TestInsertPoolTx(t *testing.T) {
	helper.ResetDb()
	helper.LoadTestSeed("seedBasic.sql")

	tx := &lws.Transaction{
		NVersion: uint32(1),
		Hash:     []byte{1, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
		NAmount:  3000000,
		NTxFee:   100,
		CDestination: &lws.Transaction_CDestination{
			Prefix: uint32(2),
			Data:   []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
		},
		VInput: []*lws.Transaction_CTxIn{
			&lws.Transaction_CTxIn{
				Hash: []byte{0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
				N:    1,
			},
		},
	}

	err := StartPoolTxHandler(tx)
	if err != nil {
		t.Fatalf("insert pool tx failed [%s]", err)
	}
}

func TestPoolTxGetNilSender(t *testing.T) {
	tx := &model.Tx{
		Hash:      []byte{1, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
		Version:   uint16(1),
		TxType:    uint16(0x0000),
		LockUntil: uint32(0),
		Amount:    300000,
		Fee:       100,
		SendTo:    []byte{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
	}

	connection := db.GetConnection()
	sender, err := getSingleTxSender(connection, tx)
	if err != nil {
		t.Fatalf("query tx pool nil sender failed [%s]", err)
	}
	if sender != nil {
		t.Fatalf("sender expected to be nil, but [%v]", sender)
	}
}

func TestPoolTxGetNonExistInputSender(t *testing.T) {
	tx := &model.Tx{
		Hash:      []byte{1, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
		Version:   uint16(1),
		TxType:    uint16(0x0000),
		LockUntil: uint32(0),
		Amount:    300000,
		Fee:       100,
		SendTo:    []byte{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
		Inputs:    []byte{1, 2, 3, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
	}

	connection := db.GetConnection()
	sender, err := getSingleTxSender(connection, tx)
	if err != nil {
		t.Fatalf("query tx pool nil sender failed [%s]", err)
	}
	if sender != nil {
		t.Fatalf("sender expected to be nil, but [%v]", sender)
	}
}
