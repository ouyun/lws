package mqtt

import (
	"bytes"
	"encoding/hex"
	"log"
	"math"
	"os"

	"github.com/FissionAndFusion/lws/internal/config"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/db/service/block"
	"github.com/FissionAndFusion/lws/internal/gateway/crypto"
	"github.com/eclipse/paho.mqtt.golang"
)

var syncReqHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	// TODO ：
	log.Println("[DEBUG] Received syncReq !")
	var UTXOs []UTXO
	var utxos []UTXO
	var poolUtxos []UTXO
	s := SyncPayload{}
	cliMap := CliMap{}
	pool := GetRedisPool()
	user := model.User{}

	payload := msg.Payload()
	err := DecodePayload(payload, &s)
	if err != nil {
		log.Printf("err: %+v\n", err)
	}
	log.Printf("[DEBUG] Received syncReq from addr [%d]!", s.AddressId)
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
		log.Printf("[INFO] syncReq check address id err: %+v", err)
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 16, 0)
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
		log.Printf("[ERROR] forkID config err: %+v", err)
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 16, 0)
		return
	}
	if bytes.Compare(forkId, s.ForkID) != 0 {
		// 无效分支
		log.Printf("[ERROR] syncReq fork id err: %+v", err)
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
		"LEFT OUTER JOIN utxo_pool "+
		"ON utxo.idx = utxo_pool.idx AND utxo_pool.is_delete = true "+
		"WHERE utxo_pool.is_delete is NULL "+
		"ORDER BY REVERSE(utxo.tx_hash) ASC, utxo.out ASC ", cliMap.Address).Find(&utxos).Error
	if err != nil {
		log.Printf("[ERROR] syncReq query utxo err [%s]", err)
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 16, 0)
		return
	}

	err = connection.Raw("SELECT "+
		"new_utxo.tx_hash AS tx_id, "+
		"new_utxo.out, "+
		"0xffffffff, "+
		"tx_pool.tx_type, "+
		"new_utxo.amount, "+
		"tx_pool.sender AS sender, "+
		"tx_pool.lock_until, "+
		"tx_pool.data "+
		"FROM utxo_pool new_utxo "+
		"INNER JOIN tx_pool "+
		"ON new_utxo.tx_hash = tx_pool.hash "+
		"AND new_utxo.destination = ? "+
		"LEFT OUTER JOIN utxo_pool used_utxo "+
		"ON new_utxo.idx = used_utxo.idx AND used_utxo.is_delete = true "+
		"WHERE new_utxo.is_delete = false AND used_utxo.is_delete is NULL "+
		"ORDER BY REVERSE(new_utxo.tx_hash) ASC, new_utxo.out ASC ", cliMap.Address).Find(&poolUtxos).Error
	if err != nil {
		log.Printf("[ERROR] syncReq query utxo pool err [%s]", err)
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 16, 0)
		return
	}

	log.Printf("[DEBUG] cliMap.Addrsss [%s]", hex.EncodeToString(cliMap.Address))
	log.Printf("[DEBUG] utxo cnt [%d]", len(utxos))
	log.Printf("[DEBUG] utxo pool cnt [%d]", len(poolUtxos))

	for idx, item := range utxos {
		log.Printf("[DEBUG] utxos only syncReply utxo[%d] hash[%s] out[%d]", idx, hex.EncodeToString(item.TXID), item.Out)
	}

	// merge and order utxos
	UTXOs = mergeAndOrderUtxos(utxos, poolUtxos)

	log.Printf("[DEBUG] syncReply utxos list below: ")
	for idx, item := range UTXOs {
		log.Printf("[DEBUG] syncReply utxo[%d] hash[%s] out[%d]", idx, hex.EncodeToString(item.TXID), item.Out)
	}
	// create sync addr chan
	go NewSyncAddrChan(s.AddressId)

	// 计算utxo hash
	utxoHash := UTXOHash(&UTXOs)
	if bytes.Compare(utxoHash, []byte(s.UTXOHash)) == 0 {
		log.Printf("client hash equals local hash!")
		ReplySyncReq(&client, &s, &UTXOs, &cliMap, 0, 0)

		updateRedis(&redisConn, &cliMap, "Nonce", s.Nonce)
		updateRedis(&redisConn, &cliMap, "LwsId", config.GetConfig().INSTANCE_ID)
		return
	}
	// 计算utxo数量
	// 如果utxo 数量超过replyUtxo长度，分多次发送list
	if cliMap.ReplyUTXON < uint16(len(UTXOs)) && cliMap.ReplyUTXON != 0 {
		// 多次发送
		c := make(chan int, 1)
		// 发送次数
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
	updateRedis(&redisConn, &cliMap, "LwsId", config.GetConfig().INSTANCE_ID)
	err = updateRedis(&redisConn, &cliMap, "Nonce", s.Nonce)
	if err != nil {
		log.Printf("save nonce err: %+v\n", err)
	}
}

