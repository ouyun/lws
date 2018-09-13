package tx

import (
	"log"
	"os"
	"testing"

	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	dbmodule "github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/internal/db/model"
	"github.com/lomocoin/lws/test/helper"
)

func TestMain(m *testing.M) {
	db := dbmodule.GetGormDb()
	db.LogMode(true)

	exitCode := m.Run()

	db.Close()
	os.Exit(exitCode)
}

func TestInsertTxs(t *testing.T) {
	helper.ResetDb()

	txs := []*lws.Transaction{
		&lws.Transaction{
			NVersion: uint32(1),
			Hash:     []byte("12345678901234567890123456789012"),
			NAmount:  100,
			NTxFee:   11,
			CDestination: &lws.Transaction_CDestination{
				Prefix: uint32(1),
				Data:   []byte("ffffff78901234567890123456789013"),
			},
			VInput: []*lws.Transaction_CTxIn{
				&lws.Transaction_CTxIn{
					Hash: []byte("fffffffffffffffffffffffffffffff3"),
					N:    0,
				},
				&lws.Transaction_CTxIn{
					Hash: []byte("fffffffffffffffffffffffffffffff5"),
					N:    1,
				},
			},
		},
		&lws.Transaction{
			NVersion: uint32(1),
			Hash:     []byte("12345678901234567890123456789013"),
			NAmount:  200,
			NTxFee:   10,
			CDestination: &lws.Transaction_CDestination{
				Prefix: uint32(0),
				Data:   []byte("ffffff78901234567890123456789015"),
			},
		},
	}

	gormdb := dbmodule.GetGormDb()
	handler := &BlockTxHandler{
		dbtx: gormdb,
	}

	ormBlock := &model.Block{
		Hash:   []byte("33333333333333333333333333333333"),
		Height: uint32(2),
	}
	ormBlock.ID = 1

	err := handler.insertTxs(txs[:], ormBlock)
	if err != nil {
		t.Errorf("insert txs err: [%s]", err)
	}
}

func TestQueryExistanceTx(t *testing.T) {
	helper.ResetDb()

	gormdb := dbmodule.GetGormDb()
	handler := &BlockTxHandler{
		dbtx: gormdb,
	}
	hashes := [][]byte{
		[]byte("12345678901234567890123456789012"),
		[]byte("00000000001234567890123456789012"),
	}

	newHashes, err := handler.queryExistanceTxids(hashes)
	if err != nil {
		t.Errorf("queryExistanceTxids err: [%s]", err)
	}

	log.Printf("newHashes = %+v\n", newHashes)

	// TODO seed data  check len / values etc.
}
