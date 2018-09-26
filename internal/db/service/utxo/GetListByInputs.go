package utxo

import (
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/jinzhu/gorm"
)

func GetListByInputs(utxoList []*model.Utxo, db *gorm.DB) ([]*model.Utxo, error) {
	query := db.Model(&model.Utxo{}).Where(&model.Utxo{})

	for _, utxo := range utxoList {
		query = query.Or(&model.Utxo{
			TxHash: utxo.TxHash,
			Out:    utxo.Out,
		})
	}

	var utxos []*model.Utxo
	result := query.Find(&utxos)

	if result.Error != nil {
		return nil, result.Error
	}

	return utxos, nil
}
