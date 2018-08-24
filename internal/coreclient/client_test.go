package coreclient

import (
	// "github.com/lomocoin/soolws/client"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	dbp "github.com/lomocoin/lws/internal/coreclient/DBPMsg/go"
	"testing"
)

func TestGenerateMsgId(t *testing.T) {
	c := &CoreClient{}
	id := c.generateMsgId()
	if id == "" {
		t.Error("generated msg id empty string")
	}
}

func TestPack(t *testing.T) {
	c := &CoreClient{}
	var err error

	data := &dbp.Connect{
		Version: 12,
		Session: "testsession",
		Client:  "lws",
	}

	msgBytes, err := c.Pack(data)
	if err != nil {
		t.Error("pack failed", err)
	}

	fmt.Println("msg", string(msgBytes[:]))

	baseMsg := &dbp.Base{}
	err = proto.Unmarshal(msgBytes, baseMsg)
	if err != nil {
		t.Error("unpack failed", err)
	}

	fmt.Println("baseMsg: ", baseMsg)

	unpackedConnect := &dbp.Connect{}

	err = ptypes.UnmarshalAny(baseMsg.Object, unpackedConnect)
	if err != nil {
		t.Error("unpack Object failed", err)
	}

	fmt.Printf("unpackedConnect = %+v\n", unpackedConnect)
}

func TestPack2(t *testing.T) {
	c := &CoreClient{}
	var err error

	data := &dbp.Ping{}

	msgBytes, err := c.Pack(data)
	if err != nil {
		t.Error("pack failed", err)
	}

	fmt.Println("msg", string(msgBytes[:]))

	baseMsg := &dbp.Base{}
	err = proto.Unmarshal(msgBytes, baseMsg)
	if err != nil {
		t.Error("unpack failed", err)
	}

	fmt.Println("baseMsg: ", baseMsg)

	object := &dbp.Ping{}

	err = ptypes.UnmarshalAny(baseMsg.Object, object)
	if err != nil {
		t.Error("unpack Object failed", err)
	}

	fmt.Printf("object = %+v\n", object)

}
