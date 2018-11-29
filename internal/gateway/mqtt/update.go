package mqtt

import (
	"encoding/hex"
	"log"
	"math"
	"os"
	"strconv"
	"sync"

	"github.com/FissionAndFusion/lws/internal/db/service/block"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
)

var mu sync.Mutex
var cliProgram *Program

// send UTXO update list
func SendUTXOUpdate(u *[]UTXOUpdate, address []byte) {
	utxoUList := *u

	log.Printf("[DEBUG] send utxo update cnt[%d] address: %+v !", len(utxoUList), address)
	updatePayload := UpdatePayload{}

	pool := GetRedisPool()
	redisConn := pool.Get()

	defer redisConn.Close()
	c := make(chan int, 1)

	// get user by address
	cliMap, err := GetUserByAddress(address, &redisConn)
	if err != nil {
		log.Printf("[ERROR] getUserByAddress failed: %+v", err)
		return
	}

	if cliMap == nil {
		// log.Printf("[ERROR] address not found")
		return
	}

	client := GetProgram()
	if client.Client == nil {
		log.Printf("[ERROR] new mqtt client err: client == nil")
		return
	}
	updatePayload.Nonce = cliMap.Nonce
	updatePayload.AddressId = cliMap.AddressId
	tailBlock := block.GetTailBlock()
	if tailBlock == nil {
		log.Printf("[ERROR] tailBlock get nil")
		return
	}

	updatePayload.BlockHash = tailBlock.Hash
	updatePayload.Height = tailBlock.Height
	updatePayload.BlockTime = tailBlock.Tstamp
	forkId, err := hex.DecodeString(os.Getenv("FORK_ID"))
	if err != nil {
		log.Printf("[ERROR] Getenv FORK_ID err: %+v", err)
		return
	}
	updatePayload.ForkId = forkId

	replyUTXON := cliMap.ReplyUTXON

	// send
	if replyUTXON < uint16(len(*(u))) && replyUTXON != 0 {
		// 多次发送

		// 发送次数
		times := int(math.Ceil(float64(uint16(len(*u)) / replyUTXON)))
		for index := 0; index < times; index++ {
			if index != (times - 1) {
				// TODO: sync
				var rightIndex uint16
				if (replyUTXON * uint16(index+1)) <= uint16(len(*u)) {
					rightIndex = (replyUTXON * uint16(index+1)) - 1
				} else {
					rightIndex = uint16(len(*u)) - 1
				}
				SendUpdateMessage(&(*client).Client, &redisConn, &updatePayload, utxoUList[replyUTXON*uint16(index):rightIndex], cliMap, 1, c)
				<-c
				continue
			}
			SendUpdateMessage(&(*client).Client, &redisConn, &updatePayload, utxoUList[replyUTXON*uint16(index):], cliMap, 0, c)
			<-c
		}
	} else {
		// 发送一次
		SendUpdateMessage(&(*client).Client, &redisConn, &updatePayload, utxoUList, cliMap, 0, c)
		<-c
	}
}

// send update message
func SendUpdateMessage(client *mqtt.Client, redisConn *redis.Conn, uPayload *UpdatePayload, u []UTXOUpdate, cliMap *CliMap, end int, c chan int) {
	log.Printf("[DEBUG] UTXOUpdate addressId[%d] cnt[%d] continue[%d]:", cliMap.AddressId, len(u), end)
	addressIdStr := strconv.FormatUint(uint64(cliMap.AddressId), 10)
	cli := CliMap{}
	value, err := redis.Values((*redisConn).Do("hgetall", addressIdStr))
	if err != nil {
		log.Printf("[ERROR] redis  err: %+v\n", err)
		return
	}
	redis.ScanStruct(value, &cli)
	if cli.AddressId != 0 && cli.Nonce == 0 {
		log.Printf("[ERROR] discard  err: %+v\n", err)
		return
	}
	uPayload.UpdateNum = uint16(len(u))
	uList, errs := UTXOUpdateListToByte(&u)
	if errs != nil {
		log.Printf("[ERROR] struct To bytes err: %+v\n", errs)
		return
	}
	uPayload.UpdateList = uList
	uPayload.Continue = uint8(end)
	result, errs := StructToBytes(*uPayload)
	if errs != nil {
		log.Printf("[ERROR] struct to bytes err: %+v\n", errs)
		return
	}
	t := cliMap.TopicPrefix + "/fnfn/UTXOUpdate"
	// TODO: topicprefix
	token := (*client).Publish(t, 1, false, result)
	token.Wait()
	err = token.Error()
	if err != nil {
		log.Printf("[ERROR] publish [%s] err: %v", t, err)
	}
	c <- 1
}

func GetProgram() *Program {
	if cliProgram == nil {
		mu.Lock()
		defer mu.Unlock()
		if cliProgram == nil {
			id := os.Getenv("MQTT_STREAM_CLIENT_ID")
			if id == "" {
				id = "lws-001-update"
			}
			cliProgram = &Program{Id: id, IsLws: false}
			cliProgram.Init()
			if err := cliProgram.Start(); err != nil {
				log.Printf("[ERROR] client start failed: %s", err)
			}
		}
	}
	return cliProgram
}
