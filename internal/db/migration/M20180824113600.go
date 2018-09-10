package migration

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func M20180824113600() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20180824113600",
		Migrate: func(tx *gorm.DB) error {
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
				// 区块时间戳
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

			type Tx struct {
				gorm.Model

				// tx hash
				Hash []byte `gorm:"size:32;unique;"`
				// tx 版本
				Version uint16
				// tx 类型
				Type uint16
				// block ID <- hash
				BlockID string `gorm:"size:32;"`
				// block height
				BlockHash []byte `gorm:"size:32;"`
				// 冻结高度
				LockUntil int
				// 输出金额
				Amount int64
				// 网络交易费
				Fee int64
				// 接收地址 1 字节前缀 + 32 字节地址
				SendTo []byte `gorm:"size:33;"`
				// tx 包含的数据
				Data []byte `gorm:"type:blob;"`
				// tx 签名
				Sig []byte
			}

			type Utxo struct {
				gorm.Model

				TxHash      []byte `gorm:"primary_key;"`
				Destination []byte `gorm:"size:33;"`
				Amount      int64
				BlockHeight uint32
				Out         uint8
			}

			return tx.AutoMigrate(&Block{}, &Tx{}, &Utxo{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTableIfExists("block", "tx", "utxo").Error
		},
	}
}
