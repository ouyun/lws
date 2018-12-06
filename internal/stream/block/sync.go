package block

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"sync"

	"github.com/FissionAndFusion/lws/internal/constant"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	blockService "github.com/FissionAndFusion/lws/internal/db/service/block"
	"github.com/FissionAndFusion/lws/internal/gateway/mqtt"
	"github.com/FissionAndFusion/lws/internal/stream/tx"
	"github.com/FissionAndFusion/lws/test/helper"
)

// return error and bool skiped
func handleSyncBlock(block *lws.Block, shouldRecover bool) (error, bool) {
	defer helper.MeasureTime(helper.MeasureTitle("handle block [%s](#%d)", hex.EncodeToString(block.Hash), block.NHeight))

	var err error
	// log.Printf("Receive Block hash [%s]", block.Hash)
	log.Printf("[INFO] Receive Block hash v [%s] type[%d] (#%d)", hex.EncodeToString(block.Hash), block.NType, block.NHeight)

	var skip bool
	var write bool
	err, skip, write = validateBlock(block, shouldRecover)

	// recovery
	if err != nil {
		return err, skip
	}

	if write {
		err = writeBlock(block)
	}

	return err, skip
}

// error, skip
func validateBlock(block *lws.Block, shouldRecover bool) (error, bool, bool) {
	// 1. 根据高度快速查找该区块是否已经在链上, 在链上则跳过本次操作
	if ok := isBlockExisted(block.NHeight, block.Hash); ok {
		log.Printf("[DEBUG] Block hash [%s](#%d) is already existed", hex.EncodeToString(block.Hash), block.NHeight)
		return nil, true, false
	}

	// 2. 判断hashPrev是否与链尾区块hash一致 或是 初始块
	if ok := isTailOrOrigin(block); !ok {
		// 3A. 不一致则启动错误恢复流程
		hashStr := hex.EncodeToString(block.Hash)
		log.Printf("[INFO] Block hash [%s], prev[%s] trigger recovery", hashStr, hex.EncodeToString(block.HashPrev))
		// start recovery
		if shouldRecover {
			FetchBlocks(block)
			log.Printf("[INFO] Block hash [%s] recovery done", hashStr)
		}

		err := fmt.Errorf("trigger recovery")
		// skip current trigger-block (fetchBlocks will handle it)
		return err, true, false
	}

	// 3B. 一致则为校验通过
	return nil, true, true
}

// 判断hashPrev是否与链尾区块hash一致 或是 初始块
func isTailOrOrigin(block *lws.Block) bool {
	if uint16(block.NType) == constant.BLOCK_TYPE_GENESIS {
		return true
	}

	tail := blockService.GetTailBlock()
	if tail != nil {
		// log.Printf("tail [%s](%d), block [%s](%d) prev[%s]", tail.Hash, tail.Height, block.Hash, block.NHeight, block.HashPrev)
		// new primary block
		isRightHeight := block.NHeight == tail.Height+1
		if uint16(block.NType) == constant.BLOCK_TYPE_EXTENDED {
			// new extent block
			isRightHeight = block.NHeight == tail.Height
		}

		if bytes.Compare(tail.Hash, block.HashPrev) == 0 && isRightHeight {
			return true
		}
	}

	return false
}

func writeBlock(block *lws.Block) error {
	ormBlock := convertBlockFromDbpToOrm(block)
	connection := db.GetConnection()
	dbtx := connection.Begin()
	// dbtx := connection

	res := dbtx.Create(ormBlock)
	log.Printf("[INFO] write block (#%d) done", ormBlock.Height)
	if res.Error != nil {
		dbtx.Rollback()
		return res.Error
	}

	// txs
	txs := make([]*lws.Transaction, 0)
	txs = append(txs, block.Vtx...)

	// append tx mint to tx list
	if block.TxMint != nil {
		txs = append(txs, block.TxMint)
	} else {
		log.Printf("[INFO] block [%s](#%d) has no tx mint field", hex.EncodeToString(block.Hash), block.NHeight)
	}
	updates, err := tx.StartBlockTxHandler(dbtx, txs, ormBlock)
	if err != nil {
		log.Printf("[DEBUG] sync rollback block [%s](#%d)", hex.EncodeToString(block.Hash), block.NHeight)
		dbtx.Rollback()
		return err
	}

	dbtx.Commit()

	defer helper.MeasureTime(helper.MeasureTitle("block send utxo update [%s](#%d)", hex.EncodeToString(block.Hash), block.NHeight))
	var wg sync.WaitGroup
	for destination, item := range updates {
		wg.Add(1)
		go mqtt.NewUTXOUpdate(item, destination[:], &wg)
	}
	wg.Wait()

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

func isBlockExisted(height uint32, hash []byte) bool {
	connection := db.GetConnection()
	var count int

	tx := connection.Model(&model.Block{}).Where("height = ? AND hash = ?", height, hash)

	res := tx.Count(&count)
	if res.Error != nil {
		log.Println("[ERROR] isBlockExisted failed", res.Error)
		return false
	}
	return count == 1
}
