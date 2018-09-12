package block

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/golang/protobuf/ptypes"
	"github.com/lomocoin/lws/internal/coreclient"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"

	// "github.com/lomocoin/lws/internal/db/model"
	cclientModule "github.com/lomocoin/lws/internal/stream/cclient"
)

// type FetchRangeCalculator struct {
// 	TriggerBlock *lws.Block
// 	tailBlock    *model.Block
// }

const (
	FETCH_NUMBER = 1000
)

type BlockFetcher struct {
	FetchNumber int
}

func FetchBlocks(triggerBlock *lws.Block) {
	b := &BlockFetcher{
		FetchNumber: FETCH_NUMBER,
	}

	b.startFetchBlocks(triggerBlock)
}

func (b *BlockFetcher) startFetchBlocks(triggerBlock *lws.Block) {
	// cclient := stream.GetPrimaryClient()
	log.Println("start Fetch blocks")

	// 1. if the fork chain is detected, remove the bad-chain data

	for tail := GetTailBlock(); tail == nil || bytes.Compare(tail.Hash, triggerBlock.Hash) != 0; tail = GetTailBlock() {
		var hash []byte
		if tail != nil {
			hash = tail.Hash
		}

		b.fetchAndHandleBlocks(hash)
	}
}

func (b *BlockFetcher) fetchAndHandleBlocks(hash []byte) error {
	blocks, err := b.fetch(hash)
	if err != nil {
		log.Fatalf("block fetch err[%s]", err)
	}

	err = b.handle(blocks)
	if err != nil {
		log.Fatalf("block fetch handle err[%s]", err)
	}
	return err
}

func (b *BlockFetcher) fetch(hash []byte) ([]*lws.Block, error) {
	var err error
	var response interface{}
	cclient := cclientModule.GetPrimaryClient()
	hashStr := hex.EncodeToString(hash)

	params := &lws.GetBlocksArg{
		Hash:   hashStr,
		Number: FETCH_NUMBER,
	}

	serializedParams, err := ptypes.MarshalAny(params)
	if err != nil {
		log.Fatal("could not serialize any field")
	}

	method := &dbp.Method{
		Method: "getblocks",
		Params: serializedParams,
	}

	for response, err = cclient.Call(method); isClientTimeoutError(err); {
		log.Printf("fetch block [%s] timeout, retry", hashStr)
		return nil, err
	}

	if err != nil {
		log.Printf("fetch block err [%s]", err)
		return nil, err
	}

	result, ok := response.(*dbp.Result)
	if !ok {
		log.Printf("fetch block non-result response [%s]", response)
		return nil, err
	}

	if result.Error != "" {
		log.Printf("fetch block result error [%s]", result.Error)
		return nil, err
	}

	blocksLen := len(result.Result)
	blocks := make([]*lws.Block, blocksLen)

	for idx, serializedAny := range result.Result {
		blocks[idx] = &lws.Block{}
		err = ptypes.UnmarshalAny(serializedAny, blocks[idx])
		if err != nil {
			log.Printf("unmashall result error [%s]", err)
			return nil, err
		}
	}

	return blocks, nil
}

func (b *BlockFetcher) handle(blocks []*lws.Block) error {
	for _, block := range blocks {
		err, _ := handleSyncBlock(block, nil)
		if err != nil {
			log.Printf("handle sync block error [%s]", err)
		}
	}
	return nil
}

func isClientTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	cliErr, ok := err.(*coreclient.ClientError)

	if !ok {
		return false
	}

	if !cliErr.Timeout {
		return false
	}

	return true
}
