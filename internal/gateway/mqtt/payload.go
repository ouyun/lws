package mqtt

import (
	"bytes"
	"encoding/binary"
	"log"
	// "time"
	"errors"
	"math/rand"
	"reflect"
	"strconv"
	// cRand "crypto/rand"
	// "lws/internal/gateway/crypto"
	"golang.org/x/crypto/ed25519"
)

type ServicePayload struct {
	Nonce       uint16 `len:"2"`
	Address0    uint8  `len:"1"`
	Address     string `len:"32"`
	Version     uint32 `len:"4"`
	TimeStamp   uint32 `len:"4"`
	ForkNum     uint8  `len:"1"`
	ForkList    string `len:"32"`
	ReplyUTXON  uint16 `len:"2"`
	TopicPrefix string `len:"0"`
	Signature   string `len:"64"`
}

type ServiceReply struct {
	Nonce      uint16
	Version    uint32
	Error      uint8
	AddressId  uint32
	ForkBitmap uint64
	ApiKeySeed string
}

type SyncPayload struct {
	Nonce     uint16
	AddressId uint32
	ForkID    string
	UTXOHash  string
	Signature string
}

type SyncReply struct {
	Nonce       uint16
	Error       uint8
	BlockHash   string
	BlockHeight uint32
	UTXONum     uint16
	UTXOList    string
	Continue    uint8
}

type UpdatePayload struct {
	Nonce      uint16
	AddressId  uint32
	ForkId     string
	BlockHash  string
	Height     uint32
	UpdateNum  uint16
	UpdateList string
	Continue   uint8
}

type AbortPayload struct {
	Nonce     uint16
	AddressId uint32
	Reason    uint8
	Signature string
}

type SendTxPayload struct {
	Nonce     uint16
	AddressId uint32
	ForkID    string
	TxData    string // 20 byte
	Signature string
}

type SendTxReply struct {
	Nonce   uint16
	Error   uint8
	ErrCode uint8
	ErrDesc string
}

func GenerateService(s interface{}) (result []byte, err error) {
	// buff := []byte{}
	// log.Printf("收到 interface: %+v\n", s)
	buf := bytes.NewBuffer([]byte{})

	value := reflect.ValueOf(s)
	for i := 0; i < value.NumField(); i++ {
		var tempByte []byte
		switch value.Field(i).Type().Name() {
		case "string":
			tempByte = []byte(value.Field(i).String())
		case "uint8":
			// log.Printf("收到 int: %d\n", IntToBytes(uint8(value.Field(i).Uint())))
			tempByte = IntToBytes(uint8(value.Field(i).Uint()))
		case "uint16":
			tempByte = IntToBytes(uint16(value.Field(i).Uint()))
		case "uint32":
			tempByte = IntToBytes(uint32(value.Field(i).Uint()))
		case "uint64":
			tempByte = IntToBytes(uint64(value.Field(i).Uint()))
		default:
			err = errors.New("unsuport type")
		}
		buf.Write(tempByte)
	}
	result = buf.Bytes()
	log.Printf("generate 结构体payload: %+v\n", result)
	return result, err
}

func GeneratePayload(i interface{}) (value []byte, err error) {
	switch payload := i.(type) {
	case ServicePayload:
		value := make([]byte, 152)
		copy(value, IntToBytes(payload.Nonce))
		copy(value[2:], []byte{uint8(payload.Address0)})
		copy(value[3:], []byte(payload.Address))
		copy(value[35:], IntToBytes(payload.Version))
		copy(value[39:], IntToBytes(payload.TimeStamp))
		copy(value[41:], []byte{byte(payload.ForkNum)})
		copy(value[42:], []byte(payload.ForkList))
		copy(value[74:], IntToBytes(payload.ReplyUTXON))
		copy(value[76:], []byte(payload.TopicPrefix))
		copy(value[78:], []byte(payload.Signature))
		// log.Printf("generate 结构体payload: %+v\n", payload)
		return value, err
	case SyncPayload:
		value := make([]byte, 90)
		copy(value, IntToBytes(payload.Nonce))
		copy(value[2:], IntToBytes(payload.AddressId))
		copy(value[6:], []byte(payload.ForkID))
		copy(value[38:], []byte(payload.UTXOHash))
		copy(value[70:], []byte(payload.Signature))
		log.Printf("generate 结构体SyncPayload: %+v\n", payload)
		return value, err
	case UpdatePayload:
		log.Printf("结构体UpdatePayload: %+v\n", payload)
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
		value := make([]byte, 27)
		copy(value, IntToBytes(payload.Nonce))
		copy(value[2:], IntToBytes(payload.AddressId))
		copy(value[6:], []byte{byte(payload.Reason)})
		copy(value[7:], []byte(payload.Signature))
		log.Printf("结构体AbortPayload: %+v\n", payload)
		return value, err
	case SendTxPayload:
		value := make([]byte, 78)
		copy(value, IntToBytes(payload.Nonce))
		copy(value[2:], IntToBytes(payload.AddressId))
		copy(value[6:], []byte(payload.ForkID))
		copy(value[38:], []byte(payload.TxData))
		copy(value[58:], []byte(payload.Signature))
		log.Printf("结构体SendTxPayload: %+v\n", payload)
		return value, err
	}
	return value, err
}

