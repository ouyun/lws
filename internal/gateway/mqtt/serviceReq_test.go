package mqtt

import (
	// "bytes"
	cryptoH "crypto"
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	// "strconv"
	// "strings"
	"testing"
	"time"

	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/gateway/crypto"
	// "github.com/gomodule/redigo/redis"
	edwards25519 "golang.org/x/crypto/ed25519"
)

// func TestSENDDD(t *testing.T) {
// 	// cli := &Program{
// 	// 	Id:    "clilassssss",
// 	// 	IsLws: false,
// 	// }
// 	// cli.Init()
// 	// if err := cli.Start(); err != nil {
// 	// 	t.Errorf("client start failed")
// 	// }
// 	code1 := "84 69 83 84 45 64"
// 	codeArr1 := strings.Split(code1, " ")
// 	log.Printf("len : %+v", len(codeArr1))
// 	codeAdd1 := make([]byte, 6)
// 	log.Printf("len : %+v", codeArr1)
// 	for index := 0; index < len(codeArr1); index++ {
// 		value, _ := strconv.Atoi(codeArr1[index])
// 		codeAdd1[index] = byte(value)
// 	}
// 	// stss, _ := hex.DecodeString("544553542d4344")
// 	log.Printf("stss : %+v", string([]byte("TEST-CD")))
// 	log.Printf("code : %+v", codeAdd1)
// 	// err := cli.Publish("TEST01/fnfn/ServiceReply", 0, false, codeAdd1)
// 	// if err != nil {
// 	// 	log.Printf("err : %+v", err)
// 	// }
// }

// func TestAPISS(t *testing.T) {

// 	code1 := "233 249 230 49 102 97 68 115 223 140 189 218 85 57 40 239 114 182 160 237 75 50 138 97 171 14 40 114 0 252 172 102"
// 	codeArr1 := strings.Split(code1, " ")
// 	log.Printf("len : %+v", len(codeArr1))
// 	codeAdd1 := make([]byte, 32)
// 	log.Printf("len : %+v", codeArr1)
// 	for index := 0; index < len(codeArr1); index++ {
// 		value, _ := strconv.Atoi(codeArr1[index])
// 		codeAdd1[index] = byte(value)
// 	}

// 	code := "0 126 44 55 193 57 193 90 149 130 22 46 166 14 119 117 30 33 10 196 12 255 221 14 180 133 55 97 135 13 8 89"
// 	codeArr := strings.Split(code, " ")
// 	log.Printf("len : %+v", len(codeArr))
// 	codeAdd := make([]byte, 32)
// 	log.Printf("len : %+v", codeArr)
// 	for index := 0; index < len(codeArr); index++ {
// 		value, _ := strconv.Atoi(codeArr[index])
// 		codeAdd[index] = byte(value)
// 	}
// 	log.Printf("len : %+v", codeAdd)
// 	// pubKey, privKey, _ := crypto.GenerateKeyPair(nil)
// 	var address crypto.PrivateKey
// 	copy(address[:], codeAdd[:])
// 	var addressPub crypto.PublicKey
// 	copy(addressPub[:], codeAdd1[:])
// 	apiKey := crypto.GenerateApiKey(&address, &addressPub)
// 	log.Printf("apiKey : %+v", apiKey)
// }

