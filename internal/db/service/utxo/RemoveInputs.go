package utxo

import (
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/jinzhu/gorm"
)

func RemoveInputs(utxoList []*model.Utxo, db *gorm.DB) error {
	query := db.Model(&model.Utxo{})

	for _, utxo := range utxoList {
		query = query.Or(map[string]interface{}{
			"tx_hash": utxo.TxHash,
			"out":     utxo.Out,
		})
	}

	result := query.Unscoped().Delete(model.Utxo{})

	if result.Error != nil {
		return result.Error
	}

	return nil
}
