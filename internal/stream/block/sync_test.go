package block

import (
	// "github.com/btcsuite/btcutil/base58"
	// "github.com/joho/godotenv"
	"github.com/FissionAndFusion/lws/internal/constant"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/db"
	blockService "github.com/FissionAndFusion/lws/internal/db/service/block"
	"github.com/FissionAndFusion/lws/test/helper"

	"bytes"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	connection := db.GetConnection()
	// connection.LogMode(true)

	exitCode := m.Run()

	connection.Close()
	os.Exit(exitCode)
}

func TestIsBlockExistedTrue(t *testing.T) {
	TestHandleSyncBlockGenesis(t)
	hash := []byte("0000000000000000000000000001")

	isExisted := isBlockExisted(0, hash, false)
	if !isExisted {
		t.Error("expect existed, but non-existed")
	}
}

func TestIsBlockExistedFalse(t *testing.T) {
	helper.ResetDb()
	hash := []byte("0000000000000000000000000001")

	isExisted := isBlockExisted(0, hash, false)
	if isExisted {
		t.Error("expect existed, but non-existed")
	}
}

func TestGetTail(t *testing.T) {
	TestHandleSyncBlockExtendSuccess(t)
	tail := blockService.GetTailBlock()
	if tail == nil {
		t.Errorf("expect tail, but [%v]", tail)
		return
	}

	hash := []byte("0000000000000000000000000003")
	height := uint32(2)

	if bytes.Compare(hash, tail.Hash) != 0 || height != tail.Height {
		t.Errorf("expect hash[%s](%d), but [%s](%d)", hash, height, tail.Hash, tail.Height)
	}
}

func TestHandleSyncBlockGenesis(t *testing.T) {
	helper.ResetDb()

	block := &lws.Block{
		NVersion:   0x0000001,
		NType:      uint32(constant.BLOCK_TYPE_GENESIS),
		NTimeStamp: uint32(time.Now().Unix()),
		NHeight:    0,
		Hash:       []byte("0000000000000000000000000001"),
		TxMint: &lws.Transaction{
			NVersion: uint32(1),
			Hash:     []byte("00000000000000000000000000001"),
			NAmount:  100,
			NTxFee:   0,
			CDestination: &lws.Transaction_CDestination{
				Prefix: uint32(1),
				Data:   []byte("ffffff78901234567890123456789013"),
			},
		},
	}

	if err, skip := handleSyncBlock(block, false); err != nil || !skip {
		t.Errorf("the block should be written to chain")
	}
}

func TestHandleSyncBlockOriginSuccess(t *testing.T) {
	TestHandleSyncBlockGenesis(t)
	block := &lws.Block{
		NVersion:   0x0000001,
		NType:      uint32(constant.BLOCK_TYPE_ORIGIN),
		NTimeStamp: uint32(time.Now().Unix()),
		NHeight:    1,
		Hash:       []byte("0000000000000000000000000002"),
		HashPrev:   []byte("0000000000000000000000000001"),
		TxMint: &lws.Transaction{
			NVersion: uint32(1),
			Hash:     []byte("00000000000000000000000000002"),
			NAmount:  100,
			NTxFee:   0,
			CDestination: &lws.Transaction_CDestination{
				Prefix: uint32(1),
				Data:   []byte("ffffff78901234567890123456789013"),
			},
		},
	}

	if err, skip := handleSyncBlock(block, false); err != nil || !skip {
		t.Errorf("the block should be written to chain")
	}
}

func TestHandleSyncBlockExtendSuccess(t *testing.T) {
	TestHandleSyncBlockOriginSuccess(t)
	block := &lws.Block{
		NVersion:   0x0000001,
		NType:      uint32(constant.BLOCK_TYPE_SUBSIDIARY),
		NTimeStamp: uint32(time.Now().Unix()),
		NHeight:    2,
		Hash:       []byte("0000000000000000000000000003"),
		HashPrev:   []byte("0000000000000000000000000002"),
		TxMint: &lws.Transaction{
			NVersion: uint32(1),
			Hash:     []byte("00000000000000000000000000003"),
			NAmount:  100,
			NTxFee:   0,
			CDestination: &lws.Transaction_CDestination{
				Prefix: uint32(1),
				Data:   []byte("ffffff78901234567890123456789013"),
			},
		},
	}

	if err, skip := handleSyncBlock(block, false); err != nil || !skip {
		t.Errorf("the block should be written to chain")
	}
}

func TestHandleSyncBlockExtendSuccessWithTx(t *testing.T) {
	TestHandleSyncBlockOriginSuccess(t)
	block := &lws.Block{
		NVersion:   0x0000001,
		NType:      uint32(constant.BLOCK_TYPE_SUBSIDIARY),
		NTimeStamp: uint32(time.Now().Unix()),
		NHeight:    2,
		Hash:       []byte("0000000000000000000000000003"),
		HashPrev:   []byte("0000000000000000000000000002"),
		Vtx: []*lws.Transaction{
			&lws.Transaction{
				NVersion: uint32(1),
				Hash:     []byte("12345678901234567890123456789012"),
				NAmount:  100,
				NTxFee:   11,
				CDestination: &lws.Transaction_CDestination{
					Prefix: uint32(1),
					Data:   []byte("ffffff78901234567890123456789013"),
				},
			},
			&lws.Transaction{
				NVersion: uint32(1),
				Hash:     []byte("12345678901234567890123456789013"),
				NAmount:  200,
				NTxFee:   10,
				CDestination: &lws.Transaction_CDestination{
					Prefix: uint32(1),
					Data:   []byte("ffffff78901234567890123456789013"),
				},
			},
		},
		TxMint: &lws.Transaction{
			NVersion: uint32(1),
			Hash:     []byte("00000000000000000000000000003"),
			NAmount:  100,
			NTxFee:   0,
			CDestination: &lws.Transaction_CDestination{
				Prefix: uint32(1),
				Data:   []byte("ffffff78901234567890123456789013"),
			},
		},
	}

	if err, skip := handleSyncBlock(block, false); err != nil || !skip {
		t.Errorf("the block should be written to chain")
	}
}

func TestHandleSyncBlockExtendWrongHeightRecovery(t *testing.T) {
	helper.ResetDb()
	helper.LoadTestSeed("seedBasic.sql")

	block := &lws.Block{
		NVersion:   0x0000001,
		NType:      uint32(constant.BLOCK_TYPE_SUBSIDIARY),
		NTimeStamp: uint32(time.Now().Unix()),
		NHeight:    3,
		Hash:       []byte("0000000000000000000000000006"),
		HashPrev:   []byte("0000000000000000000000000004"),
		TxMint: &lws.Transaction{
			NVersion: uint32(1),
			Hash:     []byte("00000000000000000000000000006"),
			NAmount:  100,
			NTxFee:   0,
			CDestination: &lws.Transaction_CDestination{
				Prefix: uint32(1),
				Data:   []byte("ffffff78901234567890123456789013"),
			},
		},
	}

	err, skip := handleSyncBlock(block, false)
	if err == nil {
		t.Errorf("should got recovery error but nil")
	}

	if err.Error() != "trigger recovery" {
		t.Errorf("should got recovery error but [%s]", err)
	}

	if !skip {
		t.Errorf("should skip the trigger-block")
	}

	tail := blockService.GetTailBlock()
	if bytes.Compare(tail.Hash, block.Hash) == 0 {
		t.Errorf("should not inserted wrong height block")
	}
}
