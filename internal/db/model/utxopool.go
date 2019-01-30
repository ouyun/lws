package model

import (
	// "github.com/jinzhu/gorm"
	"time"
)

type UtxoPool struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time

	TxHash      []byte `gorm:"size:32;index:idx_tx_pool_hash_out"`
	Destination []byte `gorm:"size:33;"`
	Amount      int64
	// BlockHeight uint32
	Out      uint8  `gorm:"index:idx_tx_hash_out"` // index 0 -> destination, 1 -> change
	Idx      []byte `gorm:"size:33;index:idx_utxo_pool_idx"`
	IsDelete bool   // true for delete utxo record, otherwise new added utxo record
	TxOwner  []byte `gorm:"size:32"` // add foreign key to TxPool Hash
}