// reply sync req
func ReplySyncReq(client *mqtt.Client, s *SyncPayload, u *[]UTXO, cliMap *CliMap, err, end int) {
	log.Printf("sending update list !")
	reply := SyncReply{}
	reply.Nonce = s.Nonce
	reply.Error = uint8(err)
	if err == 0 {
		tailBlock := block.GetTailBlockFromDb()
		if tailBlock != nil {
			reply.BlockHash = tailBlock.Hash
			reply.BlockHeight = tailBlock.Height
			reply.BlockTime = tailBlock.Tstamp
		}
		reply.UTXONum = uint16(0)
		reply.Continue = uint8(end)
	}
	if err == 1 {
		tailBlock := block.GetTailBlockFromDb()
		if tailBlock != nil {
			reply.BlockHash = tailBlock.Hash
			reply.BlockHeight = tailBlock.Height
			reply.BlockTime = tailBlock.Tstamp
		}
		reply.UTXONum = uint16(len(*u))
		byteList, _ := UTXOListToByte(u)
		reply.UTXOList = byteList
		reply.Continue = uint8(end)
	}
	result, errs := StructToBytes(reply)
	if errs != nil {
		log.Printf("err: %+v\n", err)
	}
	log.Printf("[DEBUG] send syncReply addr [%d]", cliMap.AddressId)
	t := cliMap.TopicPrefix + "/fnfn/SyncReply"
	token := (*client).Publish(t, 1, false, result)
	if token.Wait() {
		tokenErr := token.Error()
		if tokenErr != nil {
			log.Printf("[ERROR] publish err: %+v", tokenErr)
		} else {
			log.Printf("[DEBUG] done send syncReply addr [%d] done", cliMap.AddressId)
		}
	}
}

// reply sync req with chan
func ReplySyncReqWithChan(client *mqtt.Client, s *SyncPayload, u []UTXO, cliMap *CliMap, err, end int, send chan int) {
	log.Printf("send update list with chan !")
	reply := SyncReply{}
	reply.Nonce = s.Nonce
	reply.Error = uint8(err)
	if err == 0 || err == 1 {
		tailBlock := block.GetTailBlockFromDb()
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
	// for {
	// 	if token.Wait() && token.Error() == nil {
	// 		send <- 1
	// 		log.Printf("err: %+v\n", token.Error())
	// 		break
	// 	}
	// }

	if token.Wait() {
		send <- 1
		tokenErr := token.Error()
		if tokenErr != nil {
			log.Printf("[ERROR] publish err: %+v", tokenErr)
		} else {
			log.Printf("[DEBUG] done send syncReply addr [%d] done", cliMap.AddressId)
		}
	}
}

func mergeAndOrderUtxos(a []UTXO, b []UTXO) []UTXO {
	aLen := len(a)
	bLen := len(b)
	totalUtxos := make([]UTXO, aLen+bLen)
	j := 0
	i := 0
	pos := 0
	for ; i < aLen; i++ {
		shouldLastInsert := true
		for j < bLen {
			aItem := a[i]
			bItem := b[j]
			aBytes := reverseBytes(aItem.TXID)
			bBytes := reverseBytes(bItem.TXID)
			compRes := bytes.Compare(aBytes, bBytes)
			if compRes == 0 {
				compRes = int(aItem.Out - bItem.Out)
			}

			if compRes < 0 {
				log.Printf("[DEBUG] insert a item i[%d] j[%d] idx[%d]", i, j, pos)
				totalUtxos[pos] = aItem
				pos += 1
				shouldLastInsert = false
				break
			} else {
				log.Printf("[DEBUG] insert b item i[%d] j[%d] idx[%d]", i, j, pos)
				totalUtxos[pos] = bItem
				pos += 1
				j += 1
			}
		}

		if shouldLastInsert {
			log.Printf("[DEBUG] last insert a item i[%d] j[%d] idx[%d]", i, j, pos)
			totalUtxos[pos] = a[i]
			pos += 1
		}
	}

	for ; j < bLen; j++ {
		log.Printf("[DEBUG] append insert b item j[%d] idx[%d]", j, pos)
		totalUtxos[pos] = b[j]
		pos += 1
	}
	return totalUtxos
}
