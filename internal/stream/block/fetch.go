package block

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/FissionAndFusion/lws/internal/coreclient"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	dbmodule "github.com/FissionAndFusion/lws/internal/db"
	model "github.com/FissionAndFusion/lws/internal/db/model"
	blockService "github.com/FissionAndFusion/lws/internal/db/service/block"
	utxoService "github.com/FissionAndFusion/lws/internal/db/service/utxo"
	"github.com/golang/protobuf/ptypes"

	// "github.com/FissionAndFusion/lws/internal/db/model"
	cclientModule "github.com/FissionAndFusion/lws/internal/coreclient/instance"
)

// type FetchRangeCalculator struct {
// 	TriggerBlock *lws.Block
// 	tailBlock    *model.Block
// }

const (
	FETCH_NUMBER              = 10
	TXPOOL_HEIGHT             = 0xFFFFFFFF
	STALE_BLOCK_HEIGHT_OFFSET = 10
)

type BlockFetcher struct {
	FetchNumber          int
	TriggerBlock         *lws.Block
	isTriggerBlockSynced bool
}

func FetchBlocks(triggerBlock *lws.Block) {
	b := &BlockFetcher{
		FetchNumber:  FETCH_NUMBER,
		TriggerBlock: triggerBlock,
	}

	b.startFetchBlocks(triggerBlock)
}

func isSyncDone(tail *model.Block, triggerBlock *lws.Block) bool {
	if tail == nil {
		return false
	}

	if tail.Height > triggerBlock.NHeight {
		return true
	}

	if tail.Height == triggerBlock.NHeight && bytes.Compare(tail.Hash, triggerBlock.Hash) != 0 {
		return true
	}

	return false
}

func (b *BlockFetcher) startFetchBlocks(triggerBlock *lws.Block) {
	// cclient := stream.GetPrimaryClient()
	log.Println("[DEBUG] start Fetch blocks")

	// 1. if the fork chain is detected, remove the bad-chain data
	tail := b.checkForkedChain()

	for ; tail == nil || !b.isTriggerBlockSynced; tail = blockService.GetTailBlock() {
		var hash []byte
		if tail != nil {
			hash = tail.Hash
		}

		b.fetchAndHandleBlocks(hash)
	}

	// clear stable blocks need to create anther new rabbitmq connection
	// so only if height offset larger than HEIGHT_OFFSET, we will clear stale blocks
	if tail != nil && (tail.Height-triggerBlock.NHeight > STALE_BLOCK_HEIGHT_OFFSET) {
		log.Printf("[INFO] clear stale blocks start with tail[#%d] trigger[#%d]", tail.Height, triggerBlock.NHeight)
		clearStaleBlocksInQueue(tail.Height)
		log.Printf("[DEBUG] clear stale blocks done")
	}
}

func (b *BlockFetcher) checkForkedChain() *model.Block {
	tail := blockService.GetTailBlock()
	if tail == nil {
		return tail
	}
	if tail.Height+1 < b.TriggerBlock.NHeight {
		// no forked chain, return
		// log.Printf("safe tail(#%d) trigger(#%d)", tail.Height, b.TriggerBlock.NHeight)
		return tail
	}

	if tail.Height-1 == b.TriggerBlock.NHeight && bytes.Compare(tail.Hash, b.TriggerBlock.HashPrev) == 0 {
		// log.Printf("safe tail(#%d) trigger(#%d), tail hash match trigger prevhash", tail.Height, b.TriggerBlock.NHeight)
		return tail
	}

	log.Printf("[INFO] forked block[%s](#%d) is detected", hex.EncodeToString(tail.Hash), tail.Height)
	// forked block is detected
	// find the initial forked block
	forkedBlock := b.findForkedBlock()
	if forkedBlock != nil {
		log.Printf("[INFO] found pre-forked block [%s](#%d)", hex.EncodeToString(forkedBlock.Hash), forkedBlock.NHeight)
		err := clearForkedChain(forkedBlock.NHeight)
		if err != nil {
			log.Printf("[ERROR] clear forked chain failed: [%s]", err)
		}
	}

	// retrieve new cleared tail
	tail = blockService.GetTailBlock()
	return tail
}

func (b *BlockFetcher) findForkedBlock() *lws.Block {
	prevHash := b.TriggerBlock.HashPrev

	for {
		blocks, err := b.fetch(prevHash, 1)
		if err != nil {
			// the situation couldn't happed
			log.Fatalf("[ERROR] reverse fetch failed [%s]", err)
		}

		block := blocks[0]
		isExist := checkBlockExistanceByHash(block.Hash)
		if isExist {
			// found
			return block
		}

		prevHash = block.HashPrev
	}
}

