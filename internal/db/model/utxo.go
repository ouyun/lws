package model

import (
	"github.com/jinzhu/gorm"
)

type Utxo struct {
	gorm.Model

	Destination []byte `gorm:"size:33;"`
	Amount      int64
	TxID        int
}
