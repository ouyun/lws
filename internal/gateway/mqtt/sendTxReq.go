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
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/golang/protobuf/ptypes"
)

type UTXOIndex struct {
	TXID []byte `len:"32"`
	Out  uint8  `len:"1"`
}

var sendTxReqReqHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Println("Received sendTxReq !")
	s := SendTxPayload{}
	payload := msg.Payload()
	cliMap := CliMap{}
	user := model.User{}
	err := DecodePayload(payload, &s)
	log.Printf("SendTxPayload: %+v\n", s)
	if err != nil {
		log.Printf("err: %+v\n", err)
		return
	}
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
		log.Printf("verify failed ！discard request！ ")
		return
	}
	if err != nil {
		ReplySendTx(&client, &s, 16, 0, "", &cliMap)
		log.Printf("sendReq16 internal error SignApiKey: [%s]", err)
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
		log.Printf("err: %+v", err)
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
		ReplySendTx(&client, &s, 16, 0, "", &cliMap)
		log.Printf("sendReq16 internal error TxDataToStruct: [%s]", err)
		return
	}

	// get amount
	amount, _, err := utxo.GetSummary(getUtxoIndex(&txData.UtxoIndex), connection)
	if err != nil {
		ReplySendTx(&client, &s, 16, 0, "", &cliMap)
		log.Printf("sendReq16 internal error GetSummary: [%s]", err)
		return
	}

	//校验 tx amount
	balance := txData.NAmount - amount - txData.NTxFee
	if balance < 0 {
		// return fail
		ReplySendTx(&client, &s, 4, 0, "balance err", &cliMap)
		log.Printf("balance do not enough")
		return
	}

	// 验证打包费
	txFee, err := strconv.ParseInt(os.Getenv("TX_FEE"), 10, 64)
	if txData.NTxFee < txFee {
		ReplySendTx(&client, &s, 4, 0, "txFee err", &cliMap)
		log.Printf("txFee do not enough")
		return
	}

	// TODO：send tx
	coreClient := StartCoreClient()
	// defer coreClient.Stop()
	result, err := SendTxToCore(coreClient, &s)
	if err != nil {
		ReplySendTx(&client, &s, 16, 0, "", &cliMap)
		log.Printf("sendReq16 internal error SendTxToCore: [%s]", err)
		log.Printf("err : %+v", err)
		return
	}
	if result.Result == "failed" {
		ReplySendTx(&client, &s, 3, 0, result.Reason, &cliMap)
		log.Printf("result : %+v", result)
		return
	}
	ReplySendTx(&client, &s, 0, 0, "", &cliMap)
	log.Printf("send tx success !")
	return
}

// reply send tx
func ReplySendTx(client *mqtt.Client, s *SendTxPayload, err int, errCode int, errDesc string, cliMap *CliMap) {
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
	t := cliMap.TopicPrefix + "/fnfn/SendTxReply"
	// TODO
	(*client).Publish(t, 1, false, result)
}

func SendTxToCore(client *coreclient.Client, s *SendTxPayload) (resultMessage *lws.SendTxRet, err error) {
	params := &lws.SendTxArg{
		Data: s.TxData,
	}
	log.Printf("SendTxArg data: %+v", hex.EncodeToString(s.TxData))
	serializedParams, err := ptypes.MarshalAny(params)
	if err != nil {
		log.Fatal("could not serialize any field")
		return nil, err
	}

	method := &dbp.Method{
		Method: "sendtransaction",
		Params: serializedParams,
	}
	if err != nil {
		return resultMessage, err
	}
	response, err := client.Call(method)
	if err != nil {
		log.Printf("sendTx failed, get err: %+v \n", err)
		return nil, err
	}
	result, ok := response.(*dbp.Result)
	if !ok {
		log.Println("unsuport dbp type")
		err = errors.New("response did not match dbp type")
		return nil, err
	}

	sendTxResponse := &lws.SendTxRet{}
	err = ptypes.UnmarshalAny(result.GetResult()[0], sendTxResponse)
	if err != nil {
		log.Printf("unmashall result error [%s] \n", err)
		return nil, err
	}

	return sendTxResponse, err
}

func getUtxoIndex(index *[]byte) []*model.Utxo {
	legnth := (len(*index) / 33)
	utxos := make([]*model.Utxo, legnth)
	log.Printf("index: %+v", index)
	log.Printf("length : %+v", len(*index))
	// TODO: array bound check
	for i := 0; i < (len(*index) / 33); i++ {
		log.Printf("i: %+v", i)
		ut := &model.Utxo{}
		ut.Out = uint8((*index)[(i * 33)])
		ut.TxHash = (*index)[((i * 33) + 1) : ((i+1)*33)-1]
		utxos[i] = ut
	}
	return utxos
}
