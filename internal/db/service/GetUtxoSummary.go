package service

import (
	"github.com/jinzhu/gorm"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/lomocoin/lws/internal/db/model"
)

func GetUtxoSummary(utxos []*lws.Transaction_CTxIn, db *gorm.DB) (int64, int, error) {
	type sumResult struct {
		Sum   int64
		Count int
	}
	var r sumResult
	query := db.Model(&model.Utxo{}).Select("SUM(utxo.amount) as sum, COUNT(*) as count")

	for _, utxo := range utxos {
		query = query.Or(map[string]interface{}{
			"tx_hash": utxo.Hash,
			"out":     utxo.N,
		})
	}

	result := query.Scan(&r)

	if result.Error != nil {
		return 0, 0, result.Error
	}

	return r.Sum, r.Count, nil
}
