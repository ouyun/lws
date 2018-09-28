package mqtt

import (
	// "encoding/hex"
	// "flag"
	"log"
	// "os"
	"testing"

	"github.com/eclipse/paho.mqtt.golang"
	// "github.com/FissionAndFusion/lws/internal/db"
	// "github.com/FissionAndFusion/lws/test/helper"
)

var servicReplyHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("get reply:  %+v\n", msg.Payload())
	// if msg.Topic() == "wqweqwasasqw/fnfn/ServiceReply" {
	// DecodePayload(msg.Payload(), &s)
	// log.Printf("ServiceReply: %+v\n", s)
	// }
}

func TestClient(t *testing.T) {
	p := &Program{
		Id:    "lws",
		IsLws: false,
	}
	err := ClientStart(p)
	if err != nil {
		t.Errorf("run client fail %v", err)
	}
}

func TestStart(t *testing.T) {
	p := &Program{Id: "LWS/lws/ServiceReq", IsLws: true}
	p.Init()
	if err := p.Start(); err != nil {
		t.Errorf("init client fail %v", err)
	}
	err := p.Subscribe("wqweqwasasqw/fnfn/SendTxReply", 1, servicReplyHandler)
	if err != nil {
		t.Errorf("Subscribe client fail %v", err)
	}
	p.Stop()
}

func ClientStart(service Service) error {
	service.Init()
	if err := service.Start(); err != nil {
		return err
	}
	return service.Stop()
}

func TestUTXOAbort(t *testing.T) {
	cli := &Program{
		Id:    "cli",
		IsLws: false,
	}
	cli.Init()
	if err := cli.Start(); err != nil {
		t.Errorf("client start failed")
	}
	abortPayload := AbortPayload{ //abort
		Nonce:     uint16(1231),
		AddressId: uint32(5363),
		Reason:    uint8(1),
		Signature: []byte(RandStringBytesRmndr(20)),
	}
	abortMsg, err := StructToBytes(abortPayload)
	if err != nil {
		t.Errorf("client publish fail")
	}
	err = cli.Publish("LWS/lws/UTXOAbort", 1, false, abortMsg)
	if err != nil {
		t.Errorf(" client publish fail")
	}
	cli.Stop()
}

func TestSendTxReq(t *testing.T) {
	cli := &Program{
		Id:    "cli",
		IsLws: false,
	}
	cli.Init()
	if err := cli.Start(); err != nil {
		t.Errorf("client start failed")
	}
	cli.Subscribe("wqweqwasasqw/fnfn/SendTxReply", 1, servicReplyHandler)
	sendTxPayload := SendTxPayload{ //send
		Nonce:     uint16(1231),
		AddressId: uint32(5363),
		ForkID:    []byte(RandStringBytesRmndr(32)),
		TxData:    []byte(RandStringBytesRmndr(20)),
		Signature: []byte(RandStringBytesRmndr(20)),
	}
	sendMsg, err := StructToBytes(sendTxPayload)
	if err != nil {
		t.Errorf("client publish fail")
	}
	err = cli.Publish("LWS/lws/SendTxReq", 1, false, sendMsg)
	if err != nil {
		t.Errorf(" client publish fail")
	}
	cli.Stop()
}

// func TestMain(m *testing.M) {
// 	// helper.ResetDb()
// 	connection := db.GetConnection()
// 	connection.LogMode(true)
// 	flag.Parse()
// 	c := make(chan int, 1)
// 	go func() {
// 		lws := &Program{
// 			Id:    "lws",
// 			isLws: false,
// 		}
// 		lws.Init()
// 		if err := lws.Start(); err != nil {
// 			log.Printf("init client fail %v", err)
// 			return
// 		}
// 		c <- 1
// 		err := lws.Stop()
// 		if err != nil {
// 			log.Printf("stop client fail %v", err)
// 			return
// 		}
// 	}()
// 	code := m.Run()
// 	<-c
// 	connection.Close()
// 	os.Exit(code)
// }
