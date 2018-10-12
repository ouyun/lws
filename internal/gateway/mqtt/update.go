package mqtt

import (
	"encoding/hex"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/db/service/block"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
)

// send UTXO update list
func SendUTXOUpdate(u *[]UTXOUpdate, address []byte) {
	log.Println("update utxoUpdate !")
	client, err := NewClient()
	if err != nil {
		log.Printf("new mqtt client err: %+v", err)
	}
	updatePayload := UpdatePayload{}
	user := model.User{}
	cliMap := CliMap{}
	pool := GetRedisPool()
	utxoUList := *u

	// get user by address
	redisConn := pool.Get()
	connection := db.GetConnection()
	GetUserByAddress(address, connection, &redisConn, &user, &cliMap)
	updatePayload.Nonce = cliMap.Nonce
	updatePayload.AddressId = cliMap.AddressId
	tailBlock := block.GetTailBlock()
	updatePayload.BlockHash = tailBlock.Hash
	updatePayload.Height = tailBlock.Height
	forkId, err := hex.DecodeString(os.Getenv("FORK_ID"))
	if err != nil {
		log.Printf("Getenv FORK_ID err: %+v", err)
		return
	}
	updatePayload.ForkId = forkId

	// send
	c := make(chan int, 1)
	if user.ReplyUTXON < uint16(len(*(u))) && user.ReplyUTXON != 0 {
		// 多次发送

		// 发送次数
		times := int(math.Ceil(float64(uint16(len(*u)) / user.ReplyUTXON)))
		for index := 0; index < times; index++ {
			if index != (times - 1) {
				// TODO: sync
				var rightIndex uint16
				if (user.ReplyUTXON * uint16(index+1)) <= uint16(len(*u)) {
					rightIndex = (user.ReplyUTXON * uint16(index+1)) - 1
				} else {
					rightIndex = uint16(len(*u)) - 1
				}
				SendUpdateMessage(&client.Client, &redisConn, &updatePayload, utxoUList[user.ReplyUTXON*uint16(index):rightIndex], &cliMap, 1, c)
				<-c
				continue
			}
			SendUpdateMessage(&client.Client, &redisConn, &updatePayload, utxoUList[user.ReplyUTXON*uint16(index):], &cliMap, 0, c)
			<-c
		}
	} else {
		//发送一次
		SendUpdateMessage(&client.Client, &redisConn, &updatePayload, utxoUList, &cliMap, 0, c)
		<-c
	}
	log.Printf("--------------------------")
}

// send update message
func SendUpdateMessage(client *mqtt.Client, redisConn *redis.Conn, uPayload *UpdatePayload, u []UTXOUpdate, cliMap *CliMap, end int, c chan int) {
	addressIdStr := strconv.FormatUint(uint64(cliMap.AddressId), 10)
	cli := CliMap{}
	value, err := redis.Values((*redisConn).Do("hgetall", addressIdStr))
	if err != nil {
		log.Printf("redis  err: %+v\n", err)
		return
	}
	redis.ScanStruct(value, &cli)
	if cli.AddressId != 0 && cli.Nonce == 0 {
		log.Printf("discard  err: %+v\n", err)
		return
	}
	uPayload.UpdateNum = uint16(len(u))
	uList, errs := UTXOUpdateListToByte(&u)
	if errs != nil {
		log.Printf("struct To bytes err: %+v\n", errs)
		return
	}
	uPayload.UpdateList = uList
	uPayload.Continue = uint8(end)
	result, errs := StructToBytes(uPayload)
	if errs != nil {
		log.Printf("struct to bytes err: %+v\n", errs)
		return
	}
	t := cliMap.TopicPrefix + "/fnfn/UTXOUpdate"
	// TODO: topicprefix
	token := (*client).Publish(t, 1, false, result)
	for {
		if token.Wait() && token.Error() == nil {
			c <- 1
			log.Printf("err: %+v\n", token.Error())
			break
		}
	}
}
