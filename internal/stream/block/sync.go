package block

import (
	"bytes"
	// "github.com/huandu/go-sqlbuilder"
	"github.com/lomocoin/lws/internal/constant"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	dbmodule "github.com/lomocoin/lws/internal/db"
	model "github.com/lomocoin/lws/internal/db/model"
	"log"
	"time"
)

func handleSyncBlock(block *lws.Block) bool {
	log.Printf("Receive Block hash [%s]", block.Hash)
	// 1. 判断是否为子块
	isSubBlock := uint16(block.NType) == constant.BLOCK_TYPE_SUBSIDIARY
	var shouldWrite bool
	if isSubBlock {
		shouldWrite = validateSubBlock(block)
	} else {
		shouldWrite = validateBlock(block)
	}

	if shouldWrite {
		if err := writeBlock(block); err != nil {
			return false
		}
	}

	return true
}

func validateSubBlock(block *lws.Block) bool {
	// TODO
	// 1. 判断子块hashPrev是否在链上
	return true
}

func validateBlock(block *lws.Block) bool {
	// 1. 根据高度快速查找该区块是否已经在链上, 在链上则跳过本次操作
	if ok := isBlockExisted(uint(block.Height), block.Hash, false); ok {
		log.Printf("Block hash [%s] is already existed", block.Hash)
		return false
	}

	// 2. 判断hashPrev是否与链尾区块hash一致 或是 初始块
	if ok := isTailOrOrigin(block); !ok {
		// 3A. 不一致则启动错误恢复流程
		log.Printf("Block hash [%s] trigger recovery", block.Hash)
		// TODO start recovery
		return false
	}

	// 3B. 一致则为校验通过
	return true
}

// 判断hashPrev是否与链尾区块hash一致 或是 初始块
func isTailOrOrigin(block *lws.Block) bool {
	if uint16(block.NType) == constant.BLOCK_TYPE_GENESIS {
		return true
	}

	tail := getTailBlock()
	if tail != nil {
		log.Printf("tail [%s](%d), block [%s](%d) prev[%s]", tail.Hash, tail.Height, block.Hash, block.Height, block.HashPrev)
		if bytes.Compare(tail.Hash, block.HashPrev) == 0 && block.Height == uint32(tail.Height)+1 {
			return true
		}
	}

	return false
}

func writeBlock(block *lws.Block) error {
	ormBlock := convertBlockFromDbpToOrm(block)
	gormdb := dbmodule.GetGormDb()
	res := gormdb.Create(ormBlock)
	log.Printf("res = %+v\n", res)
	if res.Error != nil {
		return res.Error
	}

	// TODO txs
	return nil
}

func convertBlockFromDbpToOrm(block *lws.Block) *model.Block {
	ormBlock := &model.Block{
		Hash:      block.Hash,
		Version:   uint16(block.NVersion),
		Type:      uint16(block.NType),
		Prev:      block.HashPrev,
		Height:    uint(block.Height),
		Sig:       block.VchSig,
		Timestamp: time.Unix(int64(block.NTimeStamp), 0),
	}
	return ormBlock
}

func getTailBlock() *model.Block {
	block := &model.Block{}
	gormdb := dbmodule.GetGormDb()
	res := gormdb.
		Where("type != ?", constant.BLOCK_TYPE_SUBSIDIARY).
		Order("height desc").
		Take(block)
	if res.Error != nil {
		log.Println("getTailBlock failed", res.Error)
		return nil
	}
	log.Printf("tail block = %+v\n", block)
	return block
}

func isBlockExisted(height uint, hash []byte, isSubBlock bool) bool {
	log.Println("isBlockExisted called")
	gormdb := dbmodule.GetGormDb()
	var count int

	tx := gormdb.Model(&model.Block{}).Where("height = ? AND hash = ?", height, hash)
	if isSubBlock {
		tx = tx.Where("type = ?", constant.BLOCK_TYPE_SUBSIDIARY)
	} else {
		// tx = tx.Where("type != ?", constant.BLOCK_TYPE_SUBSIDIARY)
		tx = tx.Where("type != ?", 0x0002)
	}

	res := tx.Count(&count)
	log.Printf("res = %+v\n", res)

	// sb := sqlbuilder.NewSelectBuilder()
	// sb.Select("COUNT(id)")
	// sb.From("block")
	// sb.Where(
	// 	sb.E("height", height),
	// 	sb.E("hash", hash),
	// )

	// if isSubBlock {
	// 	sb.Where(
	// 		sb.E("type", constant.BLOCK_TYPE_SUBSIDIARY),
	// 	)
	// } else {
	// 	sb.Where(
	// 		sb.NE("type", constant.BLOCK_TYPE_SUBSIDIARY),
	// 	)
	// }

	// sql, args := sb.Build()

	// log.Printf("sql = %+v\n", sql)
	// log.Printf("args = %+v\n", args)

	// var count int
	// db := gormdb.DB()
	// err := db.QueryRow(sql, args...).Scan(&count)
	// if err != nil {
	// 	log.Println("query row error: ", err)
	// }

	log.Printf("count = %+v\n", count)
	return count == 1
}
