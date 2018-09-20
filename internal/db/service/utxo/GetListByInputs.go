package utxo

import (
	"github.com/jinzhu/gorm"
	"github.com/lomocoin/lws/internal/db/model"
)

func GetListByInputs(utxoList []*model.Utxo, db *gorm.DB) ([]*model.Utxo, error) {
	query := db.Model(&model.Utxo{})

	for _, utxo := range utxoList {
		query = query.Or(map[string]interface{}{
			"tx_hash": utxo.TxHash,
			"out":     utxo.Out,
		})
	}

	var utxos []*model.Utxo
	result := query.Find(&utxos)

	if result.Error != nil {
		return nil, result.Error
	}

	return utxos, nil
}
