package utxo

import (
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/test/helper"
	"github.com/jinzhu/gorm"
)

func RemoveInputs(utxoList []*model.Utxo, db *gorm.DB) error {
	defer helper.MeasureTime(helper.MeasureTitle("RemoveInputs"))
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
