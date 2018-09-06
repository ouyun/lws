package block

import (
	// "github.com/btcsuite/btcutil/base58"
	// "github.com/joho/godotenv"
	"github.com/lomocoin/lws/internal/constant"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	dbmodule "github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/test/testdb"

	"log"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	db := dbmodule.GetGormDb()
	log.Println("testing.M")
	log.Println("DATABASE_URL: ", os.Getenv("DATABASE_URL"))

	exedir, _ := os.Executable()
	log.Printf("exedir = %+v\n", exedir)
	cwd, _ := os.Getwd()
	log.Printf("cwd = %+v\n", cwd)

	exitCode := m.Run()

	db.Close()
	os.Exit(exitCode)
}

func TestIsBlockExisted(t *testing.T) {
	var (
		height uint
		// hash   []byte
	)

	height = 3
	hash := []byte{1, 2, 3, 4}

	isExisted := isBlockExisted(height, hash, true)
	log.Printf("isExisted = %+v\n", isExisted)
}

func TestGetTail(t *testing.T) {
	tail := getTailBlock()
	log.Printf("tail = %+v\n", tail)
}

func TestHandleSyncBlockGenesis(t *testing.T) {
	testdb.ResetDb()

	gormdb := dbmodule.GetGormDb()
	_, err := gormdb.DB().Exec("INSERT INTO block (created_at,updated_at,deleted_at,hash,version,`type`,prev,`timestamp`,merkle,height,mint_tx_id,sig) VALUES ('2018-09-05 10:35:15.000','2018-09-05 10:35:15.000',NULL,0x30303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303031,1,65535,NULL,'2018-09-05 10:35:15.000','',0,'',NULL) ;INSERT INTO block (created_at,updated_at,deleted_at,hash,version,`type`,prev,`timestamp`,merkle,height,mint_tx_id,sig) VALUES ('2018-09-05 10:51:34.000','2018-09-05 10:51:34.000',NULL,0x30303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303032,1,65280,0x30303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303031,'2018-09-05 10:51:33.000','',1,'',NULL) ;")
	if err != nil {
		t.Fatal("batch error", err)
	}

	genesis := &lws.Block{
		NVersion:   0x00000001,
		NType:      uint32(constant.BLOCK_TYPE_GENESIS),
		NTimeStamp: uint32(time.Now().Unix()),
		Height:     0,
		Hash:       []byte("0000000000000000000000000000000000000000000000000000000000000001"),
	}

	if ok := handleSyncBlock(genesis); !ok {
		t.Errorf("sync failed")
	}
}

func TestHandleSyncBlockOriginSuccess(t *testing.T) {
	block := &lws.Block{
		NVersion:   0x00000001,
		NType:      uint32(constant.BLOCK_TYPE_ORIGIN),
		NTimeStamp: uint32(time.Now().Unix()),
		Height:     1,
		Hash:       []byte("0000000000000000000000000000000000000000000000000000000000000002"),
		HashPrev:   []byte("0000000000000000000000000000000000000000000000000000000000000001"),
	}

	if ok := handleSyncBlock(block); !ok {
		t.Errorf("sync failed")
	}
}

func TestHandleSyncBlockExtendSuccess(t *testing.T) {
	block := &lws.Block{
		NVersion:   0x00000001,
		NType:      uint32(constant.BLOCK_TYPE_EXTENDED),
		NTimeStamp: uint32(time.Now().Unix()),
		Height:     2,
		Hash:       []byte("0000000000000000000000000000000000000000000000000000000000000003"),
		HashPrev:   []byte("0000000000000000000000000000000000000000000000000000000000000002"),
	}

	if ok := handleSyncBlock(block); !ok {
		t.Errorf("sync failed")
	}
}

func TestHandleSyncBlockExtendHeight(t *testing.T) {
	block := &lws.Block{
		NVersion:   0x00000001,
		NType:      uint32(constant.BLOCK_TYPE_EXTENDED),
		NTimeStamp: uint32(time.Now().Unix()),
		Height:     3,
		Hash:       []byte("0000000000000000000000000000000000000000000000000000000000000003"),
		HashPrev:   []byte("0000000000000000000000000000000000000000000000000000000000000002"),
	}

	if ok := handleSyncBlock(block); !ok {
		t.Errorf("sync failed")
	}
}

func TestHandleSyncBlockExtendExisted(t *testing.T) {
	block := &lws.Block{
		NVersion:   0x00000001,
		NType:      uint32(constant.BLOCK_TYPE_EXTENDED),
		NTimeStamp: uint32(time.Now().Unix()),
		Height:     2,
		Hash:       []byte("0000000000000000000000000000000000000000000000000000000000000003"),
		HashPrev:   []byte("0000000000000000000000000000000000000000000000000000000000000002"),
	}

	if ok := handleSyncBlock(block); !ok {
		t.Errorf("sync failed")
	}
}

func TestHandleSyncBlockExtendBreak(t *testing.T) {
	block := &lws.Block{
		NVersion:   0x00000001,
		NType:      uint32(constant.BLOCK_TYPE_EXTENDED),
		NTimeStamp: uint32(time.Now().Unix()),
		Height:     2,
		Hash:       []byte("0000000000000000000000000000000000000000000000000000000000000013"),
		HashPrev:   []byte("0000000000000000000000000000000000000000000000000000000000000009"),
	}

	if ok := handleSyncBlock(block); !ok {
		t.Errorf("sync failed")
	}
}
