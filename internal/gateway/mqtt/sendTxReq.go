package mqtt

import (
	"bytes"
	"encoding/hex"
	"log"
	"os"
	"strconv"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/lomocoin/lws/internal/coreclient"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/internal/db/model"
	"github.com/lomocoin/lws/internal/db/service"
	"github.com/lomocoin/lws/internal/gateway/crypto"
)

type UTXOIndex struct {
	TXID []byte `len:"32"`
	Out  uint8  `len:"1"`
}

var sendTxReqReqHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	s := SendTxPayload{}
	payload := msg.Payload()
	cliMap := CliMap{}
	user := model.User{}
	err := DecodePayload(payload, &s)
	// log.Printf("SendTxPayload: %+v\n", s)
	if err != nil {
		log.Printf("err: %+v\n", err)
		return
	}
	// 连接 redis
	pool := NewRedisPool()
	redisConn := pool.Get()
	connection := db.GetConnection()
	defer redisConn.Close()

	inRedis, inDb, err := CheckAddressId(s.AddressId, connection, &redisConn, &user, &cliMap)
	// 验证签名
	signed := crypto.SignWithApiKey(cliMap.ApiKey, payload[:len(payload)-20])
	if bytes.Compare(signed, payload[len(payload)-20:]) != 0 {
		// 丢弃 内容
		return
	}
	if err != nil {
		ReplySendTx(&client, &s, 16, 0, "")
		return
	}
	// 无效addressId
	if !inRedis && !inDb {
		ReplySendTx(&client, &s, 1, 0, "")
		return
	}

	// 验证分支
	forkId, err := hex.DecodeString(os.Getenv("Fork_Id"))
	if err != nil {
		log.Printf("err: %+v", err)
	}
	if bytes.Compare(forkId, s.ForkID) != 0 {
		ReplySendTx(&client, &s, 2, 0, "")
		return
	}

	// get txdata
	var txData TxData
	err = TxDataToStruct(s.TxData, &txData)
	if err != nil {
		// fail
		ReplySendTx(&client, &s, 16, 0, "")
		return
	}

	// get amount
	amount, _, err := service.GetUtxoSummary(getUtxoIndex(&txData.UtxoIndex), connection)
	if err != nil {
		ReplySendTx(&client, &s, 16, 0, "")
		return
	}

	//校验 tx amount
	balance := txData.NAmount - amount - txData.NTxFee
	if balance < 0 {
		// return fail
		ReplySendTx(&client, &s, 1, 4, "")
		return
	}

	// 验证打包费
	txFee, err := strconv.ParseInt(os.Getenv("TxFee"), 10, 64)
	if txData.NTxFee != txFee {
		ReplySendTx(&client, &s, 1, 4, "")
		return
	}

	// TODO：send tx
	result, err := SendTxToCore(StartCoreClient(), &s)
	if err != nil {
		ReplySendTx(&client, &s, 16, 0, "")
		return
	}
	if result.Result == "failed" {
		ReplySendTx(&client, &s, 3, 0, result.Reason)
		return
	}
	ReplySendTx(&client, &s, 0, 0, "")
	return
}

// reply send tx
func ReplySendTx(client *mqtt.Client, s *SendTxPayload, err int, errCode int, errDesc string) {
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
	t := "/fnfn/SendTxReply"
	// TODO
	(*client).Publish(t, 1, false, result)
}

func SendTxToCore(client *coreclient.Client, s *SendTxPayload) (resultMessage *lws.SendTxRet, err error) {
	response, err := client.Call(&lws.SendTxArg{
		Data: s.TxData,
	})
	if err != nil {
		return resultMessage, err
	}
	result, _ := response.(*dbp.Result)
	resultMessage = interface{}(result.GetResult()[0]).(*lws.SendTxRet)
	return resultMessage, err
}

func StartCoreClient() *coreclient.Client {
	addr := os.Getenv("CORECLIENT_URL")

	log.Printf("Connect to core client [%s]", addr)
	client := coreclient.NewTCPClient(addr)

	client.Start()
	return client
}

func getUtxoIndex(index *[]byte) []*model.Utxo {
	var utxo []*model.Utxo
	for i := 0; i < (len(*index) / 33); i++ {
		utxo[i].TxHash = (*index)[(i * 33):(((i + 1) * 33) - 1)]
		utxo[i].Out = uint8((*index)[((i+1)*33)-1])
	}
	return utxo
}
