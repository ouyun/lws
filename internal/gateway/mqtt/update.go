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
	"github.com/FissionAndFusion/lws/test/helper"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
)

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

func NewUTXOUpdate(utxoUpdateList []UTXOUpdate, address []byte, wg *sync.WaitGroup) {
	defer helper.MeasureTime(helper.MeasureTitle("queue utxo"))
	defer wg.Done()
	defer log.Printf("[DEBUG] wg done")
	log.Printf("[DEBUG] new utxo update [%s]", hex.EncodeToString(address))
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
		log.Printf("[DEBUG] can not found climap address [%s]", hex.EncodeToString(address))
		return
	}

	if cliMap.Nonce == 0 {
		log.Printf("[DEBUG] can not found climap address [%s] topic[%s] nonce 0", hex.EncodeToString(address), cliMap.TopicPrefix)
		return
	}

	log.Printf("[DEBUG] queue utxo update to %s: addr[%s]", cliMap.TopicPrefix, hex.EncodeToString(address))

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
	topic := conf.UTXO_UPDATE_QUEUE_NAME + "lwsid"
	log.Printf("[DEBUG] redis topic %s", topic)
	pubRedisConn := redisPool.Get()
	defer pubRedisConn.Close()
	redisConn.Do("PUBLISH", topic, msg)
	// _, err = redis.DoWithTimeout(redisPublisher.GetConn(), 5*time.Second, "PUBLISH", topic, msg)
	// if err != nil {
	// 	log.Printf("[ERROR] send redis failed [%s]", err)
	// }
	log.Printf("[DEBUG] New utxo update done")
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

	if cliMap.Nonce == 0 {
		// sync is not subscribed
		return
	}

	log.Printf("[DEBUG] send utxo update to %s", cliMap.TopicPrefix)

	client := GetProgram()
	if client == nil || client.Client == nil {
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
	defer helper.MeasureTime(helper.MeasureTitle("send utxo update message"))
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
					log.Printf("[INFO] redis subscription 0, close")
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
