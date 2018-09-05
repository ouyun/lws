package mqtt

import (
	"encoding/hex"
	// "encoding/hex"
	"log"
	"testing"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	// "github.com/lomocoin/lws/internal/gateway/crypto"
)

var servicReplyHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("msg: %+v\n", msg.Payload())
	if msg.Topic() == "wqweqwasasqw/fnfn/ServiceReply" {
		s := ServiceReply{}
		DecodePayload(msg.Payload(), &s)
		log.Printf("ServiceReply: %+v\n", s)
	}
}

func TestClient(t *testing.T) {
	p := &Program{
		Id:    "lws",
		isLws: true,
		subs: []string{
			"LWS/lws/ServiceReq",
			"LWS/lws/SyncReq",
			"LWS/lws/UTXOAbort",
			"LWS/lws/SendTxReq",
		},
	}

	err := ClientStart(p)
	if err != nil {
		t.Errorf("run client fail %v", err)
	}
}

func TestInit(t *testing.T) {
	p := &Program{Id: "LWS/lws/ServiceReq", isLws: false}
	if err := p.Init(); err != nil {
		t.Errorf("init client fail %v", err)
	}
}

func TestStart(t *testing.T) {
	p := &Program{Id: "LWS/lws/ServiceReq", isLws: false}
	p.Init()
	// ready1 := make(chan string)
	if err := p.Start(); err != nil {
		t.Errorf("init client fail %v", err)
	}
	time.Sleep(10 * time.Second)
	// <-ready1
	p.Stop()
}

func ClientStart(service Service) error {
	service.Init()
	service.Start()
	return service.Stop()
}

func TestPublish(t *testing.T) {
	var err error
	lws := &Program{
		Id:    "lws",
		isLws: true,
	}
	lws.Init()
	lws.Start()
	cli := &Program{
		Id:    "cli",
		isLws: false,
	}
	cli.Init()
	cli.Start()
	cli.Subscribe("wqweqwasasqw/fnfn/ServiceReply", 0, servicReplyHandler)
	// address, _, _ := crypto.GenerateKeyPair(nil)

	address, _ := hex.DecodeString("6f937c2f5944f5da2a118cebb067cd2c9c92c75955ce05aa05158a1af28e1607")
	// hex.EncodeToString
	// log.Printf("ServiceReply: %+v\n", hex.EncodeToString(address[:]))
	topicPrefix := "wqweqwasasqw" + string(byte(0x00))
	servicePayload := ServicePayload{ //serviceRequ
		Nonce:       uint16(1231),
		Address0:    uint8(1),
		Address:     string(address[:]),
		Version:     uint32(5363),
		TimeStamp:   uint32(time.Now().Unix()),
		ForkNum:     uint8(1),
		ForkList:    RandStringBytesRmndr(32 * 1),
		ReplyUTXON:  uint16(2),
		TopicPrefix: topicPrefix,
		Signature:   RandStringBytesRmndr(64),
	}
	servicMsg, err := GenerateService(servicePayload)
	if err != nil {
		t.Errorf("client publish fail")
	}
	err = cli.Publish("LWS/lws/ServiceReq", 0, false, servicMsg)
	time.Sleep(10 * time.Second)
	cli.Stop()
}

func TestSyncReq(t *testing.T) {
	cli := &Program{
		Id:    "cli",
		isLws: false,
	}
	cli.Init()
	cli.Start()
	cli.Subscribe("wqweqwasasqw/fnfn/SyncReply", 1, servicReplyHandler)
	syncPayload := SyncPayload{ //Sync
		Nonce:     uint16(1231),
		AddressId: uint32(5363),
		ForkID:    RandStringBytesRmndr(32),
		UTXOHash:  RandStringBytesRmndr(32),
		Signature: RandStringBytesRmndr(20),
	}
	syncMsg, err := GenerateService(syncPayload)
	if err != nil {
		t.Errorf("client publish fail")
	}
	err = cli.Publish("LWS/lws/SyncReq", 1, false, syncMsg)
	if err != nil {
		t.Errorf(" client publish fail")
	}
	cli.Stop()
}

func TestUTXOAbort(t *testing.T) {
	cli := &Program{
		Id:    "cli",
		isLws: false,
	}
	cli.Init()
	cli.Start()
	abortPayload := AbortPayload{ //abort
		Nonce:     uint16(1231),
		AddressId: uint32(5363),
		Reason:    uint8(1),
		Signature: RandStringBytesRmndr(20),
	}
	abortMsg, err := GenerateService(abortPayload)
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
		isLws: false,
	}
	cli.Init()
	cli.Start()
	cli.Subscribe("wqweqwasasqw/fnfn/SendTxReply", 1, servicReplyHandler)
	sendTxPayload := SendTxPayload{ //send
		Nonce:     uint16(1231),
		AddressId: uint32(5363),
		ForkID:    RandStringBytesRmndr(32),
		TxData:    RandStringBytesRmndr(20),
		Signature: RandStringBytesRmndr(20),
	}
	sendMsg, err := GenerateService(sendTxPayload)
	if err != nil {
		t.Errorf("client publish fail")
	}
	err = cli.Publish("LWS/lws/SendTxReq", 1, false, sendMsg)
	if err != nil {
		t.Errorf(" client publish fail")
	}
	cli.Stop()
}

func Test(t *testing.T) {
	p := &Program{Id: "LWS/lws/ServiceReq", isLws: false}
	p.Init()
	// ready1 := make(chan string)
	if err := p.Start(); err != nil {
		t.Errorf("init client fail %v", err)
	}
	// time.Sleep(10 * time.Second)
	// <-ready1
	p.Stop()
}

func runClient() {
	lws := &Program{
		Id:    "lws",
		isLws: true,
	}
	lws.Init()
	// ready1 := make(chan string)
	if err := lws.Start(); err != nil {
		log.Printf("init client fail %v", err)
	}
	time.Sleep(10 * time.Second)
	lws.Stop()
}

func TestMain(m *testing.M) {
	// go runClient()
	m.Run()
	log.Printf("there")
}
