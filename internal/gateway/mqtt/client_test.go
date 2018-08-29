package mqtt

import (
	"testing"
	// "fmt"
	"time"
	// "bytes"
)

func TestClient(t *testing.T) {
	p := &Program{Id: "LWS/lws/ServiceReq", isLws: false}
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
	p := &Program{Id: "LWS/lws/ServiceReq", isLws: false}
	// err := ClientStart(p)
	p.Init()
	p.Start()
	payload := ServicePayload{
			Nonce: uint16(1231),
			Address: RandStringBytesRmndr(33),
			Version: uint32(5363),
			TimeStamp: uint32(time.Now().Unix()),
			ForkNum:  uint8(1),
			ForkList: RandStringBytesRmndr(32*1),
			ReplyUTXON: uint16(2),
			TopicPrefix: "DE0",
		}
	msg, err := GeneratePayload(payload)
	if err != nil {
		t.Errorf("client publish fail", )
	}
	err = p.Publish("LWS/lws/ServiceReq", msg)

	time.Sleep(10 * time.Second)
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
