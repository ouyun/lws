package tx

import (
	"log"
	"os"
	"testing"

	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/internal/db/model"
	"github.com/lomocoin/lws/test/helper"
)

func TestMain(m *testing.M) {
	connection := db.GetConnection()
	connection.LogMode(true)

	exitCode := m.Run()

	connection.Close()
	os.Exit(exitCode)
}

func TestInsertTxs(t *testing.T) {
	helper.ResetDb()
	helper.LoadTestSeed("seedBasic.sql")

	blockModel := &model.Block{
		Hash:     []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 4},
		Prev:     []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3},
		Tstamp:   uint32(1537502211),
		Height:   uint32(3),
		MintTXID: []byte{1, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	}

	connection := db.GetConnection().Begin()
	result := connection.Create(blockModel)
	if result.Error != nil {
		connection.Rollback()
		t.Errorf("insert block err: %s", result.Error)
	}

	handler := &BlockTxHandler{
		dbtx: connection,
	}

	txs := []*lws.Transaction{
		&lws.Transaction{
			NVersion: uint32(1),
			Hash:     []byte{1, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			NAmount:  15000100,
			NTxFee:   0,
			CDestination: &lws.Transaction_CDestination{
				Prefix: uint32(2),
				Data:   []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
			},
		},
		&lws.Transaction{
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
		},
	}

	err := handler.insertTxs(txs[:], blockModel)
	connection.Commit()
	if err != nil {
		t.Errorf("insert txs err: [%s]", err)
	}
}

func TestQueryExistanceTx(t *testing.T) {
	helper.ResetDb()
	helper.LoadTestSeed("seedBasic.sql")

	connection := db.GetConnection()
	handler := &BlockTxHandler{
		dbtx: connection,
	}
	hashes := [][]byte{
		[]byte{0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		[]byte{0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	}

	newHashes, err := handler.queryExistanceTxids(hashes)
	if err != nil {
		t.Errorf("queryExistanceTxids err: [%s]", err)
	}

	log.Printf("newHashes = %+v\n", newHashes)

	// TODO seed data  check len / values etc.
}
