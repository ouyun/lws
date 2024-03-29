package mqtt

import (
	"encoding/hex"
	// "log"
	"os"
	"testing"
	"time"
	// "github.com/gomodule/redigo/redis"
	// "github.com/FissionAndFusion/lws/internal/db"
	// "github.com/FissionAndFusion/lws/internal/db/model"
	// "github.com/FissionAndFusion/lws/internal/gateway/crypto"
)

func TestSyncReq(t *testing.T) {
	lws := &Program{
		Id:    "lws syncReq",
		IsLws: true,
	}
	lws.Init()
	if err := lws.Start(); err != nil {
		t.Errorf("client start failed")
	}
	cli := &Program{
		Id:    "clissss",
		IsLws: false,
	}
	cli.Init()
	if err := cli.Start(); err != nil {
		t.Errorf("client start failed")
	}
	cli.Subscribe("wqweqwasasqw/fnfn/SyncReply", 1, servicReplyHandler)
	forkId, _ := hex.DecodeString(os.Getenv("FORK_ID"))
	syncPayload := SyncPayload{ //Sync
		Nonce:     uint16(1231),
		AddressId: uint32(4),
		ForkID:    forkId,
		UTXOHash:  []byte(RandStringBytesRmndr(32)),
		Signature: []byte(RandStringBytesRmndr(20)),
	}
	syncMsg, err := StructToBytes(syncPayload)
	if err != nil {
		t.Errorf("client publish fail")
	}
	err = cli.Publish("LWS/lws/SyncReq", 1, false, syncMsg)
	if err != nil {
		t.Errorf(" client publish fail")
	}
	time.Sleep(1 * time.Second)
	cli.Stop()
}
