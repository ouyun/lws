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

func TestSENDDD(t *testing.T) {

	pubKey01, _ := hex.DecodeString(reverseString("0cac100f5a539891d5ff6a1dca61bb3e8cacbb05dcb8490931f5da5738d65ab7"))
	log.Printf("pubKey01: %+v \n", hex.EncodeToString(pubKey01))
	// log.Printf("pubKey01: %+v \n", "dff7f7eaf8f5d19a08ffe596817e8c537196402e2370ec7cff8c7a3766893efb")
	// pubKey01 = reverseBytes(pubKey01)
	// log.Printf("reverseBytes pubKey01: %+v \n", pubKey01)
	privKey01, _ := hex.DecodeString(reverseString("05eedc532a8d68df9f2d45feea297be8bec5c7aee9ac1999b88ff1c98840cda7"))
	log.Printf("privKey01: %+v \n", hex.EncodeToString(privKey01))
	// log.Printf("privKey01: %+v \n", "84e57a4940d910677631bb767c4052e8ca72782609ce186a6d876eed4328cac7")
	lwsPubkey, lwsPrivKey, signKey := crypto.GenerateKeyPairBySeed(privKey01, 1)
	log.Printf("lwsPubkey : %+v \n", hex.EncodeToString(lwsPubkey[:]))
	log.Printf("lwsPrivKey : %+v \n", hex.EncodeToString(lwsPrivKey[:]))
	log.Printf("signKey : %+v \n", hex.EncodeToString(signKey[:]))

	payload, _ := hex.DecodeString(reverseString("65fe2685e91a07e2212ab39034c308fa874754bfc6bd31f6007e79ea8ab62f70"))
	log.Printf("payload : %+v \n", hex.EncodeToString(payload[:]))

	signStr := edwards25519.Sign(signKey[:], payload)
	log.Printf("signStr : %+v \n", hex.EncodeToString(signStr[:]))

	signature, _ := hex.DecodeString(reverseString("09c2431a868e249b5a19df7e52ad834b18903821d13fb2e82e2cea04cfd1141c959955cb4370c9c45514efae01331334dad03763e1a887c1d18b4ac820e30d60"))
	log.Printf("signature : %+v \n", hex.EncodeToString(signature[:]))

	// privKey01 = reverseBytes(privKey01)
	// log.Printf("reverseBytes privKey01: %+v \n", privKey01)
	// pubKey02, _ := hex.DecodeString(reverseString("b20aead7d116fa3f1b45731b2e845042ae4458ad43d8b6d55e3606b836df1f15"))
	// log.Printf("pubKey02: %+v \n", hex.EncodeToString(pubKey02))
	// // log.Printf("pubKey02: %+v \n", "98197a8df8bb51817022237a4da7d97263c01fe81e218fecfee84034e3a4b241")
	// // pubKey02 = reverseBytes(pubKey02)
	// // log.Printf("reverseBytes pubKey02: %+v \n", pubKey02)
	// privKey02, _ := hex.DecodeString(reverseString("59080d87613785b40eddff0cc40a211e75770ea62e1682955ac139c1372c7e00"))
	// log.Printf("privKey02: %+v \n", hex.EncodeToString(privKey02))
	// // log.Printf("privKey02: %+v \n", "3df8569f659033e437c80df0f80f32fc354541a80e249e50ab8b796239a86acd")
	// cliPubkey, cliPrivKey, _ := crypto.GenerateKeyPairBySeed(privKey02, 2)
	// log.Printf("cliPubkey : %+v \n", hex.EncodeToString(cliPubkey[:]))
	// log.Printf("cliPrivKey : %+v \n", hex.EncodeToString(cliPrivKey[:]))
	// // privKey02 = reverseBytes(privKey02)
	// // log.Printf("reverseBytes privKey02: %+v \n", privKey02)
	// var pubKeyLws crypto.PublicKey
	// var pubKeyCli crypto.PublicKey
	// copy(pubKeyLws[:], lwsPubkey[:])
	// copy(pubKeyCli[:], cliPubkey[:])
	// var privKeyLws crypto.PrivateKey
	// var privKeyCli crypto.PrivateKey
	// copy(privKeyLws[:], lwsPrivKey[:])
	// copy(privKeyCli[:], cliPrivKey[:])
	// apikey := crypto.GenerateApiKey(&privKeyCli, &pubKeyLws)
	// log.Printf("apikey: %+v \n", hex.EncodeToString(apikey[:]))

	// apikey02 := crypto.GenerateApiKey(&privKeyLws, &pubKeyCli)
	// log.Printf("apikey02: %+v \n", hex.EncodeToString(apikey02[:]))

	// sharedKey01, err := hex.DecodeString("91bbc9bc3a7b3d282ee538ed3f74fe2b1a14088c6caa1ad7c68663604301269d")
	// if err != nil {
	// 	t.Error("gene privKey01 failed!")
	// }
	// log.Printf("sharedKey01: %+v \n", sharedKey01)
}

