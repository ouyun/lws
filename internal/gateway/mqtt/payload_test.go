package mqtt

import (
	"bytes"
	"encoding/hex"
	// "log"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

type casepair struct {
	input  interface{}
	result []byte
}

type casepair2 struct {
	input  []byte
	result interface{}
}

func TestTxDataToStruct(t *testing.T) {
	address, _ := hex.DecodeString("6f937c2f5944f5da2a118cebb067cd2c9c92c75955ce05aa05158a1af28e1607")
	anch, _ := hex.DecodeString("6f937c2f6f937c2f")

	txs := TxData{
		NVersion:   uint16(45),
		NType:      uint16(121),
		NLockUntil: uint32(1212),
		HashAnchor: anch,
		Size:       uint64(8),
		UtxoIndex:  genaBytes(),
		Prefix:     uint8(122),
		Data:       address,
		NAmount:    int64(121),
		NTxFee:     int64(121),
	}
	txData := TxData{}
	result, err := StructToBytes(txs)
	if err != nil {
		t.Errorf("struct to bytes failed ! ")
	}
	err = TxDataToStruct(result, &txData)
	if err != nil {
		t.Errorf("txData to struct failed ! ")
	}
}

func genaBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	for index := 0; index < 264; index++ {
		buf.Write([]byte{byte(rand.Intn(255))})
	}
	return buf.Bytes()
}

func TestStructToBytesAndDecode(t *testing.T) {
	addr, _ := hex.DecodeString("6f937c2f5944f5da2a118cebb067cd2c9c92c75955ce05aa05158a1af28e1607")
	address := make([]byte, 1)
	address[0] = uint8(1)
	address = append(address, addr[:]...)
	topicPrefix := "wqweqwasasqw" + string(byte(0x00))
	servicePayload := ServicePayload{ //serviceRequ
		Nonce:         uint16(1231),
		Address:       address,
		Version:       uint32(5363),
		TimeStamp:     uint32(time.Now().Unix()),
		ForkNum:       uint8(9),
		ForkList:      []byte(RandStringBytesRmndr(32 * 9)),
		ReplyUTXON:    uint16(10),
		TopicPrefix:   topicPrefix,
		SignBytes:     uint16(64),
		ServSignature: RandStringBytesRmndr(64),
	}
	result, err := StructToBytes(servicePayload)
	if err != nil {
		t.Errorf("struct to bytes failed ! ")
	}
	decodeServicePayload := ServicePayload{}
	err = DecodePayload(result, &decodeServicePayload)
	if err != nil {
		t.Errorf("Decode Payload failed ! ")
	}
	decodeServicePayload.TopicPrefix = decodeServicePayload.TopicPrefix + string(byte(0x00))
	if !reflect.DeepEqual(decodeServicePayload, servicePayload) {
		t.Errorf("decode struct did not equal sended service struct! ")
	}
}

func TestIntToBytes(t *testing.T) {
	var cases = []casepair{
		{uint64(1212), []byte{188, 4, 0, 0, 0, 0, 0, 0}},
		{uint16(2131), []byte{83, 8}},
		{uint32(112), []byte{112, 0, 0, 0}},
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
	var cases = []casepair2{
		{[]byte{188, 4, 0, 0, 0, 0, 0, 0}, uint64(1212)},
		{[]byte{83, 8}, uint16(2131)},
		{[]byte{112, 0, 0, 0}, uint32(112)},
		{[]byte{80}, uint8(80)},
	}
	for _, pair := range cases {
		v := BytesToInt(pair.input)
		if v != pair.result {
			t.Errorf("case (%v) expect (%v) but got (%v)", pair.input, pair.result, v)
		}
	}
}

// func TestReverse(t *testing.T) {
// 	src := []byte{1, 2, 3, 12, 31, 233}
// 	log.Printf("reverse bytes get : %+v \n", reverseBytes(src))
// 	int64Srv := IntToBytes(uint64(1121212))
// 	log.Printf("uint64 bytes get : %+v", int64Srv)
// 	int64S := BytesToInt(int64Srv)
// 	log.Printf("int64S value : %+v", int64S)
// 	log.Printf("reverse uint64 bytes get : %+v", reverseBytes(int64Srv))
// 	log.Printf("reverse uint64 bytes get : %+v", reverseBytes(int64Srv))
// 	int32Srv := IntToBytes(uint32(2312))
// 	log.Printf("uint32 bytes get : %+v", byUint32)
// 	log.Printf("reverse uint32 bytes get : %+v", reverseBytes(int32Srv))
// 	int16Srv := IntToBytes(uint32(2553))
// 	log.Printf("uint16 bytes get : %+v", int16Srv)
// 	log.Printf("reverse uint16 bytes get : %+v", reverseBytes(int16Srv))
// 	int8Srv := IntToBytes(uint8(254))
// 	log.Printf("uint8 bytes get : %+v", int8Srv)
// 	log.Printf("reverse uint8 bytes get : %+v", reverseBytes(int8Srv))
// }
