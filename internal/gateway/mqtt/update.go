package mqtt

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/hex"
	"log"
	"math"
	"os"
	"strconv"
	"sync"

	"github.com/FissionAndFusion/lws/internal/config"
	"github.com/FissionAndFusion/lws/internal/db/service/block"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
)

var mu sync.Mutex
var cliProgram *Program

type RedisPublisher struct {
	conn redis.Conn
}

var redisPublisher *RedisPublisher

func (this *RedisPublisher) CloseConn() {
	if this.conn != nil {
		this.conn.Close()
		this.conn = nil
	}
}

func (this *RedisPublisher) GetConn() redis.Conn {
	if this.conn != nil {
		return this.conn
	}
	redisConn := this.NewConn()
	return redisConn
}

func (this *RedisPublisher) NewConn() redis.Conn {
	redisPool := GetRedisPool()
	return redisPool.Get()
}

func InitPubInstance(ctx context.Context) *RedisPublisher {
	redisPublisher = &RedisPublisher{}
	redisPublisher.NewConn()

	go func() {
		defer redisPublisher.CloseConn()
		<-ctx.Done()
	}()
	return redisPublisher
}

func NewUTXOUpdate(utxoUpdateList []UTXOUpdate, address []byte) {
	log.Printf("[DEBUG] new utxo update")
	// check should-write via reddis
	redisPool := GetRedisPool()
	redisConn := redisPool.Get()
	defer redisConn.Close()
	cliMap, err := GetUserByAddress(address, &redisConn)
	if err != nil {
		log.Printf("[ERROR] getUserByAddress failed: %+v", err)
		return
	}

	if cliMap == nil {
		// log.Printf("[ERROR] address not found")
		return
	}

	// client := GetProgram()
	// if client.Client == nil {
	// 	log.Printf("[ERROR] new mqtt client err: client == nil")
	// 	return
	// }

	updatePayload := UpdatePayload{}
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

	queueItem := UTXOUpdateQueueItem{
		UpdatePayload: updatePayload,
		UpdateList:    utxoUpdateList,
		ReplyUTXON:    replyUTXON,
		Address:       address,
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(queueItem)
	if err != nil {
		log.Printf("[ERROR] encode utxoUpdate error: %s, item: %v", err, queueItem)
	}
	msg := buf.Bytes()
	// publish utxo update list to mq

	conf := config.GetConfig()
	log.Printf("[DEBUG] conf %v", conf)
	topic := conf.UTXO_UPDATE_QUEUE_NAME + "lwsid"
	redisPublisher.GetConn().Do("PUBLISH", topic, msg)
}

// send UTXO update list
func SendUTXOUpdate(item *UTXOUpdateQueueItem) {
	utxoUList := item.UpdateList
	log.Printf("[DEBUG] send utxo update cnt[%d] address: %+v !", len(utxoUList), item.Address)

	pool := GetRedisPool()
	redisConn := pool.Get()

	address := item.Address
	updatePayload := item.UpdatePayload
	replyUTXON := item.ReplyUTXON

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

	// send
	if replyUTXON < uint16(len(utxoUList)) && replyUTXON != 0 {
		// 多次发送

		// 发送次数
		times := int(math.Ceil(float64(uint16(len(utxoUList)) / replyUTXON)))
		for index := 0; index < times; index++ {
			if index != (times - 1) {
				// TODO: sync
				var rightIndex uint16
				if (replyUTXON * uint16(index+1)) <= uint16(len(utxoUList)) {
					rightIndex = (replyUTXON * uint16(index+1)) - 1
				} else {
					rightIndex = uint16(len(utxoUList)) - 1
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

func ConsumerUTXOUpdate(channel string, message []byte) {
	log.Printf("[DEBUG] Consumer utxo update channel [%s]", channel)
	dec := gob.NewDecoder(bytes.NewBuffer(message))
	var item UTXOUpdateQueueItem
	err := dec.Decode(&item)
	if err != nil {
		log.Printf("[ERROR] decode error %s", err)
		return
	}
	SendUTXOUpdate(&item)
}

func ListenUTXOUpdateConsumer(ctx context.Context) error {
	// pool := GetRedisPool()
	// redisConn := pool.Get()

	redisConn, err := redis.DialURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Printf("[ERROR] connect redis pubsub conn failed %s", err)
		return err
	}

	// redisConn.Ping("haha")

	psc := redis.PubSubConn{Conn: redisConn}

	conf := config.GetConfig()
	log.Printf("[DEBUG] conf %v", conf)
	topic := conf.UTXO_UPDATE_QUEUE_NAME + "lwsid"

	if err := psc.Subscribe(redis.Args{}.AddFlat(topic)...); err != nil {
		log.Printf("[ERROR] subscribe [%s] failed", topic)
		return err
	}

	done := make(chan struct{}, 1)

	go (func(done chan struct{}) {

		for {
			switch n := psc.Receive().(type) {
			case error:
				log.Printf("[ERROR] redis subscribe received error %s", n)
				close(done)
			case redis.Message:
				ConsumerUTXOUpdate(n.Channel, n.Data)
			case redis.Subscription:
				if n.Count == 0 {
					close(done)
				}
			}
		}

	})(done)

	select {
	case <-ctx.Done():
	case <-done:
	}
	log.Printf("[INFO] unsubscribe utxo udpate listener")
	psc.Unsubscribe()
	redisConn.Close()

	return nil
}
