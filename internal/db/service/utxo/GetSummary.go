package utxo

import (
	"github.com/jinzhu/gorm"
	"github.com/lomocoin/lws/internal/db/model"
)

func GetSummary(utxoList []*model.Utxo, db *gorm.DB) (int64, int, error) {
	type sumResult struct {
		Sum   int64
		Count int
	}
	var r sumResult
	query := db.Model(&model.Utxo{}).Select("SUM(utxo.amount) as sum, COUNT(*) as count")

	for _, utxo := range utxoList {
		query = query.Or(map[string]interface{}{
			"tx_hash": utxo.TxHash,
			"out":     utxo.Out,
		})
	}

	result := query.Scan(&r)

	if result.Error != nil {
		return 0, 0, result.Error
	}

	return r.Sum, r.Count, nil
}