func TestServiceReq(t *testing.T) {
	// lws := &Program{
	// 	Id:    "lws serviceReq",
	// 	IsLws: true,
	// }
	// lws.Init()
	// if err := lws.Start(); err != nil {
	// 	t.Errorf("client start failed")
	// }
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
	// c := make(chan int, 1)

	// code := "1 0 1 21 31 223 54 184 6 54 94 213 182 216 67 173 88 68 174 66 80 132 46 27 115 69 27 63 250 22 209 215 234 10 178 101 0 0 0 225 239 173 91 1 111 147 124 47 89 68 245 218 42 17 140 235 176 103 205 44 156 146 199 89 85 206 5 170 5 21 138 26 242 142 22 7 2 0 84 69 83 84 0 64 0 176 18 160 197 218 210 26 249 201 156 79 194 145 80 135 123 184 142 224 237 10 97 130 94 194 36 56 122 242 145 35 231 106 193 159 2 250 31 22 235 112 216 114 76 31 56 197 76 119 57 212 139 92 40 188 146 12 34 247 239 197 242 213 7"
	// codeArr := strings.Split(code, " ")
	// codeAddd := make([]byte, 149)
	// log.Printf("len : %+v", len(codeArr))
	// for index := 0; index < len(codeArr); index++ {
	// 	value, _ := strconv.Atoi(codeArr[index])
	// 	codeAddd[index] = byte(value)
	// }
	// token := cli.Client.Publish("LWS/lws/ServiceReq", 0, false, codeAddd)

	// for {
	// 	if token.Wait() {
	// 		c <- 1
	// 		break
	// 	}
	// }
	// <-c
	err = cli.Publish("LWS/lws/ServiceReq", 0, false, servicMsg)
	if err != nil {
		t.Errorf("client publish failed")
	}
	time.Sleep(14 * time.Second)
	cli.Stop()
}

// func TestSignCase(t *testing.T) {
// 	code := "1 0 1 21 31 223 54 184 6 54 94 213 182 216 67 173 88 68 174 66 80 132 46 27 115 69 27 63 250 22 209 215 234 10 178 101 0 0 0 225 239 173 91 1 111 147 124 47 89 68 245 218 42 17 140 235 176 103 205 44 156 146 199 89 85 206 5 170 5 21 138 26 242 142 22 7 2 0 84 69 83 84 0 62 0 175 0 182 224 100 110 90 225 12 50 185 111 187 154 240 196 70 75 74 145 218 86 188 179 168 35 47 155 206 96 156 73 243 171 20 81 148 139 9 4 53 83 42 239 214 112 134 221 246 166 243 98 249 180 33 21 74 130 159 149 81 94"
// 	codeArr := strings.Split(code, " ")
// 	log.Printf("len : %+v", len(codeArr))
// 	codeAdd := make([]byte, 147)
// 	log.Printf("len : %+v", codeArr)
// 	for index := 0; index < len(codeArr); index++ {
// 		value, _ := strconv.Atoi(codeArr[index])
// 		codeAdd[index] = byte(value)
// 	}
// 	s := ServicePayload{}

// 	err := DecodePayload(codeAdd, &s)
// 	if err != nil {
// 		log.Printf("err: %+v", err)
// 	}
// 	connection := db.GetConnection()
// 	user := model.User{}

// 	connection.LogMode(true)
// 	found := connection.Where("address_id = ?", 2).Find(&user).RecordNotFound()
// 	if !found {
// 		log.Printf("user: %+v", user)
// 	}
// 	log.Printf("user: %s", hex.EncodeToString(user.Address))
// 	var UTXOs []UTXO
// 	var utxol []model.Utxo

// 	// connection.Raw("SELECT * FROM utxo WHERE ID = ?", 1).Scan(&utxol)
// 	log.Printf("utxol: %+v", utxol)

// 	connection.Raw("SELECT "+
// 		"utxo.tx_hash AS tx_id, "+
// 		"utxo.out, "+
// 		"utxo.block_height, "+
// 		"tx.tx_type, "+
// 		"utxo.amount, "+
// 		"tx.sender AS sender, "+
// 		"tx.lock_until, "+
// 		"tx.data "+
// 		"FROM utxo "+
// 		"INNER JOIN tx "+
// 		"ON utxo.tx_hash = tx.hash "+
// 		"AND utxo.destination = ? "+
// 		"ORDER BY utxo.amount ASC, utxo.out ASC ", user.Address).Find(&UTXOs)
// 	// Scan(&UTXOs)
// 	log.Printf("UTXOs: %+v", UTXOs)

// }

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
	// log.Printf("user : %+v \n", user)
	// cliMap := CliMap{}
	// value, err := redis.Values(redisConn.Do("hgetall", 9))
}
