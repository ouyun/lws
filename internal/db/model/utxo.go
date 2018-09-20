package model

import (
	"github.com/jinzhu/gorm"
)

type Utxo struct {
	gorm.Model

	TxHash      []byte `gorm:"primary_key;index:idx_tx_hash_out"`
	Destination []byte `gorm:"size:33;"`
	Amount      int64
	BlockHeight uint32
	Out         uint8 `gorm:"index:idx_tx_hash_out"` // index 0 -> destination, 1 -> change
}
