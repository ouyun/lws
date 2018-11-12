package mqtt

import (
	"bytes"
	"encoding/hex"
	"log"
	"os"
	"strconv"

	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/blake2b"
	edwards25519 "golang.org/x/crypto/ed25519"
)

var serviceReqHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Println("Received serviceReq !")
	s := ServicePayload{}
	cliMap := CliMap{}
	user := model.User{}
	newUser := model.User{}
	var pubKey []byte
	var forkBitmap uint64 = 1
	err := DecodePayload(msg.Payload(), &s)
	if err != nil {
		log.Printf("Discard req with decodePayload err: %+v\n", err)
		//丢弃请求
		return
	}
	pubKey = PayloadToUser(&newUser, &s)

	// 验证签名
	if !VerifyAddress(&s, msg.Payload()) {
		//丢弃请求
		log.Printf("sign err ！discard serviceReq data!\n")
		return
	}

	//TODO: 检查分支
	forkId, err := hex.DecodeString(os.Getenv("FORK_ID"))
	if err != nil {
		ReplyServiceReq(&client, forkBitmap, 16, &s, &newUser, pubKey)
		return
	}

	suportForkId := false
	for index := 0; index < (len(s.ForkList) / 32); index++ {
		if bytes.Compare(s.ForkList[(index*32):((index+1)*32)], forkId) == 0 {
			suportForkId = true
			forkBitmap = forkBitmap << uint(index)
			break
		}
	}
	if !suportForkId {
		ReplyServiceReq(&client, forkBitmap, 2, &s, &newUser, pubKey)
		return
	}

	// 连接 redis && db
	pool := GetRedisPool()
	redisConn := pool.Get()
	defer redisConn.Close()
	connection := db.GetConnection()
	transaction := connection.Begin()
	if err != nil {
		log.Printf("err: %+v\n", err)
		ReplyServiceReq(&client, forkBitmap, 16, &s, &newUser, pubKey)
	}

	// 查询 redis
	// 没有 -> 查询 数据库
	found := transaction.Where("address = ?", s.Address).First(&user).RecordNotFound()
	copyUserStruct(&newUser, &user)
	if !found {
		// update user
		err = transaction.Save(&user).Error // update user
		if err != nil {
			// fail
			log.Printf("err: %+v\n", err)
			transaction.Rollback()
			ReplyServiceReq(&client, forkBitmap, 16, &s, &user, pubKey)
			return
		}
		log.Printf("user: %+v\n", user)
		copyUserToCliMap(&user, &cliMap)
		log.Println("update user !")
	} else {
		// save user
		err = SaveUser(transaction, &user)
		if err != nil {
			// fail
			transaction.Rollback()
			log.Printf("SaveUser get err: %+v\n", err)
			ReplyServiceReq(&client, forkBitmap, 16, &s, &user, pubKey)
			return
		}
		copyUserToCliMap(&user, &cliMap)
		log.Printf("saved User! \n")
	}
	// 保存到 redis
	err = SaveToRedis(&redisConn, &cliMap)
	if err != nil {
		// fail
		RemoveRedis(&redisConn, &cliMap)
		transaction.Rollback()
		log.Printf("SaveToRedis get err: %+v\n", err)
		ReplyServiceReq(&client, forkBitmap, 16, &s, &user, pubKey)
		return
	}
	transaction.Commit()
	ReplyServiceReq(&client, forkBitmap, 0, &s, &user, pubKey)
}

// reply service req
func ReplyServiceReq(client *mqtt.Client, forkBitmap uint64, err int, s *ServicePayload, user *model.User, pubKey []byte) {
	reply := ServiceReply{}
	reply.Nonce = s.Nonce
	reply.Version = s.Version
	reply.Error = uint8(err)
	if err == 0 {
		reply.AddressId = user.AddressId
		reply.ForkBitmap = forkBitmap
		reply.ApiKeySeed = pubKey[:]
	}
	result, errs := StructToBytes(reply)
	if errs != nil {
		log.Printf("err: %+v\n", err)
	}
	t := s.TopicPrefix + "/fnfn/ServiceReply"
	// TODO
	token := (*client).Publish(t, 0, false, result)
	if token.Wait() {
		log.Printf("publish err: %+v", token.Error())
	}
}

// save to redis
func SaveToRedis(conn *redis.Conn, cliMap *CliMap) (err error) {
	// save struct
	_, err = (*conn).Do("HMSET", redis.Args{}.Add(strconv.FormatUint(uint64(cliMap.AddressId), 10)).AddFlat(cliMap)...)
	if err != nil {
		return err
	}
	// save set
	_, err = (*conn).Do("SET", hex.EncodeToString(cliMap.Address), cliMap.AddressId)
	return err
}

func RemoveRedis(conn *redis.Conn, cliMap *CliMap) (err error) {
	_, err = (*conn).Do("DEL", strconv.FormatUint(uint64(cliMap.AddressId), 10))
	if err != nil {
		log.Printf("delete key failed")
	}
	// delete key
	_, err = (*conn).Do("DEL", hex.EncodeToString(cliMap.Address))
	if err != nil {
		log.Printf("delete key failed")
	}
	return err
}

func VerifyAddress(s *ServicePayload, payload []byte) bool {
	messageLen := uint16(len(payload)) - (s.SignBytes + 2)
	log.Printf("pub Key: %+v", hex.EncodeToString(s.Address[1:]))
	log.Printf("payload : %+v", payload[:messageLen])
	if messageLen > uint16(len(payload)) {
		log.Printf("VerifyAddress failed with err: slice bounds out of range")
		return false
	}
	if uint8(s.Address[0]) == 1 {
		// 验证签名
		log.Printf("VerifyAddress with address type = 1 \n")
		signatureHash := blake2b.Sum256(payload[:messageLen])
		return edwards25519.Verify(s.Address[1:], signatureHash[:], s.ServSignature)
	}
	log.Printf("VerifyAddress with address type = 2 \n")
	// 验证 模版地址
	log.Printf("ServSignature: %d", s.ServSignature)
	templateDataLen := s.SignBytes - 96
	templateData := s.ServSignature[:templateDataLen]
	pubKey := s.ServSignature[templateDataLen:(templateDataLen + 32)]
	signature := s.ServSignature[(templateDataLen + 32):]
	hash := blake2b.Sum256(templateData)
	if bytes.Compare(hash[:30], s.Address[3:]) != 0 {
		return false
	}
	return edwards25519.Verify(pubKey, payload[:messageLen], signature)
}

func copyUserStruct(user *model.User, dbUser *model.User) {
	dbUser.ApiKey = user.ApiKey
	dbUser.Address = user.Address
	dbUser.ForkList = user.ForkList
	dbUser.ForkNum = user.ForkNum
	dbUser.TopicPrefix = user.TopicPrefix
	dbUser.TimeStamp = user.TimeStamp
	dbUser.ReplyUTXON = user.ReplyUTXON
}
