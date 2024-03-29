package mqtt

import (
	"bytes"
	"log"

	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/gateway/crypto"
	"github.com/eclipse/paho.mqtt.golang"
)

var uTXOAbortReqHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	a := AbortPayload{}
	payload := msg.Payload()
	cliMap := CliMap{}
	user := model.User{}
	err := DecodePayload(payload, &a)
	// log.Printf("AbortPayload: %+v\n", a)
	if err != nil {
		log.Printf("err: %+v\n", err)
	}
	// 连接 redis
	pool := GetRedisPool()
	redisConn := pool.Get()
	connection := db.GetConnection()
	defer redisConn.Close()

	inRedis, inDb, err := CheckAddressId(a.AddressId, connection, &redisConn, &user, &cliMap)
	// 验证签名
	signed := crypto.SignWithApiKey(cliMap.ApiKey, payload[:len(payload)-20])
	if bytes.Compare(signed, a.Signature) != 0 {
		// 丢弃 内容
		return
	}
	if err != nil {
		log.Printf("abort err: %+v\n", err)
		return
	}
	if !inRedis && !inDb {
		log.Printf("can not found any user by  addressId \n")
		return
	}
	updateRedis(&redisConn, &cliMap, "Nonce", 0)
}
