package model

import (
	"github.com/jinzhu/gorm"
)

type Utxo struct {
	gorm.Model

	TxHash      []byte `gorm:"size:32;index:idx_tx_hash_out"`
	Destination []byte `gorm:"size:33;"`
	Amount      int64
	BlockHeight uint32
	Out         uint8  `gorm:"index:idx_tx_hash_out"` // index 0 -> destination, 1 -> change
	Idx         []byte `gorm:"size:33;index:idx_utxo_idx"`
}
