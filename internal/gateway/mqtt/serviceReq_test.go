package mqtt

import (
	// "bytes"
	"encoding/hex"
	"log"
	"os"
	"testing"
	"time"

	// "github.com/gomodule/redigo/redis"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/gateway/crypto"
	edwards25519 "golang.org/x/crypto/ed25519"
)

func TestServiceReq(t *testing.T) {
	lws := &Program{
		Id:    "lws serviceReq",
		IsLws: true,
	}
	lws.Init()
	if err := lws.Start(); err != nil {
		t.Errorf("client start failed")
	}
	cli := &Program{
		Id:    "cli",
		IsLws: false,
	}
	cli.Init()
	if err := cli.Start(); err != nil {
		t.Errorf("client start failed")
	}

	cli.Subscribe("wqweqwasasqw/fnfn/ServiceReply", 0, servicReplyHandler)
	addr, _, signKey := crypto.GenerateKeyPair(nil)
	address := make([]byte, 1)
	address[0] = uint8(1)
	address = append(address, addr[:]...)
	// address, _ := hex.DecodeString("6f937c2f5944f5da2a118cebb067cd2c9c92c75955ce05aa05158a1af28e1607")
	// hex.EncodeToString
	// log.Printf("ServiceReply: %+v\n", hex.EncodeToString(address[:]))
	topicPrefix := "wqweqwasasqw" + string(byte(0x00))
	forkList, _ := hex.DecodeString(os.Getenv("FORK_ID"))
	forkList = append(forkList, []byte(RandStringBytesRmndr(32*8))...)
	servicePayload := ServicePayload{ //serviceRequ
		Nonce:         uint16(1231),
		Address:       address[:],
		Version:       uint32(5363),
		TimeStamp:     uint32(time.Now().Unix()),
		ForkNum:       uint8(9),
		ForkList:      forkList,
		ReplyUTXON:    uint16(10),
		TopicPrefix:   topicPrefix,
		SignBytes:     uint16(64),
		ServSignature: []byte(RandStringBytesRmndr(64)),
	}
	servicMsgWithoutSign, err := StructToBytes(servicePayload)
	GetSignature(&servicePayload, signKey[:], servicMsgWithoutSign)
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

func GetSignature(s *ServicePayload, signKey []byte, payload []byte) {

	s.ServSignature = edwards25519.Sign(edwards25519.PrivateKey(signKey), payload[:(len(payload)-66)])
}
func TestUser(t *testing.T) {
	connection := db.GetConnection()
	if connection == nil {
		log.Println("conn db fail ")
	}
	user := model.User{}
	pool := GetRedisPool()
	redisConn := pool.Get()
	defer connection.Close()
	defer redisConn.Close()
	connection.Where("address_id = ?", 2).First(&user).RecordNotFound()
	// log.Printf("user : %+v \n", user)
	// cliMap := CliMap{}
	// value, err := redis.Values(redisConn.Do("hgetall", 9))
}