// func TestAPISS(t *testing.T) {

// code1 := "90 134 255 167 38 42 220 84 131 70 81 116 43 64 75 75 2 0 0 0 0 0 0 0 100 0 0 0 0 0 0 39 16 178 10 234 215 209 22 250 63 27 69 115 27 46 132 80 66 174 68 88 173 67 216 182 213 94 54 6 184 54 223 31 21 1"
// codeArr1 := strings.Split(code1, " ")
// log.Printf("len : %+v", len(codeArr1))
// codeAdd1 := make([]byte, 66)
// log.Printf("len : %+v", codeArr1)
// for index := 0; index < len(codeArr1); index++ {
// 	value, _ := strconv.Atoi(codeArr1[index])
// 	codeAdd1[index] = byte(value)
// }
// log.Printf("codeAdd1 : %+v", codeAdd1)
// log.Printf("cliPubkey : %+v", hex.EncodeToString(codeAdd1[:]))
// pubKey02, _ := hex.DecodeString(reverseString("b20aead7d116fa3f1b45731b2e845042ae4458ad43d8b6d55e3606b836df1f15"))
// log.Printf("pubKey02: %+v \n", hex.EncodeToString(pubKey02))
// privKey02, _ := hex.DecodeString("c70f5d194f5093c5066889790af15069fa8fa3a8a8045f1462e7a49c019dbaf0")
// var pubKeyCli crypto.PublicKey
// copy(pubKeyCli[:], pubKey02[:])
// var privKeyLws crypto.PrivateKey
// copy(privKeyLws[:], privKey02[:])
// apikey := crypto.GenerateApiKey(&privKeyLws, &pubKeyCli)
// log.Printf("apikey: %+v \n", hex.EncodeToString(apikey[:]))
// log.Printf("codeAdd1 : %+v", len("c8f10736fb9b03a2d224c9d79b60ccc156b4bf9c28072fb332d0ea5fc104e085"))
// fork, _ := hex.DecodeString("c8f10736fb9b03a2d224c9d79b60ccc156b4bf9c28072fb332d0ea5fc104e085")
// log.Printf("codeAdd1 : %+v", reverseString(hex.EncodeToString(codeAdd1[:32])))
// cliPubkey, _, _ := crypto.GenerateKeyPairBySeed(codeAdd1[:32], 2)
// log.Printf("cliPubkey : %+v", reverseString(hex.EncodeToString(cliPubkey[:])))
// s := ServicePayload{}
// err := DecodePayload(codeAdd1, &s)
// if err != nil {
// 	log.Printf("err: %+v", err)
// }
// log.Printf("s: %+v", s)

// }

