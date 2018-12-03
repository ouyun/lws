package utxo

import (
	"context"
	"encoding/gob"
	"log"

	"github.com/FissionAndFusion/lws/internal/gateway/mqtt"
	"github.com/furdarius/rabbitroutine"
	"github.com/gomodule/redigo/redis"
)

var pubInstance rabbitroutine.Publisher
var redisConn redis.Conn

func InitPubInstance(ctx context.Context) {
	amqpUrl := os.Getenv("AMQP_URL")
	conn := rabbitroutine.NewConnector(rabbitroutine.Config{
		// Max reconnect attempts
		ReconnectAttempts: 0,
		// How long wait between reconnect
		Wait: 30 * time.Second,
	})

	pool := rabbitroutine.NewPool(conn)
	ensurePub := rabbitroutine.NewEnsurePublisher(pool)
	pub := rabbitroutine.NewRetryPublisher(ensurePub)

	pubInstance = pub

	redisPool := GetRedisPool()
	redisConn = redisPool.Get()

	go func() {
		defer redisConn.Close()
		err := conn.Dial(ctx, amqpUrl)
		if err != nil {
			log.Println("[ERROR] failed to establish RabbitMQ connection:", err)
		}

		<-ctx.Done()
	}()
}

func GetPubInstance() rabbitroutine.Publisher {
	return pubInstance
}

func NewUtxoUpdate(utxoUpdateList []mqtt.UTXOUpdate, address []byte) {
	// check should-write via reddis
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
		updatePayload: &updatePayload,
		updateList:    utxoUpdateList,
		replySize:     replyUTXON,
	}

	// publish utxo update list to mq
}
