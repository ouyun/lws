package mqtt

import (
	"encoding/hex"
	"log"
	"testing"

	"github.com/golang/protobuf/ptypes"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
)

func TestSendTxReqM(t *testing.T) {
	str := "01000000000000005993f754769f516f9c5a185227091f1f31e21bf653947024417d8df85edcd1e3010614d0a788afe75da0981cd474b8e8d7c3bb669cdcc9df24ea5578a08a6339c20101d93f1b85e91ced486bb96c04dfcc8c3ff231f0522838ff70bbd5ac8a6f8b6c0940420f0000000000640000000000000000810eedd2ff9edde5c8317a66817e4c52191cad131f9bd611dd998ea95779d061ac01d93f1b85e91ced486bb96c04dfcc8c3ff231f0522838ff70bbd5ac8a6f8b6c0945265712868a742deed0aa6f738a05f06002141874b88d34a3bb0c5bb4e74b2496de86d272f2475cfa8de58256ce392249b42e34e35776a2761fb694c853150c"
	txda, err := hex.DecodeString(str)
	client := StartCoreClient()
	params := &lws.SendTxArg{
		Data: txda,
	}
	serializedParams, err := ptypes.MarshalAny(params)
	if err != nil {
		log.Fatal("could not serialize any field")
	}

	method := &dbp.Method{
		Method: "sendtransaction",
		Params: serializedParams,
	}

	response, err := client.Call(method)
	if err != nil {
		log.Printf("err: %+v", err)
	}

	log.Printf("response: %+v", response)
	result, ok := response.(*dbp.Result)

	log.Printf("result: %+v", result)
	if !ok {
		t.Error("type did not suport")
	}

	sendTx := &lws.SendTxRet{}
	err = ptypes.UnmarshalAny(result.GetResult()[0], sendTx)
	if err != nil {
		log.Printf("unmashall result error [%s]", err)
	}
	log.Printf("get result  %+v", sendTx)

}
