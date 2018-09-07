package block

import (
	"bytes"
	"fmt"
	"log"

	"github.com/lomocoin/lws/internal/constant"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	dbmodule "github.com/lomocoin/lws/internal/db"
	model "github.com/lomocoin/lws/internal/db/model"
)

// return error and bool skiped
func handleSyncBlock(block *lws.Block) (error, bool) {
	var err error
	// log.Printf("Receive Block hash [%s]", block.Hash)
	log.Printf("Receive Block hash v [%+v] type[%d] (#%d)", block.Hash, block.NType, block.NHeight)

	// 1. 判断是否为子块
	isSubBlock := uint16(block.NType) == constant.BLOCK_TYPE_SUBSIDIARY
	var skip bool
	if isSubBlock {
		err, skip = validateSubBlock(block)
	} else {
		err, skip = validateBlock(block)
	}

	// recovery or skip
	if err != nil || skip {
		return err, skip
	}

	err = writeBlock(block)
	return err, skip
}

// error, skip
func validateSubBlock(block *lws.Block) (error, bool) {
	// 1. 判断子块是否在链上
	if ok := isBlockExisted(block.NHeight, block.Hash, true); ok {
		log.Printf("Block hash [%s] is already existed", block.Hash)
		return nil, true
	}
	// 2. 判断所连的主块是否在链上
	if ok := isBlockExisted(block.NHeight, block.HashPrev, false); !ok {
		log.Printf("subBlock HashPrev [%s](#%d) is not existed, skip the sub block [%s]", block.HashPrev, block.NHeight, block)
		return nil, true
	}
	return nil, false
}

// error, skip
func validateBlock(block *lws.Block) (error, bool) {
	// 1. 根据高度快速查找该区块是否已经在链上, 在链上则跳过本次操作
	if ok := isBlockExisted(block.NHeight, block.Hash, false); ok {
		log.Printf("Block hash [%s](#%d) is already existed", block.Hash, block.NHeight)
		return nil, true
	}

	// 2. 判断hashPrev是否与链尾区块hash一致 或是 初始块
	if ok := isTailOrOrigin(block); !ok {
		// 3A. 不一致则启动错误恢复流程
		log.Printf("Block hash [%s] trigger recovery", block.Hash)
		// TODO start recovery
		err := fmt.Errorf("trigger recovery")
		return err, false
	}

	// 3B. 一致则为校验通过
	return nil, false
}

// 判断hashPrev是否与链尾区块hash一致 或是 初始块
func isTailOrOrigin(block *lws.Block) bool {
	if uint16(block.NType) == constant.BLOCK_TYPE_GENESIS {
		return true
	}

	tail := GetTailBlock()
	if tail != nil {
		// log.Printf("tail [%s](%d), block [%s](%d) prev[%s]", tail.Hash, tail.Height, block.Hash, block.Height, block.HashPrev)
		if bytes.Compare(tail.Hash, block.HashPrev) == 0 && block.NHeight == tail.Height+1 {
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
		BlockType: uint16(block.NType),
		Prev:      block.HashPrev,
		Height:    block.NHeight,
		Sig:       block.VchSig,
		Tstamp:    block.NTimeStamp,
	}
	return ormBlock
}

func GetTailBlock() *model.Block {
	block := &model.Block{}
	gormdb := dbmodule.GetGormDb()
	res := gormdb.
		Where("block_type != ?", constant.BLOCK_TYPE_SUBSIDIARY).
		Order("height desc").
		Take(block)
	if res.Error != nil {
		log.Println("GetTailBlock failed", res.Error)
		return nil
	}
	log.Printf("Tail: [%s](%d) type[%d]", block.Hash, block.Height, block.BlockType)
	return block
}

func isBlockExisted(height uint32, hash []byte, isSubBlock bool) bool {
	gormdb := dbmodule.GetGormDb()
	var count int

	tx := gormdb.Model(&model.Block{}).Where("height = ? AND hash = ?", height, hash)
	if isSubBlock {
		tx = tx.Where("block_type = ?", constant.BLOCK_TYPE_SUBSIDIARY)
	} else {
		tx = tx.Where("block_type != ?", constant.BLOCK_TYPE_SUBSIDIARY)
	}

	res := tx.Count(&count)
	if res.Error != nil {
		log.Println("isBlockExisted failed", res.Error)
		return false
	}
	return count == 1
}
