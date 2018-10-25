package tx

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/test/helper"
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

	handler := &BlockTxHandler{
		dbtx: connection,
		txs:  txs,
	}
	err := handler.prepareSenders()
	if err != nil {
		t.Errorf("prepare senders err: [%s]", err)
	}
	_, err = handler.insertTxs(txs[:], blockModel)
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

func TestPrepareSender(t *testing.T) {
	helper.ResetDb()
	helper.LoadTestSeed("seedBasic.sql")

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
				Data:   []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4},
			},
			VInput: []*lws.Transaction_CTxIn{
				&lws.Transaction_CTxIn{
					Hash: []byte{0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
					N:    1,
				},
			},
		},
		&lws.Transaction{
			NVersion: uint32(1),
			Hash:     []byte{1, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3},
			NAmount:  200,
			NTxFee:   100,
			CDestination: &lws.Transaction_CDestination{
				Prefix: uint32(2),
				Data:   []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
			},
			VInput: []*lws.Transaction_CTxIn{
				&lws.Transaction_CTxIn{
					Hash: []byte{1, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2},
					N:    0,
				},
			},
		},
	}

	connection := db.GetConnection()
	handler := &BlockTxHandler{
		dbtx: connection,
		txs:  txs,
	}

	err := handler.prepareSenders()
	if err != nil {
		t.Fatalf("preprare sender failed [%s]", err)
	}

	mapTxToSender := handler.GetMapTxToSender()

	if length := len(mapTxToSender); length != 2 {
		t.Fatalf("len(mapTxToSender) expect 2, but [%d]", length)
	}

	txHash := [32]byte{1, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}
	sender, ok := mapTxToSender[txHash]
	if !ok {
		t.Fatalf("expect tx hash sender, but none")
	}
	expectedSender := []byte{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}
	if bytes.Compare(sender, expectedSender) != 0 {
		t.Fatalf("expect sender [%v], but [%v]", expectedSender, sender)
	}

	txHash2 := [32]byte{1, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}
	sender2, ok := mapTxToSender[txHash2]
	if !ok {
		t.Fatalf("expect tx hash sender, but none")
	}
	expectedSender2 := []byte{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}
	if bytes.Compare(sender2, expectedSender2) != 0 {
		t.Fatalf("expect sender [%v], but [%v]", expectedSender2, sender2)
	}

}
