package block

import (
	"context"
	"log"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

func clearStaleBlocksInQueue(height uint32) {
	return
}

func createClearHandler(cancel context.CancelFunc, height uint32) func(body []byte) bool {
	return func(body []byte) bool {
		var err error

		added := &dbp.Added{}
		if err = proto.Unmarshal(body, added); err != nil {
			log.Println("[ERROR] unkonwn message received", body, err)
			return true
		}

		block := &lws.Block{}
		err = ptypes.UnmarshalAny(added.Object, block)
		if err != nil {
			log.Println("[ERROR] unpack Object failed", err)
			return true
		}

		if block.NHeight >= height {
			log.Printf("[INFO] detect block height [#%d], clearing done", block.NHeight)
			cancel()
			return false
		}

		log.Printf("[INFO] delete block[#%d] from consumer queue", block.NHeight)
		return true
	}
}
