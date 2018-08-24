package mqtt

import (
	"fmt"
	"encoding/binary"
	"bytes"
	// "time"
	"math/rand"
	cRand "crypto/rand"

	"golang.org/x/crypto/ed25519"
)

type ServicePayload struct {
		Nonce         uint16
		Address       string
		Version       uint32
		TimeStamp     uint32
		ForkNum       uint8
		ForkList      string
		ReplyUTXON    uint16
		TopicPrefix   string
		Signature     string
}

type ServiceReply struct {
		Nonce         uint16
		Version       uint32
		Error     		uint8
		AddressId     uint32
		ForkBitmap    uint64
		ApiKeySeed    string
}

type SyncPayload struct {
		Nonce         uint16
		AddressId     uint32
		ForkID        string
		UTXOHash      string
		Signature     string
}

type SyncReply struct {
		Nonce         uint16
		Error         uint8
		BlockHash     string
		BlockHeight   uint32
		UTXONum       uint16
		UTXOList      string
		Continue      uint8
}

type UpdatePayload struct {
		Nonce         uint16
		AddressId     uint32
		ForkId        string
		BlockHash     string
		Height        uint32
		UpdateNum     uint16
		UpdateList    string
		Continue      uint8
}

type AbortPayload struct {
		Nonce         uint16
		AddressId     uint32
		Reason        uint8
		Signature     string
}

type SendTxPayload struct {
		Nonce         uint16
		AddressId     uint32
		ForkID        string
		TxData        string
		Signature     string
}

type SendTxReply struct {
		Nonce         uint16
		Error         uint8
		ErrCode       uint8
		ErrDesc       string
}


func GeneratePayload(i interface{}) (value []byte, err error){
	switch payload := i.(type) {
		case ServicePayload:
			// payload := v
			// &ServicePayload{
			// 	Nonce: uint16(1231),
			// 	Address: RandStringBytesRmndr(33),
			// 	Version: uint32(5363),
			// 	TimeStamp: uint32(time.Now().Unix()),
			// 	ForkNum:  uint8(1),
			// 	ForkList: RandStringBytesRmndr(32*1),
			// 	ReplyUTXON: uint16(2),
			// 	TopicPrefix: "DE0",
			// }
			pubK, privK, err := ed25519.GenerateKey(cRand.Reader)
			if err != nil {
				fmt.Printf("生成key失败")
			}
			fmt.Printf("结构体payload:%+v\n", payload)
			value := make([]byte, 2)
			value = IntToBytes(payload.Nonce)
			value = append(value, byte(uint8(1)))
			value = append(value, pubK...)
			value = append(value, IntToBytes(payload.Version)...)
			value = append(value, IntToBytes(payload.TimeStamp)...)
			value = append(value, byte(payload.ForkNum))
			value = append(value, []byte(payload.ForkList)...)
			value = append(value, IntToBytes(payload.ReplyUTXON)...)
			value = append(value, []byte(payload.TopicPrefix)...)
			value = append(value, Sign(privK, value)...)
			return value, err
		case SyncPayload:
			// payload := SyncPayload{
			// 	Nonce: uint16(1231),
			// 	AddressId: uint32(1231),
			// 	ForkID: RandStringBytesRmndr(32),
			// 	UTXOHash: RandStringBytesRmndr(32),
			// 	Signature:  RandStringBytesRmndr(20),
			// }
			fmt.Printf("结构体payload:%+v\n", payload)
			value := make([]byte, 2)
			value = IntToBytes(payload.Nonce)
			value = append(value, IntToBytes(payload.AddressId)...)
			value = append(value, []byte(payload.ForkID)...)
			value = append(value, []byte(payload.UTXOHash)...)
			value = append(value, []byte(payload.Signature)...)
			return value, err
		case UpdatePayload:
			// payload := UpdatePayload{
			// 	Nonce: uint16(1231),
			// 	AddressId: uint32(1231),
			// 	ForkId: RandStringBytesRmndr(32),
			// 	BlockHash: RandStringBytesRmndr(32),
			// 	Height:  uint32(1231),
			// 	UpdateNum: uint16(1231),
			// 	UpdateList:  RandStringBytesRmndr(20),
			// 	Continue: uint8(0),
			// }
			fmt.Printf("结构体payload:%+v\n", payload)
			value := make([]byte, 2)
			value = IntToBytes(payload.Nonce)
			value = append(value, IntToBytes(payload.AddressId)...)
			value = append(value, []byte(payload.ForkId)...)
			value = append(value, []byte(payload.BlockHash)...)
			value = append(value, IntToBytes(payload.Height)...)
			value = append(value, IntToBytes(payload.UpdateNum)...)
			value = append(value, []byte(payload.UpdateList)...)
			value = append(value, byte(payload.Continue))
			return value, err
		case AbortPayload:
			// payload := AbortPayload{
			// 	Nonce: uint16(1231),
			// 	AddressId: uint32(1231),
			// 	Reason: uint8(1),
			// 	Signature: RandStringBytesRmndr(20),
			// }
			fmt.Printf("结构体payload:%+v\n", payload)
			value := make([]byte, 2)
			value = IntToBytes(payload.Nonce)
			value = append(value, IntToBytes(payload.AddressId)...)
			value = append(value, byte(payload.Reason))
			value = append(value, []byte(payload.Signature)...)
			return value, err
		case SendTxPayload:
			// payload := SendTxPayload{
			// 	Nonce: uint16(1231),
			// 	AddressId: uint32(1231),
			// 	ForkID: RandStringBytesRmndr(32),
			// 	TxData: RandStringBytesRmndr(10),
			// 	Signature: RandStringBytesRmndr(20),
			// }
			fmt.Printf("结构体payload:%+v\n", payload)
			value := make([]byte, 2)
			value = IntToBytes(payload.Nonce)
			value = append(value, IntToBytes(payload.AddressId)...)
			value = append(value, []byte(payload.ForkID)...)
			value = append(value, []byte(payload.TxData)...)
			value = append(value, []byte(payload.Signature)...)
			return value, err
		}
		return value, err
	}

	func GenerateReply(types string) []byte {
		switch types {
			case "ServiceReply":
				reply := ServiceReply{
					Nonce: uint16(1231),
					Version: uint32(12312),
					Error: uint8(0),
					AddressId: uint32(3222),
					ForkBitmap: uint64(6112),
					ApiKeySeed: RandStringBytesRmndr(32),
				}
				fmt.Printf("结构体payload:%+v\n", reply)
				value := make([]byte, 2)
				value = IntToBytes(reply.Nonce)
				value = append(value, IntToBytes(reply.Version)...)
				value = append(value, byte(reply.Error))
				value = append(value, IntToBytes(reply.AddressId)...)
				value = append(value, IntToBytes(reply.ForkBitmap)...)
				value = append(value, []byte(reply.ApiKeySeed)...)
				return value
			case "SyncReply":
				reply := SyncReply{
					Nonce: uint16(1231),
					Error: uint8(0),
					BlockHash: RandStringBytesRmndr(32),
					BlockHeight: uint32(3222),
					UTXONum: uint16(6112),
					UTXOList: RandStringBytesRmndr(20),
					Continue: uint8(0),
				}
				fmt.Printf("结构体payload:%+v\n", reply)
				value := make([]byte, 2)
				value = IntToBytes(reply.Nonce)
				value = append(value, byte(reply.Error))
				value = append(value, []byte(reply.BlockHash)...)
				value = append(value, IntToBytes(reply.BlockHeight)...)
				value = append(value, IntToBytes(reply.UTXONum)...)
				value = append(value, []byte(reply.UTXOList)...)
				value = append(value, byte(reply.Continue))
				return value
			case "SendTxReply":
				reply := SendTxReply{
					Nonce: uint16(1231),
					Error: uint8(0),
				}
				fmt.Printf("结构体payload:SendTxReply%+v\n", reply)
				value := make([]byte, 2)
				value = IntToBytes(reply.Nonce)
				value = append(value, byte(reply.Error))
				return value
	}
	return nil
}

