package crypto

import (
	"bytes"
	"testing"
	"encoding/hex"
	"crypto/hmac"
)

func TestApiKey(t *testing.T) {
	for i := 0; i < 100000; i++ {
		lwsPubK, lwsPriK := GenerateKeyPair(nil)

		cliPubK, cliPriK := GenerateKeyPair(nil)

		lwsApiKey := GenerateKeyApiKey(&lwsPriK, &cliPubK)

		cliApiKey := GenerateKeyApiKey(&cliPriK, &lwsPubK)

		if bytes.Compare(lwsApiKey[:], cliApiKey[:]) != 0 {
			t.Error("generate ApiKey fail")
		}
	}
}

type casePair struct {
		message   []byte
		apiKey    []byte
		messageMac []byte
}


func TestSign(t *testing.T) {
	cases := []casePair {
		{
			[]byte("this is a message"),
			Decode("c22d784bf2a57e085c44fc1a0c5a662bec21cf3ca8fa9d4bfbdba6b12675f3d1"),
			Decode("f838deab94684499cc3c5d2df907a783309bdcfb"),
		},
		{
			[]byte(""),
			Decode("c22d784bf2a57e085c44fc1a0c5a662bec21cf3ca8fa9d4bfbdba6b12675f3d1"),
			Decode("48bb0d502f84674b34334f0ed5a1759d469a4ae8"),
		},
		{
			[]byte("这是一个测试用例！"),
			Decode("c22d784bf2a57e085c44fc1a0c5a662bec21cf3ca8fa9d4bfbdba6b12675f3d1"),
			Decode("07bd9ef8fa4fb13f2bc372cc89dde2bdd2c27e85"),
		},
	}

	for _,v := range cases {
		mac := SignWithApiKey(v.apiKey, v.message)
		if !hmac.Equal(mac, v.messageMac) {
			t.Error("sign fail")
		}
	}
}

func Decode(str string) []byte {
	v ,_ := hex.DecodeString(str)
	return v
}
