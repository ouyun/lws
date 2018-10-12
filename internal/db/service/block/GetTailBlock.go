package block

import (
	"encoding/hex"
	"log"

	"github.com/FissionAndFusion/lws/internal/constant"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
)

func GetTailBlock() *model.Block {
	block := &model.Block{}
	connection := db.GetConnection()
	res := connection.
		Where("block_type != ?", constant.BLOCK_TYPE_SUBSIDIARY).
		Order("height desc").
		Take(block)
	if res.Error != nil {
		log.Println("GetTailBlock failed", res.Error)
		return nil
	}
	hashStr := hex.EncodeToString(block.Hash)
	log.Printf("Tail: [%s](%d) type[%d]", hashStr, block.Height, block.BlockType)
	return block
}
