package mqtt

import (
	"testing"
	// "fmt"
	"bytes"
	// "time"
)

type casepair struct {
    input     interface{}
		result    []byte
}

type casepair2 struct {
    input     []byte
		result    interface{}
}

func TestIntToBytes(t *testing.T) {
	var cases = []casepair {
		{uint64(1212), []byte{188,4,0,0,0,0,0,0}},
		{uint16(2131), []byte{83,8}},
		{uint32(112), []byte{112,0,0,0}},
		{uint8(155), []byte{155}},
	}
	for _, pair := range cases {
			v := IntToBytes(pair.input)
			if !bytes.Equal(v, pair.result) {
				t.Errorf("case (%v) expect (%v) but got (%v)", pair.input, pair.result, v)
			}
	}
}

func TestBytesToInt(t *testing.T) {
	var cases = []casepair2 {
		{[]byte{188,4,0,0,0,0,0,0}, uint64(1212)},
		{[]byte{83,8}, uint16(2131)},
		{[]byte{112,0,0,0}, uint32(112)},
		{[]byte{80}, uint8(80)},
	}
	for _, pair := range cases {
			v := BytesToInt(pair.input)
			if v != pair.result {
				t.Errorf("case (%v) expect (%v) but got (%v)", pair.input, pair.result, v)
			}
	}
}

func TestDecodePayload(t *testing.T) {
	// var cases = []ServicePayload {
	// 		{
	// 			Nonce: uint16(1231),
	// 			Version: uint32(5363),
	// 			TimeStamp: uint32(time.Now().Unix()),
	// 			ForkNum:  uint8(1),
	// 			ForkList: RandStringBytesRmndr(32*1),
	// 			ReplyUTXON: uint16(2),
	// 			TopicPrefix: "DE0",
	// 		},
	// }

}
