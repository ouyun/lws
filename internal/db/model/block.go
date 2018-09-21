package model

import (
	"github.com/jinzhu/gorm"
)

type Block struct {
	gorm.Model

	// block ID <- hash
	Hash []byte `gorm:"primary_key;type:varbinary;"`
	// 区块版本
	Version uint16 `gorm:"default:1;"`
	// 区块类型
	BlockType uint16 `gorm:"default:1;"`
	// 前一区块的 hash
	Prev []byte `gorm:"type:varbinary;"`
	// 区块时间戳 timestamp
	Tstamp uint32
	// 两两校验
	Merkle string `gorm:"size:32;"`
	// 区块高度
	Height uint32
	// 矿工打包费的 tx id
	MintTXID []byte `gorm:"size:32;type:varbinary;"`
	// 区块签名
	Sig []byte
}
