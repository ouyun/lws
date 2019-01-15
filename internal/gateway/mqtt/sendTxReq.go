package mqtt

import (
	"bytes"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/FissionAndFusion/lws/internal/coreclient"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/db/service/utxo"
	"github.com/FissionAndFusion/lws/internal/gateway/crypto"
	"github.com/FissionAndFusion/lws/test/helper"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/protobuf/ptypes"
)

type UTXOIndex struct {
	TXID []byte `len:"32"`
	Out  uint8  `len:"1"`
}

var sendTxReqReqHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	workerPool.SendTxReqChan <- &ClientMsg{client: &client, msg: &msg}
}

func sendTxReqWorkerHandler(clientMsg *ClientMsg) {
	client := *clientMsg.client
	msg := *clientMsg.msg

	log.Printf("[DEBUG] Received sendTxReq msgId: [%d]!", msg.MessageID())
	s := SendTxPayload{}
	payload := msg.Payload()
	cliMap := CliMap{}
	user := model.User{}
	err := DecodePayload(payload, &s)
	if err != nil {
		log.Printf("err: %+v\n", err)
		return
	}
	log.Printf("[DEBUG] SendTxPayload addressId [%d] ", s.AddressId)
	log.Printf("[DEBUG] SendTxReq received txdata[%s]", hex.EncodeToString(s.TxData))
	defer helper.MeasureTime(helper.MeasureTitle("handle sendTxReq addressId [%d]", s.AddressId))
	// 连接 redis
	pool := GetRedisPool()
	redisConn := pool.Get()
	connection := db.GetConnection()
	defer redisConn.Close()

	inRedis, inDb, err := CheckAddressId(s.AddressId, connection, &redisConn, &user, &cliMap)
	// 验证签名
	signed := crypto.SignWithApiKey(cliMap.ApiKey, payload[:len(payload)-20])
	if bytes.Compare(signed, s.Signature) != 0 {
		// 丢弃 内容
		log.Printf("[ERROR] verify failed ！discard request！ ")
		return
	}
	if err != nil {
		ReplySendTx(&client, &s, 16, 0, "", &cliMap)
		return
	}
	// 无效addressId
	if !inRedis && !inDb {
		ReplySendTx(&client, &s, 1, 0, "", &cliMap)
		return
	}

	// 验证分支
	forkId, err := hex.DecodeString(os.Getenv("FORK_ID"))
	if err != nil {
		log.Printf("[ERROR] decode fork id err: %+v", err)
	}
	if bytes.Compare(forkId, s.ForkID) != 0 {
		ReplySendTx(&client, &s, 2, 0, "", &cliMap)
		return
	}

	// get txdata
	var txData TxData
	err = TxDataToStruct(s.TxData, &txData)
	if err != nil {
		// fail
		log.Printf("[ERROR] invalid txdata struct [%s]", err)
		ReplySendTx(&client, &s, 16, 0, "", &cliMap)
		return
	}

	log.Printf("[DEBUG] sendTxReq ready to check balance txdata[%s]", hex.EncodeToString(s.TxData))

	log.Printf("[DEBUG] addressId[%d] get utxo summery start", s.AddressId)
	// get amount
	amount, cnt, err := utxo.GetSummary(getUtxoIndex(&txData.UtxoIndex), connection)
	if err != nil {
		log.Printf("[INFO] get utxo summary failed [%s]", err)
		ReplySendTx(&client, &s, 16, 0, "", &cliMap)
		return
	}
	log.Printf("[DEBUG] addressId[%d] get utxo summary [%d] utxocnt[%d]", s.AddressId, amount, cnt)
	log.Printf("[DEBUG] addressId[%d] NAmount [%d] NTxFee[%d]", s.AddressId, txData.NAmount, txData.NTxFee)

	//校验 tx amount
	balance := amount - txData.NAmount - txData.NTxFee
	if balance < 0 {
		// return fail
		ReplySendTx(&client, &s, 4, 0, "balance err", &cliMap)
		log.Printf("[INFO] balance do not enough")
		return
	}

	// 验证打包费
	txFee, err := strconv.ParseInt(os.Getenv("TX_FEE"), 10, 64)
	if txData.NTxFee < txFee {
		ReplySendTx(&client, &s, 4, 0, "txFee err", &cliMap)
		log.Printf("[INFO] txFee do not enough")
		return
	}

	log.Printf("[DEBUG] send tx to core wallet addressId[%d]", s.AddressId)
	log.Printf("[DEBUG] sendTxReq ready to send core txdata[%s]", hex.EncodeToString(s.TxData))

	// TODO：send tx
	coreClient := StartCoreClient()
	// defer coreClient.Stop()
	result, err := SendTxToCore(coreClient, &s)
	log.Printf("[DEBUG] sendTxReq received core reply txdata[%s]", hex.EncodeToString(s.TxData))
	if err != nil {
		ReplySendTx(&client, &s, 16, 0, "", &cliMap)
		log.Printf("[INFO] txdata[%s] prefix[%s] send to corewallet err : %+v", hex.EncodeToString(s.TxData), cliMap.TopicPrefix, err)
		return
	}
	if result.Result == "failed" {
		ReplySendTx(&client, &s, 3, 0, result.Reason, &cliMap)
		log.Printf("[INFO] txdata[%s] prefix[%s] sendtx corewallet error : %+v", hex.EncodeToString(s.TxData), cliMap.TopicPrefix, result)
		return
	}
	ReplySendTx(&client, &s, 0, 0, "", &cliMap)
	log.Printf("[DEBUG] sendTxReq send reply done to iot txdata[%s]", hex.EncodeToString(s.TxData))
	log.Printf("[DEBUG] send tx success addressId[%d]!", s.AddressId)
	return
}

