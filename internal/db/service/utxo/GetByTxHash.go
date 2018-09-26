package utxo

import (
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/jinzhu/gorm"
)

func GetByTxHash(hash []byte, db *gorm.DB) ([]*model.Utxo, error) {
	var utxos []*model.Utxo
	result := db.Model(&model.Utxo{}).Where(&model.Utxo{TxHash: hash}).Order("`out` asc").Find(&utxos)

	if result.Error != nil {
		return nil, result.Error
	}

	return utxos, nil
}
