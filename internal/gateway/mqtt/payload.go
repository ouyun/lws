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
	Address     []byte `len:"32"`
	Version     uint32 `len:"4"`
	TimeStamp   uint32 `len:"4"`
	ForkNum     uint8  `len:"1"`
	ForkList    []byte `len:"0"`
	ReplyUTXON  uint16 `len:"2"`
	TopicPrefix string `len:"0"`
	Signature   string `len:"64"`
}

type ServiceReply struct {
	Nonce      uint16 `len:"2"`
	Version    uint32 `len:"4"`
	Error      uint8  `len:"1"`
	AddressId  uint32 `len:"4"`
	ForkBitmap uint64 `len:"8"`
	ApiKeySeed []byte `len:"32"`
}

type SyncPayload struct {
	Nonce     uint16 `len:"2"`
	AddressId uint32 `len:"4"`
	ForkID    []byte `len:"32"`
	UTXOHash  string `len:"32"`
	Signature string `len:"20"`
}

type SyncReply struct {
	Nonce       uint16 `len:"2"`
	Error       uint8  `len:"1"`
	BlockHash   []byte `len:"32"`
	BlockHeight uint32 `len:"4"`
	UTXONum     uint16 `len:"2"`
	UTXOList    []byte `len:"0"`
	Continue    uint8  `len:"1"`
}

type UpdatePayload struct {
	Nonce      uint16 `len:"2"`
	AddressId  uint32 `len:"4"`
	ForkId     []byte `len:"32"`
	BlockHash  []byte `len:"32"`
	Height     uint32 `len:"4"`
	UpdateNum  uint16 `len:"2"`
	UpdateList []byte `len:"0"`
	Continue   uint8  `len:"1"`
}

type AbortPayload struct {
	Nonce     uint16 `len:"2"`
	AddressId uint32 `len:"4"`
	Reason    uint8  `len:"1"`
	Signature string `len:"20"`
}

type SendTxPayload struct {
	Nonce     uint16 `len:"2"`
	AddressId uint32 `len:"4"`
	ForkID    []byte `len:"32"`
	TxData    []byte `len:"0"`
	Signature string `len:"20"`
}

type SendTxReply struct {
	Nonce   uint16 `len:"2"`
	Error   uint8  `len:"1"`
	ErrCode uint8  `len:"1"`
	ErrDesc string `len:"0"`
}

type TxData struct {
	NVersion   uint16 `len:"2"`
	NType      uint16 `len:"2"`
	NLockUntil uint32 `len:"4"`
	HashAnchor []byte `len:"8"`
	Size       uint64 `len:"0"`
	UtxoIndex  []byte `len:"0"`
	Prefix     uint8  `len:"1"`
	Data       []byte `len:"32"`
	NAmount    int64  `len:"8"`
	NTxFee     int64  `len:"8"`
}

func TxDataToStruct(tx []byte, txData *TxData) (err error) {

	resultValue := reflect.ValueOf(txData).Elem()

	resultType := reflect.TypeOf(txData).Elem()

	// log.Printf("resultType : %+v\n", resultType.NumField())
	var leftIndex uint64 = 0
	var sizeLen uint64 = 0
	for i := 0; i < resultValue.NumField(); i++ {
		// leng := resultType.Field(i).Tag.Get("len")
		leng, err := strconv.Atoi(resultType.Field(i).Tag.Get("len"))
		len64 := uint64(leng)
		if err != nil {
			return err
		}
		if resultValue.Field(i).CanSet() {
			switch resultValue.Field(i).Type().Kind() {
			case reflect.Slice:
				if leng > 0 {
					resultValue.Field(i).SetBytes(tx[leftIndex:(leftIndex + len64)])
				} else if resultType.Field(i).Name == "UtxoIndex" {
					resultValue.Field(i).SetBytes(tx[leftIndex:(uint64(leftIndex) + (sizeLen * 33))])
					len64 = sizeLen
				}
			case reflect.Uint8:
				resultValue.Field(i).Set(
					reflect.ValueOf(BytesToInt(tx[leftIndex:(leftIndex + len64)]).(uint8)))
			case reflect.Uint16:
				resultValue.Field(i).Set(
					reflect.ValueOf(BytesToInt(tx[leftIndex:(leftIndex + len64)]).(uint16)))
			case reflect.Uint32:
				resultValue.Field(i).Set(
					reflect.ValueOf(BytesToInt(tx[leftIndex:(leftIndex + len64)]).(uint32)))
			case reflect.Uint64:
				if resultType.Field(i).Name == "Size" {
					num := BytesToInt(tx[leftIndex:(leftIndex + 1)]).(uint8)
					if num < 253 {
						sizeLen = uint64(num)
						len64 = 1
					} else if num == 253 {
						sizeLen = uint64(num) + BytesToInt(tx[leftIndex+1:(leftIndex+3)]).(uint64)
						len64 = 3
					} else if num == 254 {
						sizeLen = uint64(num) + BytesToInt(tx[leftIndex+1:(leftIndex+5)]).(uint64)
						len64 = 5
					} else if num == 255 {
						sizeLen = uint64(num) + BytesToInt(tx[leftIndex+1:(leftIndex+9)]).(uint64)
						len64 = 9
					}
					resultValue.Field(i).Set(
						reflect.ValueOf(sizeLen))
				} else {
					resultValue.Field(i).Set(
						reflect.ValueOf(BytesToInt(tx[leftIndex:(leftIndex + len64)]).(uint64)))
				}
			default:
				err = errors.New("unsuport type")
			}
		} else {
			err = errors.New("can not set value ")
		}
		leftIndex += len64
	}
	return err
}

