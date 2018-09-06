package model

import (
	"github.com/jinzhu/gorm"
)

type Block struct {
	gorm.Model

	// block ID <- hash
	Hash []byte `gorm:"primary_key;"`
	// 区块版本
	Version uint16
	// 区块类型
	BlockType uint16
	// 前一区块的 hash
	Prev []byte
	// 区块时间戳 timestamp
	Tstamp uint32
	// 两两校验
	Merkle string `gorm:"size:32;"`
	// 区块高度
	Height uint32
	// 矿工打包费的 tx id
	MintTXID []byte
	// 区块签名
	Sig []byte
}
