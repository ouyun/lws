package utxo

import (
	"testing"

	"bytes"
	"github.com/FissionAndFusion/lws/internal/constant"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/db"
	utxoService "github.com/FissionAndFusion/lws/internal/db/service/utxo"
	streamModel "github.com/FissionAndFusion/lws/internal/stream/model"
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

	sender := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

	streamTx := &streamModel.StreamTx{
		Transaction: tx,
		Sender:      sender,
	}

	utxoUpdates, err := HandleTx(dbtx, streamTx, nil)
	if err != nil {
		t.Fatalf("insert utxo by pool tx failed [%s]", err)
	}

	for _, list := range utxoUpdates {
		for _, item := range list {
			if item.OpType == constant.UTXO_UPDATE_TYPE_NEW {
				if item.UTXO == nil {
					t.Errorf("new utxo should has utxo, but nil")
				}

				if item.UTXO != nil && bytes.Compare(item.UTXO.Sender, sender) != 0 {
					t.Errorf("utxo sender expect [%s], but [%s]", sender, item.UTXO.Sender)
				}
			}
		}
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
