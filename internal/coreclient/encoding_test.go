package coreclient

import (
	// "github.com/lomocoin/soolws/client"
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"
	"testing"
)

func TestPackMsgWithoutId(t *testing.T) {
	encoder := &messageEncoder{}
	var err error

	data := &dbp.Connect{
		Version: 12,
		Session: "testsession",
		Client:  "lws",
	}

	msgBytes, err := encoder.PackMsg(data, "")
	if err != nil {
		t.Error("pack failed", err)
	}

	fmt.Println("msgBytes", msgBytes[:])

	baseMsg := &dbp.Base{}
	err = proto.Unmarshal(msgBytes, baseMsg)
	if err != nil {
		t.Error("unpack failed", err)
	}

	unpackedConnect := &dbp.Connect{}

	err = ptypes.UnmarshalAny(baseMsg.Object, unpackedConnect)
	if err != nil {
		t.Error("unpack Object failed", err)
	}

	fmt.Printf("unpackedConnect = %+v\n", unpackedConnect)

	if unpackedConnect.Version != 12 {
		t.Error("expect Version to be 12, but ", unpackedConnect.Version)
	}
	if unpackedConnect.Session != "testsession" {
		t.Error("expect Session to be testsession, but ", unpackedConnect.Session)
	}
	if unpackedConnect.Client != "lws" {
		t.Error("expect Client to be lws, but ", unpackedConnect.Client)
	}

	expectMsgBytes := []byte{18, 55, 10, 31, 116, 121, 112, 101, 46, 103, 111, 111, 103, 108, 101, 97, 112, 105, 115, 46, 99, 111, 109, 47, 100, 98, 112, 46, 67, 111, 110, 110, 101, 99, 116, 18, 20, 10, 11, 116, 101, 115, 116, 115, 101, 115, 115, 105, 111, 110, 16, 12, 26, 3, 108, 119, 115}

	if bytes.Compare(msgBytes, expectMsgBytes) != 0 {
		t.Error("pack msg bytes doesn't match")
	}
}

func TestPackMsgWithId(t *testing.T) {
	encoder := &messageEncoder{}
	var err error
	expectId := "helloIdId"

	data := &dbp.Ping{}

	msgBytes, err := encoder.PackMsg(data, expectId)
	if err != nil {
		t.Error("pack failed", err)
	}

	baseMsg := &dbp.Base{}
	err = proto.Unmarshal(msgBytes, baseMsg)
	if err != nil {
		t.Error("unpack failed", err)
	}

	unpackedPing := &dbp.Ping{}

	err = ptypes.UnmarshalAny(baseMsg.Object, unpackedPing)
	if err != nil {
		t.Error("unpack Object failed", err)
	}

	if unpackedPing.Id != expectId {
		t.Errorf("expect Id to be %s, but got %s.", expectId, unpackedPing.Id)
	}

	expectMsgBytes := []byte{8, 3, 18, 43, 10, 28, 116, 121, 112, 101, 46, 103, 111, 111, 103, 108, 101, 97, 112, 105, 115, 46, 99, 111, 109, 47, 100, 98, 112, 46, 80, 105, 110, 103, 18, 11, 10, 9, 104, 101, 108, 108, 111, 73, 100, 73, 100}

	if bytes.Compare(msgBytes, expectMsgBytes) != 0 {
		t.Error("pack msg bytes doesn't match")
	}
}