func IntToBytes(i interface{}) []byte {
	// buf := new(bytes.Buffer)
	// err := binary.Write(buf, binary.LittleEndian, i)
	// if err != nil {
	// 		fmt.Println("binary.Write failed:", err)
	// }
	// fmt.Printf("% x\n", buf.Bytes())
	// return buf.Bytes()
		switch v := i.(type) {
			case uint16:
				var buf = make([]byte, 2)
				binary.LittleEndian.PutUint16(buf, v)
				return buf
			case uint32:
				var buf = make([]byte, 4)
				binary.LittleEndian.PutUint32(buf, v)
				return buf
			case uint64:
				var buf = make([]byte, 8)
				binary.LittleEndian.PutUint64(buf, v)
				return buf
			case uint8:
				buf := []byte{byte(uint8(v))}
				return buf
		}
		return []byte{}
}

func BytesToInt(buf []byte) interface{} {
	// var value interface{}
	// switch v := len(buff); v {
	// case 2:
	// 	value
	// }
	// buf := bytes.NewReader(buff)
	// // for int32, the resulting size of buf will be 4 bytes
	// // for int64, the resulting size of buf will be 8 bytes
	// err := binary.Read(buf, binary.LittleEndian, i)
	// if err != nil {
	// 		fmt.Println("binary.Write failed:", err)
	// }
	// return
	fmt.Println(buf)
	switch v := len(buf); v {
	case 2:
		return binary.LittleEndian.Uint16(buf)
	case 4:
		return binary.LittleEndian.Uint32(buf)
	case 8:
		return binary.LittleEndian.Uint64(buf)
	default:
		var value uint8
		b := bytes.NewReader(buf)
		err := binary.Read(b, binary.LittleEndian, &value)
		if err != nil {
				fmt.Println("binary.Write failed:", err)
		}
		return value
	}
}


func ByteToBinaryString(data byte) (str string) {
    var a byte
    for i:=0; i < 8; i++ {
        a = data
        data <<= 1
				data >>= 1
				fmt.Println(a)
				fmt.Println(data)
        switch (a) {
        case data: str += "0"
        default: str += "1"
        }

        data <<= 1
    }
    return str
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
func RandStringBytesRmndr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func DecodePayload(payload []byte) {
	result := ServicePayload{}
	result.Nonce = BytesToInt(payload[:2]).(uint16)
	result.Address = string(payload[2:35])
	result.Version = BytesToInt(payload[35:39]).(uint32)
	result.TimeStamp = BytesToInt(payload[39:43]).(uint32)
	result.ForkNum = BytesToInt(payload[43:44]).(uint8)
	result.ForkList = string(payload[44:76])
	result.ReplyUTXON = BytesToInt(payload[76:78]).(uint16)
	result.TopicPrefix = string(payload[78:81])
	result.Signature = string(payload[81:])
	fmt.Println(ed25519.Verify(payload[2:35], payload[0:81], payload[81:]))
	fmt.Printf("接受结构体payload:%+v\n", result)
}

func Sign(key ed25519.PrivateKey, message []byte) []byte {
	sign := ed25519.Sign(key, message)
	return sign
}

