package block

import (
	"encoding/hex"
	"log"

	// "github.com/FissionAndFusion/lws/internal/constant"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
)

var tail *model.Block

func GetTailBlock() *model.Block {
	if tail != nil {
		log.Printf("Tail cache: [%s](%d) type[%d]", hex.EncodeToString(tail.Hash), tail.Height, tail.BlockType)
		return tail
	}

	return GetAndUpdateTailBlock()
}

func GetAndUpdateTailBlock() *model.Block {
	block := GetTailBlockFromDb()
	SetTailBlock(block)
	return block
}

func GetTailBlockFromDb() *model.Block {
	block := &model.Block{}
	connection := db.GetConnection()
	res := connection.
		// Where("block_type != ?", constant.BLOCK_TYPE_EXTENDED).
		Order("height desc, tstamp desc").
		Take(block)
	if res.Error != nil {
		log.Println("GetTailBlock failed", res.Error)
		return nil
	}
	hashStr := hex.EncodeToString(block.Hash)
	log.Printf("Tail: [%s](%d) type[%d]", hashStr, block.Height, block.BlockType)
	return block
}

func SetTailBlock(block *model.Block) {
	tail = block
}
