package utxo

// import (
// 	"bytes"
// 	"context"
// 	"encoding/gob"
// 	"encoding/hex"
// 	"log"
// 	"os"

// 	"github.com/FissionAndFusion/lws/internal/config"
// 	"github.com/FissionAndFusion/lws/internal/db/service/block"
// 	"github.com/FissionAndFusion/lws/internal/gateway/mqtt"
// 	"github.com/gomodule/redigo/redis"
// )

// type RedisPublisher struct {
// 	conn redis.Conn
// }

// var redisPublisher *RedisPublisher

// func (this *RedisPublisher) CloseConn() {
// 	if this.conn != nil {
// 		this.conn.Close()
// 		this.conn = nil
// 	}
// }

// func (this *RedisPublisher) GetConn() redis.Conn {
// 	if this.conn != nil {
// 		return this.conn
// 	}
// 	redisConn := this.NewConn()
// 	return redisConn
// }

// func (this *RedisPublisher) NewConn() redis.Conn {
// 	redisPool := mqtt.GetRedisPool()
// 	return redisPool.Get()
// }

// func InitPubInstance(ctx context.Context) *RedisPublisher {
// 	redisPublisher = &RedisPublisher{}
// 	redisPublisher.NewConn()

// 	go func() {
// 		defer redisPublisher.CloseConn()
// 		<-ctx.Done()
// 	}()
// 	return redisPublisher
// }

// func NewUtxoUpdate(utxoUpdateList []mqtt.UTXOUpdate, address []byte) {
// 	// check should-write via reddis
// 	redisPool := mqtt.GetRedisPool()
// 	redisConn := redisPool.Get()
// 	defer redisConn.Close()
// 	cliMap, err := mqtt.GetUserByAddress(address, &redisConn)
// 	if err != nil {
// 		log.Printf("[ERROR] getUserByAddress failed: %+v", err)
// 		return
// 	}

// 	if cliMap == nil {
// 		// log.Printf("[ERROR] address not found")
// 		return
// 	}

// 	client := mqtt.GetProgram()
// 	if client.Client == nil {
// 		log.Printf("[ERROR] new mqtt client err: client == nil")
// 		return
// 	}

// 	updatePayload := mqtt.UpdatePayload{}
// 	updatePayload.Nonce = cliMap.Nonce
// 	updatePayload.AddressId = cliMap.AddressId
// 	tailBlock := block.GetTailBlock()
// 	if tailBlock == nil {
// 		log.Printf("[ERROR] tailBlock get nil")
// 		return
// 	}

// 	updatePayload.BlockHash = tailBlock.Hash
// 	updatePayload.Height = tailBlock.Height
// 	updatePayload.BlockTime = tailBlock.Tstamp
// 	forkId, err := hex.DecodeString(os.Getenv("FORK_ID"))
// 	if err != nil {
// 		log.Printf("[ERROR] Getenv FORK_ID err: %+v", err)
// 		return
// 	}
// 	updatePayload.ForkId = forkId

// 	replyUTXON := cliMap.ReplyUTXON

// 	queueItem := mqtt.UTXOUpdateQueueItem{
// 		UpdatePayload: &updatePayload,
// 		UpdateList:    utxoUpdateList,
// 		ReplySize:     replyUTXON,
// 	}

// 	var buf bytes.Buffer
// 	enc := gob.NewEncoder(&buf)
// 	err = enc.Encode(queueItem)
// 	if err != nil {
// 		log.Printf("[ERROR] encode utxoUpdate error: %s, item: %v", err, queueItem)
// 	}
// 	msg := buf.Bytes()
// 	// publish utxo update list to mq

// 	topic := config.Config.UTXO_UPDATE_QUEUE_NAME + "lwsid"
// 	redisPublisher.GetConn().Do("PUBLISH", topic, msg)
// }
