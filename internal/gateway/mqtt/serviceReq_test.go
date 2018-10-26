package mqtt

import (
	// "bytes"
	cryptoH "crypto"
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/gateway/crypto"
	// "github.com/gomodule/redigo/redis"
	edwards25519 "golang.org/x/crypto/ed25519"
)

func TestAPISS(t *testing.T) {

	code1 := "178 10 234 215 209 22 250 63 27 69 115 27 46 132 80 66 174 68 88 173 67 216 182 213 94 54 6 184 54 223 31 21 1 1 112 89 84 86 194 230 223 96 49 238 244 249 254 205 74 76 41 115 115 161 21 67 165 27 3 181 77 78 31 203 17 34 0 112 89 84 86 194 230 223 96 49 238 244 249 254 205 74 76 41 115 115 161 21 67 165 27 3 181 77 78 31 203 17 34"
	codeArr1 := strings.Split(code1, " ")
	log.Printf("len : %+v", len(codeArr1))
	codeAdd1 := make([]byte, 99)
	log.Printf("len : %+v", codeArr1)
	for index := 0; index < len(codeArr1); index++ {
		value, _ := strconv.Atoi(codeArr1[index])
		codeAdd1[index] = byte(value)
	}
	log.Printf("codeAdd1 : %+v", hex.EncodeToString(codeAdd1))
}

func TestServiceReq(t *testing.T) {
	cli := &Program{
		Id:    "clilassss",
		IsLws: false,
	}
	cli.Init()
	if err := cli.Start(); err != nil {
		t.Errorf("client start failed")
	}

	// cli.Subscribe("wqweqwasasqw/fnfn/ServiceReply", 0, servicReplyHandler)
	addr, _, signKey := crypto.GenerateKeyPair(nil)
	// addr, _ := hex.DecodeString("6f937c2f5944f5da2a118cebb067cd2c9c92c75955ce05aa05158a1af28e1607")
	// hex.EncodeToString
	address := make([]byte, 1)
	address[0] = uint8(1)
	address = append(address, addr[:]...)
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
	cli.Stop()
}

func GetSignature(s *ServicePayload, signKey []byte, payload []byte) {
	s.ServSignature = edwards25519.Sign(edwards25519.PrivateKey(signKey), payload[:(len(payload)-66)])
}

func TestPrivKSign(t *testing.T) {

	addr, _, signKey := crypto.GenerateKeyPair(nil)
	address := make([]byte, 1)
	address[0] = uint8(1)
	address = append(address, addr[:]...)
	// address, _ := hex.DecodeString("6f937c2f5944f5da2a118cebb067cd2c9c92c75955ce05aa05158a1af28e1607")
	topicPrefix := "wqweqwasasqw" + string(byte(0x00))
	forkList, _ := hex.DecodeString(os.Getenv("FORK_ID"))
	forkList = append(forkList, []byte(RandStringBytesRmndr(32*1))...)
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

	privteKey := edwards25519.PrivateKey(signKey[:])
	signN, err := privteKey.Sign(rand.Reader, servicMsgWithoutSign[:(len(servicMsgWithoutSign)-66)], cryptoH.Hash(0))
	if err != nil {
		t.Errorf("client publish failed")
	}
	log.Printf("signN: %+v \n", signN)
	servicePayload.ServSignature = signN
	log.Printf("servicePayload      : %+v \n", servicePayload)
	servicMsg, err := StructToBytes(servicePayload)
	log.Print(edwards25519.Verify(servicePayload.Address[1:33], servicMsg[:(len(servicMsg)-66)], servicePayload.ServSignature))

	GetSignature(&servicePayload, signKey[:], servicMsgWithoutSign)
	if err != nil {
		t.Errorf("client publish failed")
	}
	servicMsg, err = StructToBytes(servicePayload)
	if err != nil {
		t.Errorf("client publish failed")
	}
	log.Printf("after servicePayload: %+v \n", servicePayload)
	log.Print(edwards25519.Verify(servicePayload.Address[1:33], servicMsg[:(len(servicMsg)-66)], servicePayload.ServSignature))

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
	// TODO: add user case,
}