func GenerateReply(types string) []byte {
	switch types {
	case "ServiceReply":
		reply := ServiceReply{
			Nonce:      uint16(1231),
			Version:    uint32(12312),
			Error:      uint8(0),
			AddressId:  uint32(3222),
			ForkBitmap: uint64(6112),
			ApiKeySeed: RandStringBytesRmndr(32),
		}
		log.Printf("结构体payload:%+v\n", reply)
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
			Nonce:       uint16(1231),
			Error:       uint8(0),
			BlockHash:   RandStringBytesRmndr(32),
			BlockHeight: uint32(3222),
			UTXONum:     uint16(6112),
			UTXOList:    RandStringBytesRmndr(20),
			Continue:    uint8(0),
		}
		log.Printf("结构体SyncReply: %+v\n", reply)
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
		log.Printf("结构体payload:SendTxReply%+v\n", reply)
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
			log.Println("binary.Write failed:", err)
		}
		return value
	}
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
	result.Address0 = BytesToInt(payload[2:3]).(uint8)
	result.Address = string(payload[3:35])
	result.Version = BytesToInt(payload[35:39]).(uint32)
	result.TimeStamp = BytesToInt(payload[39:43]).(uint32)
	result.ForkNum = BytesToInt(payload[43:44]).(uint8)
	result.ForkList = string(payload[44:76])
	result.ReplyUTXON = BytesToInt(payload[76:78]).(uint16)
	result.TopicPrefix = string(payload[78:88])
	result.Signature = string(payload[88:])
	log.Printf("接受结构体payload: %+v\n", result)
}

func DecodeBy(payload []byte, result ServicePayload) (r interface{}, err error) {
	// s := ServicePayload{}

	resultValue := reflect.ValueOf(&result).Elem()

	resultType := reflect.TypeOf(result)

	// log.Printf("resultType : %+v\n", resultType.NumField())
	leftIndex := 0
	for i := 0; i < resultValue.NumField(); i++ {
		// leng := resultType.Field(i).Tag.Get("len")
		leng, err := strconv.Atoi(resultType.Field(i).Tag.Get("len"))
		if err != nil {
			return r, err
		}
		if resultValue.Field(i).CanSet() {
			switch resultValue.Field(i).Type().Name() {
			case "string":
				if leng > 0 {
					resultValue.Field(i).SetString(string(payload[leftIndex:(leftIndex + leng)]))
				} else {
					buff := []byte{}
					buf := bytes.NewBuffer(buff)
					delim := byte(0)
					h, _ := buf.ReadBytes(delim)
					leng = len(h)
					resultValue.Field(i).SetString(string(h[:]))
				}
			case "uint8":
				resultValue.Field(i).Set(
					reflect.ValueOf(BytesToInt(payload[leftIndex:(leftIndex + leng)]).(uint8)))
			case "uint16":
				resultValue.Field(i).Set(
					reflect.ValueOf(BytesToInt(payload[leftIndex:(leftIndex + leng)]).(uint16)))
			case "uint32":
				resultValue.Field(i).Set(
					reflect.ValueOf(BytesToInt(payload[leftIndex:(leftIndex + leng)]).(uint32)))
			case "uint64":
				resultValue.Field(i).Set(
					reflect.ValueOf(BytesToInt(payload[leftIndex:(leftIndex + leng)]).(uint64)))
			default:
				err = errors.New("unsuport type")
			}
		} else {
			err = errors.New("can not set value ")
		}
		leftIndex = (leftIndex + leng)
	}
	// log.Printf("decode 结构体payload: %+v\n", resultValue)
	r = result
	return r, err
}

func Sign(key ed25519.PrivateKey, message []byte) []byte {
	sign := ed25519.Sign(key, message)
	return sign
}
