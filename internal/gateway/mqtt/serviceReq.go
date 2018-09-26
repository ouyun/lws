package mqtt

import (
	"bytes"
	"encoding/hex"
	"log"
	"os"

	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
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
	pool := NewRedisPool()
	redisConn := pool.Get()
	defer redisConn.Close()
	connection := db.GetConnection()
	if err != nil {
		log.Printf("err: %+v\n", err)
		ReplyServiceReq(&client, forkBitmap, 16, &s, &user, pubKey)
	}

	// 查询 redis
	// 没有 -> 查询 数据库
	found := connection.Where("address = ?", s.Address).First(&user).RecordNotFound()
	if !found {
		// update user
		err = connection.Save(&user).Error // update user
		if err != nil {
			// fail
			// TODO：回退
			log.Printf("err: %+v\n", err)
			ReplyServiceReq(&client, forkBitmap, 16, &s, &user, pubKey)
			return
		}
	} else {
		// save user
		err = SaveUser(connection, &user)
		if err != nil {
			// fail
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
	_, err = (*conn).Do("HMSET", redis.Args{}.Add(
		cliMap.AddressId).AddFlat(cliMap)...)
	if err != nil {
		return err
	}
	// save set
	_, err = (*conn).Do("SET", cliMap.Address, cliMap.AddressId)
	return err
}
