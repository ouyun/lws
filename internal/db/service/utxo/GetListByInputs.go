package utxo

import (
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/test/helper"
	"github.com/jinzhu/gorm"
)

func GetListByInputs(utxoList []*model.Utxo, db *gorm.DB) ([]*model.Utxo, error) {
	defer helper.MeasureTime(helper.MeasureTitle("query GetListByInputs"))
	query := db.Model(&model.Utxo{}).Where(&model.Utxo{})

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
