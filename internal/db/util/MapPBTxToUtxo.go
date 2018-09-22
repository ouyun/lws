package util

import (
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/lomocoin/lws/internal/db/model"
)

func MapPBTxToUtxo(tx *lws.Transaction, blockHeight uint32) []*model.Utxo {
	utxoList := make([]*model.Utxo, len(tx.VInput))
	for i, input := range tx.VInput {
		utxoList[i] = &model.Utxo{
			TxHash:      input.Hash,
			Out:         uint8(input.N),
			Amount:      tx.NAmount,
			Destination: MapPBDestinationToBytes(tx.CDestination),
			BlockHeight: blockHeight,
		}
	}
	return utxoList
}