func (b *BlockFetcher) fetchAndHandleBlocks(hash []byte) error {
	blocks, err := b.fetch(hash, FETCH_NUMBER)
	if err != nil {
		log.Fatalf("[ERROR] block fetch err[%s]", err)
	}

	err = b.handle(blocks)
	if err != nil {
		log.Fatalf("[ERROR] block fetch handle err[%s]", err)
	}
	return err
}

func (b *BlockFetcher) fetch(hash []byte, num int32) ([]*lws.Block, error) {
	var err error
	var response interface{}
	cclient := cclientModule.GetPrimaryClient()
	hashStr := hex.EncodeToString(hash)

	params := &lws.GetBlocksArg{
		Hash:   hash,
		Number: num,
	}

	log.Printf("[INFO] fetch [%d] blocks start hash [%s]", params.Number, hashStr)

	serializedParams, err := ptypes.MarshalAny(params)
	if err != nil {
		log.Fatal("[ERROR] could not serialize any field")
	}

	method := &dbp.Method{
		Method: "getblocks",
		Params: serializedParams,
	}
	response, err = cclient.Call(method)
	for ; coreclient.IsClientTimeoutError(err); response, err = cclient.Call(method) {
		log.Printf("[INFO] fetch block [%s] timeout, retry.", hashStr)
	}

	if err != nil {
		log.Printf("[ERROR] fetch block err [%s]", err)
		return nil, err
	}

	result, ok := response.(*dbp.Result)
	if !ok {
		log.Printf("[ERROR] fetch block non-result response [%s]", response)
		return nil, err
	}

	if result.Error != "" {
		err = fmt.Errorf("[ERROR] fetch block result error [%s]", result.Error)
		return nil, err
	}

	blocksLen := len(result.Result)
	blocks := make([]*lws.Block, blocksLen)

	for idx, serializedAny := range result.Result {
		blocks[idx] = &lws.Block{}
		err = ptypes.UnmarshalAny(serializedAny, blocks[idx])
		if err != nil {
			log.Printf("[ERROR] unmashall result error [%s]", err)
			return nil, err
		}
	}

	return blocks, nil
}

func (b *BlockFetcher) handle(blocks []*lws.Block) error {
	log.Printf("[DEBUG] fetch list start")
	for _, block := range blocks {
		log.Printf("[DEBUG] fetch Block hash [%s] type[%d] (#%d)", hex.EncodeToString(block.Hash), block.NType, block.NHeight)
	}
	log.Printf("[DEBUG] fetch list done")

	for _, block := range blocks {
		err, _ := handleSyncBlock(block, true)
		if err != nil {
			log.Printf("[ERROR] handle sync block error [%s]", err)
		}
		if bytes.Compare(block.Hash, b.TriggerBlock.Hash) == 0 || block.NHeight > b.TriggerBlock.NHeight {
			b.isTriggerBlockSynced = true
		}
	}
	return nil
}

func checkBlockExistanceByHash(hash []byte) bool {
	count := 0
	connection := dbmodule.GetConnection()

	connection.Model(&model.Block{}).Where("hash = ?", hash).Count(&count)
	// if err occurs, consider it as non-exist
	return count == 1
}

func clearForkedChain(height uint32) error {
	connection := dbmodule.GetConnection()

	// recover used inputs
	err := utxoService.RecoverUsedInputs(height, connection)
	if err != nil {
		log.Printf("[ERROR] recover used inputs error")
		return err
	}

	// clear utxo
	res := connection.Unscoped().Where("block_height >= ? and block_height != ?", height, TXPOOL_HEIGHT).Delete(&model.Utxo{})
	if res.Error != nil {
		log.Printf("[ERROR] delete [%d] forked utxo", res.RowsAffected)
		return res.Error
	}

	// clear tx
	res = connection.Unscoped().Where("block_height >= ? and block_height != ?", height, TXPOOL_HEIGHT).Delete(&model.Tx{})
	if res.Error != nil {
		log.Printf("[ERROR] delete [%d] forked block", res.RowsAffected)
		return res.Error
	}

	// clear block
	res = connection.Unscoped().Where("height >= ?", height).Delete(&model.Block{})
	if res.Error != nil {
		log.Printf("[ERROR] delete [%d] forked block", res.RowsAffected)
		return res.Error
	}

	return nil
}
