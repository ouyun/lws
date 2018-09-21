package mqtt

import (
	"bytes"
	"encoding/hex"
	"log"
	"math"
	"os"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/internal/db/model"
	"github.com/lomocoin/lws/internal/gateway/crypto"
	"github.com/lomocoin/lws/internal/stream/block"
)

var syncReqHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	// TODO ：
	var UTXOs []UTXO
	s := SyncPayload{}
	cliMap := CliMap{}
	pool := NewRedisPool()
	user := model.User{}

	redisConn := pool.Get()
	connection := db.GetConnection()
	payload := msg.Payload()
	err := DecodePayload(payload, &s)
	if err != nil {
		log.Printf("err: %+v\n", err)
	}
	// 连接 redis
	defer redisConn.Close()

	inRedis, inDb, err := CheckAddressId(s.AddressId, connection, &redisConn, &user, &cliMap)
	// 验证签名
	signed := crypto.SignWithApiKey(cliMap.ApiKey, payload[:len(payload)-20])
	if bytes.Compare(signed, payload[:len(payload)-20]) != 0 {
		// 丢弃 请求
		return
	}
	if err != nil {
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 16, 0)
		return
	}
	if !inRedis && !inDb {
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 2, 0)
		return
	}
	// 检查分支
	forkId, err := hex.DecodeString(os.Getenv("FORK_ID"))
	if err != nil {
		log.Printf("err: %+v", err)
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 16, 0)
	}
	if bytes.Compare(forkId, s.ForkID) != 0 {
		// 无效分支
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 3, 0)
		return
	}
	//get utxo list
	err = connection.Exec("SELECT"+
		"utxo.tx_hash AS tx_id,"+
		"utxo.out,"+
		"utxo.block_height,"+
		"utxo.amount,"+
		"tx.data,"+
		"tx.lock_until,"+
		"tx.send_to,"+
		"tx.tx_type"+
		"FROM utxo"+
		"INNER JOIN tx"+
		"ON utxo.tx_hash = tx.hash"+
		"AND utxo.SendTo = ? "+
		"ORDER BY tx_hash ASC, out ASC", cliMap.Address).Find(&UTXOs).Error
	if err != nil {
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 16, 0)
		return
	}
	// 计算utxo hash
	utxoHash := UTXOHash(&UTXOs)
	if bytes.Compare(utxoHash, []byte(s.UTXOHash)) == 0 {
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 0, 0)
		updateRedis(&redisConn, &cliMap, "Nonce", s.Nonce)
		return
	}
	// 计算utxo数量
	// 如果utxo 数量超过replyUtxo长度，分多次发送list
	// 如果replyUtxo = 0 ， 计算长度是否超过256
	if cliMap.ReplyUTXON == 0 {
		maxLen := 256 * 1024
		totalLen := len(UTXOs) + 42
		if maxLen < totalLen {
			// TODO：分包 发送
		} else {
			// 一次发送
			ReplySyncReq(&client, &s, &UTXOs, &cliMap, 1, 0)
		}
	} else if cliMap.ReplyUTXON < uint16(len(UTXOs)) {
		// 多次发送
		// 发送次数
		c := make(chan bool)
		times := int(math.Ceil(float64(uint16(len(UTXOs)) / cliMap.ReplyUTXON)))
		for index := 0; index < times; index++ {
			if index != (times - 1) {
				// TODO: sync
				var rightIndex uint16
				if (cliMap.ReplyUTXON * uint16(index+1)) <= uint16(len(UTXOs)) {
					rightIndex = (cliMap.ReplyUTXON * uint16(index+1)) - 1
				} else {
					rightIndex = uint16(len(UTXOs)) - 1
				}
				ReplySyncReqWithChan(&client, &s, UTXOs[cliMap.ReplyUTXON*uint16(index):rightIndex], &cliMap, 1, 1, c)
				<-c
				continue
			}
			ReplySyncReqWithChan(&client, &s, UTXOs[cliMap.ReplyUTXON*uint16(index):], &cliMap, 1, 0, c)
			<-c
		}
	} else {
		// 一次发送
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 1, 0)
	}
	// save nonce
	updateRedis(&redisConn, &cliMap, "Nonce", s.Nonce)
}

// reply sync req
func ReplySyncReq(client *mqtt.Client, s *SyncPayload, u *[]UTXO, cliMap *CliMap, err, end int) {
	reply := SyncReply{}
	reply.Nonce = s.Nonce
	reply.Error = uint8(err)
	if err == 0 {
		tailBlock := block.GetTailBlock()
		reply.BlockHash = tailBlock.Hash
		reply.BlockHeight = tailBlock.Height
		reply.UTXONum = uint16(0)
		reply.Continue = uint8(end)
	}
	if err == 1 {
		reply.UTXONum = uint16(len(*u))
		byteList, _ := UTXOListToByte(u)
		reply.UTXOList = byteList
		reply.Continue = uint8(end)
	}
	result, errs := StructToBytes(reply)
	if errs != nil {
		log.Printf("err: %+v\n", err)
	}
	t := cliMap.TopicPrefix + "/fnfn/SyncReply"
	// TODO
	(*client).Publish(t, 1, false, result)
}

// reply sync req with chan
func ReplySyncReqWithChan(client *mqtt.Client, s *SyncPayload, u []UTXO, cliMap *CliMap, err, end int, send chan bool) {
	reply := SyncReply{}
	reply.Nonce = s.Nonce
	reply.Error = uint8(err)
	if err == 0 || err == 1 {
		tailBlock := block.GetTailBlock()
		reply.BlockHash = tailBlock.Hash
		reply.BlockHeight = tailBlock.Height
		reply.UTXONum = uint16(len(u))
		byteList, _ := UTXOListToByte(&u)
		reply.UTXOList = byteList
		reply.Continue = uint8(end)
	}
	result, errs := StructToBytes(reply)
	if errs != nil {
		log.Printf("err: %+v\n", err)
	}
	t := cliMap.TopicPrefix + "/fnfn/SyncReply"
	token := (*client).Publish(t, 1, false, result)
	if token.Wait() {
		send <- true
	}
}