func StructToBytes(s interface{}) (result []byte, err error) {
	buf := bytes.NewBuffer([]byte{})

	value := reflect.ValueOf(s)
	for i := 0; i < value.NumField(); i++ {
		switch value.Field(i).Type().Kind() {
		case reflect.Ptr:
			b, err := StructToBytes(value.Field(i).Elem())
			if err != nil {
				buf.Write(b)
			}
		case reflect.Slice:
			buf.Write(value.Field(i).Bytes())
		case reflect.String:
			buf.Write([]byte(value.Field(i).String()))
		case reflect.Uint8:
			buf.Write(IntToBytes(uint8(value.Field(i).Uint())))
		case reflect.Uint16:
			buf.Write(IntToBytes(uint16(value.Field(i).Uint())))
		case reflect.Uint32:
			buf.Write(IntToBytes(uint32(value.Field(i).Uint())))
		case reflect.Uint64:
			buf.Write(IntToBytes(uint64(value.Field(i).Uint())))
		case reflect.Int64:
			buf.Write(IntToBytes(int64(value.Field(i).Int())))
		default:
			err = errors.New("unsuport type")
		}
	}
	result = buf.Bytes()
	// log.Printf("generate 结构体payload: %+v\n", s)
	return result, err
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

func DecodePayload(payload []byte, result interface{}) (err error) {

	resultValue := reflect.ValueOf(result).Elem()
	resultType := reflect.TypeOf(result).Elem()

	// log.Printf("resultType : %+v\n", resultType.NumField())
	leftIndex := 0
	forkNum := 0
	for i := 0; i < resultValue.NumField(); i++ {
		// leng := resultType.Field(i).Tag.Get("len")
		leng, err := strconv.Atoi(resultType.Field(i).Tag.Get("len"))
		if err != nil {
			return err
		}
		if resultValue.Field(i).CanSet() {
			switch resultValue.Field(i).Type().Kind() {
			case reflect.String:
				if leng > 0 {
					resultValue.Field(i).SetString(string(payload[leftIndex:(leftIndex + leng)]))
				} else if resultType.Field(i).Name == "TxData" {
					allLength := len(payload)
					length, err := getAllLength(result)
					if err != nil {
						return err
					}
					leng = allLength - length
					resultValue.Field(i).SetString(string(payload[leftIndex:(leftIndex + leng)]))
				} else {
					buff := []byte{}
					buf := bytes.NewBuffer(buff)
					buf.Write(payload[leftIndex:])
					delim := byte(0x00)
					h, _ := buf.ReadBytes(delim)
					leng = len(h)
					resultValue.Field(i).SetString(string(h[:leng-1]))
				}
			case reflect.Slice:
				if leng > 0 {
					resultValue.Field(i).SetBytes(payload[leftIndex:(leftIndex + leng)])
				} else if resultType.Field(i).Name == "TxData" {
					allLength := len(payload)
					length, err := getAllLength(result)
					if err != nil {
						return err
					}
					leng = allLength - length
					resultValue.Field(i).SetBytes(payload[leftIndex:(leftIndex + leng)])
				} else if resultType.Field(i).Name == "ForkList" {
					leng = forkNum * 32
					resultValue.Field(i).SetBytes(payload[leftIndex:(leftIndex + leng)])
				} else {
					buff := []byte{}
					buf := bytes.NewBuffer(buff)
					buf.Write(payload[leftIndex:])
					delim := byte(0x00)
					h, _ := buf.ReadBytes(delim)
					leng = len(h)
					resultValue.Field(i).SetBytes(h[:leng-1])
				}
			case reflect.Uint8:
				resultValue.Field(i).Set(
					reflect.ValueOf(BytesToInt(payload[leftIndex:(leftIndex + leng)]).(uint8)))
				if resultType.Field(i).Name == "ForkNum" {
					forkNum = int(BytesToInt(payload[leftIndex:(leftIndex + leng)]).(uint8))
				}
			case reflect.Uint16:
				resultValue.Field(i).Set(
					reflect.ValueOf(BytesToInt(payload[leftIndex:(leftIndex + leng)]).(uint16)))
			case reflect.Uint32:
				resultValue.Field(i).Set(
					reflect.ValueOf(BytesToInt(payload[leftIndex:(leftIndex + leng)]).(uint32)))
			case reflect.Uint64:
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
	// log.Printf("result : %+v\n", result)
	return err
}

func getAllLength(result interface{}) (length int, err error) {
	resultValue := reflect.ValueOf(result).Elem()

	resultType := reflect.TypeOf(result).Elem()
	for i := 0; i < resultValue.NumField(); i++ {
		leng, err := strconv.Atoi(resultType.Field(i).Tag.Get("len"))
		if err != nil {
			return length, err
		}
		length += leng
	}
	return length, err
}

func Sign(key ed25519.PrivateKey, message []byte) []byte {
	sign := ed25519.Sign(key, message)
	return sign
}
