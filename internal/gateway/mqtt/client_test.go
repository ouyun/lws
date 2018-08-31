package mqtt

import (
	"testing"
	// "fmt"
	"encoding/hex"
	"time"
	// "bytes"
)

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
		subs: []string{
			"LWS/lws/ServiceReq",
			"LWS/lws/SyncReq",
			"LWS/lws/UTXOAbort",
			"LWS/lws/SendTxReq",
		},
	}
	lws.Init()
	lws.Start()
	cli := &Program{
		Id:    "cli",
		isLws: false,
	}
	cli.Init()
	cli.Start()
	address, _ := hex.DecodeString("7c7080ca76637738a12637d0d96b1b2a7d4d1a823c351c6478333e8f32cf1ca1")
	addressByte := [32]byte{}
	copy(addressByte[:], address)
	servicePayload := ServicePayload{ //serviceRequ
		Nonce:       uint16(1231),
		Address0:    uint8(1),
		Address:     string(addressByte[:]),
		Version:     uint32(5363),
		TimeStamp:   uint32(time.Now().Unix()),
		ForkNum:     uint8(1),
		ForkList:    RandStringBytesRmndr(32 * 1),
		ReplyUTXON:  uint16(2),
		TopicPrefix: "wqweqwasasqw0",
		Signature:   RandStringBytesRmndr(64),
	}
	servicMsg, err := GenerateService(servicePayload)
	if err != nil {
		t.Errorf("client publish fail")
	}
	err = cli.Publish("LWS/lws/ServiceReq", 0, false, servicMsg)
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
	time.Sleep(3 * time.Second)
	// err := Run(p)
	if err != nil {
		t.Errorf(" client publish fail")
	}
}

func Test(t *testing.T) {
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
