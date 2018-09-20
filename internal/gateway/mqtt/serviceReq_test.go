package mqtt

import (
	"encoding/hex"
	// "log"
	"os"
	"testing"
	"time"

	// "github.com/gomodule/redigo/redis"
	// "github.com/lomocoin/lws/internal/db"
	// "github.com/lomocoin/lws/internal/db/model"
	"github.com/lomocoin/lws/internal/gateway/crypto"
)

func TestServiceReq(t *testing.T) {
	cli := &Program{
		Id:    "cli",
		isLws: false,
	}
	cli.Init()
	if err := cli.Start(); err != nil {
		t.Errorf("client start failed")
	}

	cli.Subscribe("wqweqwasasqw/fnfn/ServiceReply", 0, servicReplyHandler)
	address, _, _ := crypto.GenerateKeyPair(nil)

	// address, _ := hex.DecodeString("6f937c2f5944f5da2a118cebb067cd2c9c92c75955ce05aa05158a1af28e1607")
	// hex.EncodeToString
	// log.Printf("ServiceReply: %+v\n", hex.EncodeToString(address[:]))
	topicPrefix := "wqweqwasasqw" + string(byte(0x00))
	forkList, _ := hex.DecodeString(os.Getenv("Fork_Id"))
	forkList = append(forkList, []byte(RandStringBytesRmndr(32*8))...)
	servicePayload := ServicePayload{ //serviceRequ
		Nonce:       uint16(1231),
		Address0:    uint8(1),
		Address:     address[:],
		Version:     uint32(5363),
		TimeStamp:   uint32(time.Now().Unix()),
		ForkNum:     uint8(9),
		ForkList:    forkList,
		ReplyUTXON:  uint16(10),
		TopicPrefix: topicPrefix,
		Signature:   RandStringBytesRmndr(64),
	}
	servicMsg, err := StructToBytes(servicePayload)
	if err != nil {
		t.Errorf("client publish failed")
	}
	err = cli.Publish("LWS/lws/ServiceReq", 0, false, servicMsg)
	if err != nil {
		t.Errorf("client publish failed")
	}
	time.Sleep(4 * time.Second)
	cli.Stop()
}

// func TestUser(t *testing.T) {
// 	connection := db.GetConnection()
// 	if connection == nil {
// 		log.Println("conn db fail ")
// 	}
// 	user := model.User{}
// 	pool := NewRedisPool()
// 	redisConn := pool.Get()
// 	defer connection.Close()
// 	defer redisConn.Close()
// 	connection.Where("address_id = ?", 2).First(&user).RecordNotFound()
// 	log.Printf("user : %+v \n", user)
// 	// cliMap := CliMap{}
// 	// value, err := redis.Values(redisConn.Do("hgetall", 9))
// }
