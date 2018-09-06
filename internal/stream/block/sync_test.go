package block

import (
	// "github.com/btcsuite/btcutil/base58"
	// "github.com/joho/godotenv"
	"github.com/lomocoin/lws/internal/constant"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	dbmodule "github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/testhelper"

	"bytes"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	db := dbmodule.GetGormDb()

	exitCode := m.Run()

	db.Close()
	os.Exit(exitCode)
}

func TestIsBlockExistedTrue(t *testing.T) {
	TestHandleSyncBlockGenesis(t)
	hash := []byte("0000000000000000000000000000000000000000000000000000000000000001")

	isExisted := isBlockExisted(0, hash, false)
	if !isExisted {
		t.Error("expect existed, but non-existed")
	}
}

func TestIsBlockExistedFalse(t *testing.T) {
	testhelper.ResetDb()
	hash := []byte("0000000000000000000000000000000000000000000000000000000000000001")

	isExisted := isBlockExisted(0, hash, false)
	if isExisted {
		t.Error("expect existed, but non-existed")
	}
}

func TestGetTail(t *testing.T) {
	TestHandleSyncBlockExtendSuccess(t)
	tail := GetTailBlock()
	if tail == nil {
		t.Errorf("expect tail, but [%v]", tail)
		return
	}

	hash := []byte("0000000000000000000000000000000000000000000000000000000000000003")
	height := uint(2)

	if bytes.Compare(hash, tail.Hash) != 0 || height != tail.Height {
		t.Errorf("expect hash[%s](%d), but [%s](%d)", hash, height, tail.Hash, tail.Height)
	}
}

func TestHandleSyncBlockGenesis(t *testing.T) {
	testhelper.ResetDb()

	block := &lws.Block{
		NVersion:   0x00000001,
		NType:      uint32(constant.BLOCK_TYPE_GENESIS),
		NTimeStamp: uint32(time.Now().Unix()),
		Height:     0,
		Hash:       []byte("0000000000000000000000000000000000000000000000000000000000000001"),
	}

	if err, skip := handleSyncBlock(block); err != nil || skip {
		t.Errorf("the block should be written to chain")
	}
}

func TestHandleSyncBlockOriginSuccess(t *testing.T) {
	TestHandleSyncBlockGenesis(t)
	block := &lws.Block{
		NVersion:   0x00000001,
		NType:      uint32(constant.BLOCK_TYPE_ORIGIN),
		NTimeStamp: uint32(time.Now().Unix()),
		Height:     1,
		Hash:       []byte("0000000000000000000000000000000000000000000000000000000000000002"),
		HashPrev:   []byte("0000000000000000000000000000000000000000000000000000000000000001"),
	}

	if err, skip := handleSyncBlock(block); err != nil || skip {
		t.Errorf("the block should be written to chain")
	}
}

func TestHandleSyncBlockExtendSuccess(t *testing.T) {
	TestHandleSyncBlockOriginSuccess(t)
	block := &lws.Block{
		NVersion:   0x00000001,
		NType:      uint32(constant.BLOCK_TYPE_EXTENDED),
		NTimeStamp: uint32(time.Now().Unix()),
		Height:     2,
		Hash:       []byte("0000000000000000000000000000000000000000000000000000000000000003"),
		HashPrev:   []byte("0000000000000000000000000000000000000000000000000000000000000002"),
	}

	if err, skip := handleSyncBlock(block); err != nil || skip {
		t.Errorf("the block should be written to chain")
	}
}

func TestHandleSyncBlockExtendWrongHeightRecovery(t *testing.T) {
	testhelper.ResetDb()
	testhelper.LoadTestSeed("seedBasic.sql")

	block := &lws.Block{
		NVersion:   0x00000001,
		NType:      uint32(constant.BLOCK_TYPE_EXTENDED),
		NTimeStamp: uint32(time.Now().Unix()),
		Height:     3,
		Hash:       []byte("0000000000000000000000000000000000000000000000000000000000000006"),
		HashPrev:   []byte("0000000000000000000000000000000000000000000000000000000000000004"),
	}

	err, skip := handleSyncBlock(block)
	if err == nil {
		t.Errorf("should got recovery error but nil")
	}

	if err.Error() != "trigger recovery" {
		t.Errorf("should got recovery error but [%s]", err)
	}

	if skip {
		t.Errorf("should not skip the trigger-block")
	}

	tail := GetTailBlock()
	if bytes.Compare(tail.Hash, block.Hash) == 0 {
		t.Errorf("should not inserted wrong height block")
	}
}