// reply send tx
func ReplySendTx(client *mqtt.Client, s *SendTxPayload, err int, errCode int, errDesc string, cliMap *CliMap) {
	t := cliMap.TopicPrefix + "/fnfn/SendTxReply"
	defer helper.MeasureTime(helper.MeasureTitle("%s txData[%s]", t, hex.EncodeToString(s.TxData)))
	reply := SendTxReply{}
	reply.Nonce = s.Nonce
	reply.Error = uint8(err)
	if err == 3 || err == 4 {
		reply.ErrCode = uint8(errCode)
		reply.ErrDesc = errDesc + string(byte(0x00))
	}
	result, errs := StructToBytes(reply)
	if errs != nil {
		log.Printf("err: %+v\n", err)
	}
	// TODO
	token := (*client).Publish(t, 1, false, result)
	token.Wait()
	tokenErr := token.Error()
	if tokenErr != nil {
		log.Printf("[ERROR] [%s] error: %s", t, tokenErr)
	} else {
		log.Printf("[DEBUG] [%s] done msgId [%d]", t, token.(*mqtt.PublishToken).MessageID())
	}
}

func SendTxToCore(client *coreclient.Client, s *SendTxPayload) (resultMessage *lws.SendTxRet, err error) {
	defer helper.MeasureTime(helper.MeasureTitle("handle send tx to core"))
	params := &lws.SendTxArg{
		Data: s.TxData,
	}
	// log.Printf("SendTxArg data: %+v", hex.EncodeToString(s.TxData))
	serializedParams, err := ptypes.MarshalAny(params)
	if err != nil {
		log.Fatal("[ERROR] could not serialize any field")
		return nil, err
	}

	method := &dbp.Method{
		Method: "sendtransaction",
		Params: serializedParams,
	}
	if err != nil {
		return resultMessage, err
	}

	log.Printf("build sendTxReq args done addrId[%d]", s.AddressId)
	response, err := client.Call(method)
	if err != nil {
		log.Printf("[ERROR] sendTx failed, get err: %+v \n", err)
		return nil, err
	}
	log.Printf("sendTxReq got result addrId[%d]", s.AddressId)
	result, ok := response.(*dbp.Result)
	if !ok {
		log.Println("[ERROR] unsuport dbp type")
		err = errors.New("response did not match dbp type")
		return nil, err
	}

	sendTxResponse := &lws.SendTxRet{}
	err = ptypes.UnmarshalAny(result.GetResult()[0], sendTxResponse)
	if err != nil {
		log.Printf("[ERROR] unmashall result error [%s] \n", err)
		return nil, err
	}

	return sendTxResponse, err
}

func getUtxoIndex(index *[]byte) []*model.Utxo {
	legnth := (len(*index) / 33)
	utxos := make([]*model.Utxo, legnth)
	log.Printf("[DEBUG] utxo index: %+v", index)
	log.Printf("[DEBUG] utxo length : %+v", len(*index))
	// TODO: array bound check
	for i := 0; i < (len(*index) / 33); i++ {
		ut := &model.Utxo{}
		ut.Out = uint8((*index)[(i*33)+32])
		ut.TxHash = (*index)[(i * 33) : (i*33)+32]
		log.Printf("[DEBUG] : utxo hash [%s] out[%d]", hex.EncodeToString(ut.TxHash), ut.Out)
		log.Printf("[DEBUG] : utxo hash [%v]", ut.TxHash)
		utxos[i] = ut
	}
	return utxos
}
