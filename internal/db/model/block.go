package model

import (
	"time"
)

type Block struct {
	// block ID <- hash
	ID string `gorm:"size:32;primary_key;"`
	// 区块版本
	Version uint16
	// 区块类型
	Type uint16
	// 前一区块的 hash
	Prev string `gorm:"size:32;"`
	// 区块时间戳
	Timestamp time.Time
	// 两两校验
	Merkle string `gorm:"size:32;"`
	// 区块高度
	Height uint
	// 矿工打包费的 tx id
	MintTXID string `gorm:"size:32;"`
	// 区块签名
	Sig []byte
}
