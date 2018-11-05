package model

import (
	"github.com/jinzhu/gorm"
)

type Tx struct {
	gorm.Model

	// tx hash
	Hash []byte `gorm:"size:32;unique;"`
	// tx 版本
	Version uint16
	// tx 类型
	TxType uint16
	// block ID
	BlockID int
	// block hash
	BlockHash []byte `gorm:"size:32;"`
	// block height
	BlockHeight uint32
	// 冻结高度
	LockUntil uint32
	// input
	Inputs []byte `gorm:"type:blob;"`
	// 输出金额
	Amount int64
	// 找零金额
	Change int64
	// 网络交易费
	Fee int64
	// 接收地址 1 字节前缀 + 32 字节地址
	SendTo []byte `gorm:"size:33;"`
	// 发送地址
	Sender []byte `gorm:"size:33;"`
	// tx 包含的数据
	Data []byte `gorm:"type:blob;"`
	// tx 签名
	Sig []byte `gorm:"type:blob;"`
}
