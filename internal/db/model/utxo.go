package model

import (
	"github.com/jinzhu/gorm"
)

type Utxo struct {
	gorm.Model

	TxHash      []byte `gorm:"primary_key;"`
	Destination []byte `gorm:"size:33;"`
	Amount      int64
	BlockHeight uint32
	Out         uint8
}