func TestSignAndVerify(t *testing.T) {
	privKey, _ := hex.DecodeString("007E2C37C139C15A9582162EA60E77751E210AC40CFFDD0EB4853761870D0859")
	log.Printf("privKey : %+v", privKey)

	payload := "1 0 1 21 31 223 54 184 6 54 94 213 182 216 67 173 88 68 174 66 80 132 46 27 115 69 27 63 250 22 209 215 234 10 178 1 0 0 0 155 135 201 91 1 200 241 7 54 251 155 3 162 210 36 201 215 155 96 204 193 86 180 191 156 40 7 47 179 50 208 234 95 193 4 224 133 0 0 84 84 0"
	payloadStr := strings.Split(payload, " ")
	log.Printf("len : %+v", len(payloadStr))
	payloadArr := make([]byte, 81)
	log.Printf("len : %+v", payloadStr)
	for index := 0; index < len(payloadStr); index++ {
		value, _ := strconv.Atoi(payloadStr[index])
		payloadArr[index] = byte(value)
	}
	log.Printf("privKeyArr : %+v", payloadArr)
	pubKey, _ := hex.DecodeString("151fdf36b806365ed5b6d843ad5844ae4250842e1b73451b3ffa16d1d7ea0ab2")
	log.Printf("pubKey : %+v", pubKey)

	pubKeys, privKeys, signKeys := crypto.GenerateKeyPairBySeed(privKey, 1)
	log.Printf("pubKeys : %+v", hex.EncodeToString(pubKeys[:]))
	log.Printf("privKeys : %+v", hex.EncodeToString(privKeys[:]))
	log.Printf("signKeys : %+v", hex.EncodeToString(signKeys[:]))
	var privateKey edwards25519.PrivateKey
	copy(privateKey[:], signKeys[:])
	// copy(privateKey[32:], pubKey[:])
	log.Printf("privateKey : %+v", privateKey)
	signStr := edwards25519.Sign(signKeys[:], payloadArr)
	log.Printf("signStr : %+v", hex.EncodeToString(signStr))
	// pubKey := "21 31 223 54 184 6 54 94 213 182 216 67 173 88 68 174 66 80 132 46 27 115 69 27 63 250 22 209 215 234 10 178"
	// pubKeyStr := strings.Split(pubKey, " ")
	// log.Printf("len : %+v", len(pubKeyStr))
	// pubKeyArr := make([]byte, 32)
	// log.Printf("len : %+v", pubKeyStr)
	// for index := 0; index < len(pubKeyStr); index++ {
	// 	value, _ := strconv.Atoi(pubKeyStr[index])
	// 	pubKeyArr[index] = byte(value)
	// }
	// log.Printf("pubKeyArr : %+v", pubKeyArr)
}

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

func TestSignCase(t *testing.T) {
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
	connection := db.GetConnection()
	user := model.User{}

	connection.LogMode(true)
	found := connection.Where("address_id = ?", 2).Find(&user).RecordNotFound()
	if !found {
		log.Printf("user: %+v", user)
	}
	// 	log.Printf("user: %s", hex.EncodeToString(user.Address))
	var UTXOs []UTXO
	// 	var utxol []model.Utxo

	// 	// connection.Raw("SELECT * FROM utxo WHERE ID = ?", 1).Scan(&utxol)
	// 	log.Printf("utxol: %+v", utxol)

	connection.Raw("SELECT "+
		"utxo.tx_hash AS tx_id, "+
		"utxo.out, "+
		"utxo.block_height, "+
		"tx.tx_type, "+
		"utxo.amount, "+
		"tx.sender AS sender, "+
		"tx.lock_until, "+
		"tx.data "+
		"FROM utxo "+
		"INNER JOIN tx "+
		"ON utxo.tx_hash = tx.hash "+
		"AND utxo.destination = ? "+
		"ORDER BY utxo.amount ASC, utxo.out ASC ", user.Address).Find(&UTXOs)
	// 	// Scan(&UTXOs)
	log.Printf("UTXOs: %+v", UTXOs)
	result, err := UTXOListToByte(&UTXOs)
	if err != nil {
		log.Printf("err: %+v", err)
	}
	log.Printf("result: %+v", result)

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
	// log.Printf("user : %+v \n", user)
	// cliMap := CliMap{}
	// value, err := redis.Values(redisConn.Do("hgetall", 9))
}
