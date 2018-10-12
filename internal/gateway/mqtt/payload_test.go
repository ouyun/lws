package mqtt

import (
	"bytes"
	"encoding/hex"
	// "log"
	"math/rand"
	"reflect"
	// "strconv"
	// "strings"
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
		ServSignature: []byte(RandStringBytesRmndr(64)),
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

// func TestPaly(t *testing.T) {
// 	code := "1 0 1 21 31 223 54 184 6 54 94 213 182 216 67 173 88 68 174 66 80 132 46 27 115 69 27 63 250 22 209 215 234 10 178 0 101 0 0 0 225 239 173 91 1 111 147 124 47 89 68 245 218 42 17 140 235 176 103 205 44 156 146 199 89 85 206 5 170 5 21 138 26 242 142 22 7 66 2 0 84 69 83 84 0 85 0 0 64 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 105 17 203 235 66 203 56 240 153 193 16 255 127 0 0 250 41 254 210 238 85 0 0 0 0 0 0 0 0 0 0 0"
// 	codeArr := strings.Split(code, " ")
// 	log.Printf("codeArr : %+v", len(codeArr))
// 	codeAddd := make([]byte, 147)
// 	log.Printf("len : %+v", len(codeArr))
// 	for index := 0; index < len(codeArr); index++ {
// 		value, _ := strconv.Atoi(codeArr[index])
// 		codeAddd[index] = byte(value)
// 	}
// 	log.Printf("codeAddd : %+v", codeAddd)

// 	ser := ServicePayload{}
// 	// payload := []byte{1, 0, 21, 31, 223, 54, 184, 6, 54, 94, 213, 182, 216, 67, 173, 88, 68, 174, 66, 80, 132, 46, 27, 115, 69, 27, 63, 250, 22, 209, 215, 234, 10, 178, 101, 0, 0, 0, 236, 191, 173, 91, 1, 111, 147, 124, 47, 89, 68, 245, 218, 42, 17, 140, 235, 176, 103, 205, 44, 156, 146, 199, 89, 85, 206, 5, 170, 5, 21, 138, 26, 242, 142, 22, 7, 2, 0, 84, 69, 83, 84, 0, 124, 210, 222, 176, 126, 230, 247, 200, 222, 69, 194, 113, 78, 89, 172, 81, 41, 102, 82, 229, 215, 234, 254, 5, 220, 45, 151, 216, 98, 58, 7, 18, 205, 104, 192, 207, 130, 255, 178, 106, 136, 95, 9, 86, 211, 186, 204, 134, 79, 58, 229, 200, 196, 110, 224, 75, 118, 105, 216, 140, 154, 91, 214, 9}
// 	log.Printf("len : %+v", len(codeAddd))
// 	err := DecodePayload(codeAddd, &ser)
// 	if err != nil {
// 		log.Printf("err: %+v", err)
// 	}
// 	log.Printf("ser: %+v", ser)
// }

// func TestStruct2Bytes(t *testing.T) {
// 	servReq := ServiceReply{}
// 	log.Printf("result : %+v", servReq)
// 	servReq.Nonce = uint16(10)
// 	servReq.Error = uint8(1)
// 	servReq.Version = uint32(1001)
// 	// servReq.AddressId = uint32(12)

// 	result, err := StructToBytes(servReq)
// 	log.Printf("result : %+v", result)
// 	if err != nil {
// 		t.Errorf("struct to bytes failed ! ")
// 	}
// }
