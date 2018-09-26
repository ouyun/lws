package mqtt

import (
	"encoding/hex"
	"log"
	"math"
	"os"

	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/stream/block"
	"github.com/eclipse/paho.mqtt.golang"
)

// send UTXO update list
func SendUTXOUpdate(client *mqtt.Client, u *[]UTXOUpdate, address []byte) {
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
		log.Printf("err: %+v", err)
		return
	}
	updatePayload.ForkId = forkId

	// send
	c := make(chan bool)
	if user.ReplyUTXON < uint16(len(*(u))) {
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
				SendUpdateMessage(client, &updatePayload, utxoUList[user.ReplyUTXON*uint16(index):rightIndex], &cliMap, 1, c)
				<-c
				continue
			}
			SendUpdateMessage(client, &updatePayload, utxoUList[user.ReplyUTXON*uint16(index):], &cliMap, 1, c)
			<-c
		}
	} else {
		//发送一次
		SendUpdateMessage(client, &updatePayload, utxoUList, &cliMap, 0, c)
	}
}

// send update message
func SendUpdateMessage(client *mqtt.Client, uPayload *UpdatePayload, u []UTXOUpdate, cliMap *CliMap, end int, c chan bool) {
	uPayload.UpdateNum = uint16(len(u))
	uList, errs := StructToBytes(u)
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
	if token.Wait() {
		c <- true
	}
}
