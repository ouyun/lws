package mqtt

import (
	"bytes"
	"encoding/hex"
	"log"
	"os"
	"strconv"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
	"github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/internal/db/model"
)

var serviceReqHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	s := ServicePayload{}
	cliMap := CliMap{}
	user := model.User{}
	var pubKey []byte
	var forkBitmap uint64 = 1
	err := DecodePayload(msg.Payload(), &s)
	if err != nil {
		log.Printf("err: %+v\n", err)
		ReplyServiceReq(&client, forkBitmap, 16, &s, &user, pubKey)
		return
	}
	pubKey = PayloadToUser(&user, &s)

	//TODO: 检查分支
	forkId, err := hex.DecodeString(os.Getenv("FORK_ID"))
	if err != nil {
		ReplyServiceReq(&client, forkBitmap, 16, &s, &user, pubKey)
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
		ReplyServiceReq(&client, forkBitmap, 3, &s, &user, pubKey)
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
		ReplyServiceReq(&client, forkBitmap, 16, &s, &user, pubKey)
	}

	// 查询 redis
	// 没有 -> 查询 数据库
	found := transaction.Where("address = ?", s.Address).First(&user).RecordNotFound()
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
	} else {
		// save user
		err = SaveUser(connection, &user)
		if err != nil {
			// fail
			transaction.Rollback()
			log.Printf("err: %+v\n", err)
			ReplyServiceReq(&client, forkBitmap, 16, &s, &user, pubKey)
			return
		}
	}
	// 保存到 redis
	copyUserToCliMap(&user, &cliMap)
	err = SaveToRedis(&redisConn, &cliMap)
	if err != nil {
		// fail
		RemoveRedis(&redisConn, &cliMap)
		transaction.Rollback()
		log.Printf("err: %+v\n", err)
		ReplyServiceReq(&client, forkBitmap, 16, &s, &user, pubKey)
		return
	}
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
	(*client).Publish(t, 0, false, result)
}

// save to redis
func SaveToRedis(conn *redis.Conn, cliMap *CliMap) (err error) {
	// save struct
	// log.Printf("cliMap: %+v", cliMap)
	_, err = (*conn).Do("HMSET", redis.Args{}.Add(strconv.FormatUint(uint64(cliMap.AddressId), 10)).AddFlat(cliMap)...)
	log.Printf("err: %+v", err)
	if err != nil {
		return err
	}
	// save set
	_, err = (*conn).Do("SET", hex.EncodeToString(cliMap.Address), cliMap.AddressId)
	log.Printf("err: %+v", err)
	return err
}

func RemoveRedis(conn *redis.Conn, cliMap *CliMap) (err error) {
	_, err = (*conn).Do("DEL", cliMap.AddressId)
	if err != nil {
		log.Printf("del key failed")
	}
	// delete key
	_, err = (*conn).Do("DEL", cliMap.Address)
	if err != nil {
		log.Printf("del key failed")
	}
	return err
}
