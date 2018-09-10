package tx

import (
	"log"
	"os"
	"testing"

	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	dbmodule "github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/testhelper"
)

func TestMain(m *testing.M) {
	db := dbmodule.GetGormDb()
	db.LogMode(true)

	exitCode := m.Run()

	db.Close()
	os.Exit(exitCode)
}

func TestInsertTxs(t *testing.T) {
	testhelper.ResetDb()

	txs := []*lws.Transaction{
		&lws.Transaction{
			NVersion: uint32(1),
			Hash:     []byte("12345678901234567890123456789012"),
			NAmount:  100,
			NTxFee:   11,
		},
		&lws.Transaction{
			NVersion: uint32(1),
			Hash:     []byte("12345678901234567890123456789013"),
			NAmount:  200,
			NTxFee:   10,
		},
	}

	gormdb := dbmodule.GetGormDb()
	handler := &BlockTxHandler{
		dbtx: gormdb,
	}

	err := handler.insertTxs(txs[:], 1)
	if err != nil {
		t.Errorf("insert txs err: [%s]", err)
	}
}

func TestQueryExistanceTx(t *testing.T) {
	testhelper.ResetDb()

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
