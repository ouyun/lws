package mqtt

import (
	"bytes"
	"encoding/hex"
	"log"
	"math"
	"os"

	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/db/service/block"
	"github.com/FissionAndFusion/lws/internal/gateway/crypto"
	"github.com/eclipse/paho.mqtt.golang"
)

var syncReqHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	// TODO ：
	log.Println("Received syncReq !")
	var UTXOs []UTXO
	s := SyncPayload{}
	cliMap := CliMap{}
	pool := GetRedisPool()
	user := model.User{}

	payload := msg.Payload()
	err := DecodePayload(payload, &s)
	if err != nil {
		log.Printf("err: %+v\n", err)
	}
	// 连接 redis
	redisConn := pool.Get()
	connection := db.GetConnection()
	defer redisConn.Close()
	if redisConn.Err() != nil {
		log.Printf("redisConn: \n")
	}

	inRedis, inDb, err := CheckAddressId(s.AddressId, connection, &redisConn, &user, &cliMap)
	// 验证签名
	signed := crypto.SignWithApiKey(cliMap.ApiKey, payload[:len(payload)-20])
	if bytes.Compare(signed, s.Signature) != 0 {
		// 丢弃 请求
		log.Printf("verify failed : \n")
		return
	}
	if err != nil {
		log.Printf("err: %+v", err)
		// ReplySyncReq(&client, &s, &UTXOs, &cliMap, 16, 0)
		return
	}
	if !inRedis && !inDb {
		log.Printf("err: %+v", err)
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 2, 0)
		return
	}
	// 检查分支
	forkId, err := hex.DecodeString(os.Getenv("FORK_ID"))
	if err != nil {
		// 内部错误
		log.Printf("err: %+v", err)
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 16, 0)
		return
	}
	if bytes.Compare(forkId, s.ForkID) != 0 {
		// 无效分支
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 3, 0)
		return
	}
	//get utxo list

	err = connection.Raw("SELECT "+
		"utxo.tx_hash AS tx_id, "+
		"utxo.out, "+
		"utxo.block_height, "+
		"tx.tx_type, "+
		"utxo.amount, "+
		"tx.sender AS sender, "+
		"tx.lock_until, "+
		"tx.data "+
		"FROM utxo "+
		"INNER JOIN tx "+
		"ON utxo.tx_hash = tx.hash "+
		"AND utxo.destination = ? "+
		"ORDER BY utxo.tx_hash ASC, utxo.out ASC ", cliMap.Address).Find(&UTXOs).Error
	if err != nil {
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 16, 0)
		return
	}
	log.Printf("utxos：%+v", UTXOs)
	// 计算utxo hash
	utxoHash := UTXOHash(&UTXOs)
	log.Printf("get UTXOs %+v\n", UTXOs)
	if bytes.Compare(utxoHash, []byte(s.UTXOHash)) == 0 {
		log.Printf("client hash equals local hash!")
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
		c := make(chan int, 1)
		times := int(math.Ceil(float64(uint16(len(UTXOs)) / cliMap.ReplyUTXON)))
		for index := 0; index < times; index++ {
			log.Printf("total：%+v, send utxo in data pack %d\n", times, index)
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
		log.Printf("send utxo in one data pack \n")
		// 一次发送
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 1, 0)
	}
	// save nonce
	updateRedis(&redisConn, &cliMap, "Nonce", s.Nonce)
}

// reply sync req
func ReplySyncReq(client *mqtt.Client, s *SyncPayload, u *[]UTXO, cliMap *CliMap, err, end int) {
	log.Printf("sending update list !")
	reply := SyncReply{}
	reply.Nonce = s.Nonce
	reply.Error = uint8(err)
	if err == 0 {
		tailBlock := block.GetTailBlock()
		reply.BlockHash = tailBlock.Hash
		reply.BlockHeight = tailBlock.Height
		reply.BlockTime = tailBlock.Tstamp
		reply.UTXONum = uint16(0)
		reply.Continue = uint8(end)
	}
	if err == 1 {
		tailBlock := block.GetTailBlock()
		reply.BlockHash = tailBlock.Hash
		reply.BlockHeight = tailBlock.Height
		reply.BlockTime = tailBlock.Tstamp
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
	token := (*client).Publish(t, 1, false, result)
	if token.Wait() {
		log.Printf("publish err: %+v", token.Error())
	}
}

// reply sync req with chan
func ReplySyncReqWithChan(client *mqtt.Client, s *SyncPayload, u []UTXO, cliMap *CliMap, err, end int, send chan int) {
	log.Printf("send update list with chan !")
	reply := SyncReply{}
	reply.Nonce = s.Nonce
	reply.Error = uint8(err)
	if err == 0 || err == 1 {
		tailBlock := block.GetTailBlock()
		reply.BlockHash = tailBlock.Hash
		reply.BlockHeight = tailBlock.Height
		reply.BlockTime = tailBlock.Tstamp
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
	for {
		if token.Wait() && token.Error() == nil {
			send <- 1
			log.Printf("err: %+v\n", token.Error())
			break
		}
	}
}
