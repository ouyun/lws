package utxo

import (
	"testing"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/db"
	utxoService "github.com/FissionAndFusion/lws/internal/db/service/utxo"
	"github.com/FissionAndFusion/lws/test/helper"
)

func TestInsertPoolTx(t *testing.T) {
	helper.ResetDb()
	helper.LoadTestSeed("seedBasic.sql")

	txHash := []byte{1, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}

	tx := &lws.Transaction{
		NVersion: uint32(1),
		Hash:     txHash,
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

	connection := db.GetConnection()

	dbtx := connection.Begin()

	_, err := HandleTx(dbtx, tx, nil)
	if err != nil {
		t.Fatalf("insert utxo by pool tx failed [%s]", err)
	}

	_, err = utxoService.GetByTxHash(txHash, dbtx)
	if err != nil {
		t.Fatalf("get utxos by tx hash failed [%s]", err)
	}

	dbtx.Commit()
}

// func TestInsertBlockTxs() {
// 	helper.ResetDb()
// 	helper.LoadTestSeed("seedBasic.sql")

// }
