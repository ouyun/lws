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
	Nonce         uint16 `len:"2"`
	Address       []byte `len:"33"`
	Version       uint32 `len:"4"`
	TimeStamp     uint32 `len:"4"`
	ForkNum       uint8  `len:"1"`
	ForkList      []byte `len:"0"`
	ReplyUTXON    uint16 `len:"2"`
	TopicPrefix   string `len:"0"`
	SignBytes     uint16 `len:"2"`
	ServSignature []byte `len:"64"`
}

type ServiceReply struct {
	Nonce      uint16 `len:"2"`
	Version    uint32 `len:"4"`
	Error      uint8  `len:"1" type:"ServiceReply"`
	AddressId  uint32 `len:"4"`
	ForkBitmap uint64 `len:"8"`
	ApiKeySeed []byte `len:"32"`
}

type SyncPayload struct {
	Nonce     uint16 `len:"2"`
	AddressId uint32 `len:"4"`
	ForkID    []byte `len:"32"`
	UTXOHash  []byte `len:"32"`
	Signature []byte `len:"20"`
}

type SyncReply struct {
	Nonce       uint16 `len:"2"`
	Error       uint8  `len:"1" type:"SyncReply"`
	BlockHash   []byte `len:"32"`
	BlockHeight uint32 `len:"4"`
	BlockTime   uint32 `len:"4"`
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
	BlockTime  uint32 `len:"4"`
	UpdateNum  uint16 `len:"2"`
	UpdateList []byte `len:"0"`
	Continue   uint8  `len:"1"`
}

type AbortPayload struct {
	Nonce     uint16 `len:"2"`
	AddressId uint32 `len:"4"`
	Reason    uint8  `len:"1"`
	Signature []byte `len:"20"`
}

type SendTxPayload struct {
	Nonce     uint16 `len:"2"`
	AddressId uint32 `len:"4"`
	ForkID    []byte `len:"32"`
	TxData    []byte `len:"0"`
	Signature []byte `len:"20"`
}

type SendTxReply struct {
	Nonce   uint16 `len:"2"`
	Error   uint8  `len:"1" return:"SendTxReply"`
	ErrCode uint8  `len:"1"`
	ErrDesc string `len:"0"`
}

type TxData struct {
	NVersion   uint16 `len:"2"`
	NType      uint16 `len:"2"`
	NLockUntil uint32 `len:"4"`
	HashAnchor []byte `len:"32"`
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

	log.Printf("TxDataToStruct get params:tx--%+v\n", tx)
	var leftIndex uint64 = 0
	var sizeLen uint64 = 0
	totalLength := uint64(len(tx))
	for i := 0; i < resultValue.NumField(); i++ {
		// leng := resultType.Field(i).Tag.Get("len")
		leng, err := strconv.Atoi(resultType.Field(i).Tag.Get("len"))
		len64 := uint64(leng)
		if err != nil {
			return err
		}
		if (leftIndex + len64) > totalLength {
			return errors.New("slice bounds out of range")
		}
		// log.Printf("txData : %+v\n", txData)
		if resultValue.Field(i).CanSet() {
			switch resultValue.Field(i).Type().Kind() {
			case reflect.Slice:
				if leng > 0 {
					resultValue.Field(i).SetBytes(reverseBytes(tx[leftIndex:(leftIndex + len64)]))
				} else if resultType.Field(i).Name == "UtxoIndex" {
					resultValue.Field(i).SetBytes(reverseBytes(tx[leftIndex:(uint64(leftIndex) + (sizeLen * 33))]))
					len64 = sizeLen * 33
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
			case reflect.Int64:
				resultValue.Field(i).Set(
					reflect.ValueOf(int64(BytesToInt(tx[leftIndex:(leftIndex + len64)]).(uint64))))
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
	log.Printf("TxDataToStruct get result : %+v\n", txData)
	return err
}

func StructToBytes(s interface{}) (result []byte, err error) {
	buf := bytes.NewBuffer([]byte{})
	log.Printf("StructToBytes get param:struct-- %+v\n", s)
	value := reflect.ValueOf(s)
	valueType := reflect.TypeOf(s)
	for i := 0; i < value.NumField(); i++ {
		if valueType.Field(i).Name == "Error" {
			switch valueType.Field(i).Tag.Get("type") {
			case "SendTxReply":
				if value.Field(i).Uint() != 3 && value.Field(i).Uint() != 4 {
					buf.Write(IntToBytes(uint8(value.Field(i).Uint())))
					return buf.Bytes(), err
				}
			case "SyncReply":
				if value.Field(i).Uint() != 0 && value.Field(i).Uint() != 1 {
					buf.Write(IntToBytes(uint8(value.Field(i).Uint())))
					return buf.Bytes(), err
				}
			case "ServiceReply":
				if value.Field(i).Uint() != 0 {
					buf.Write(IntToBytes(uint8(value.Field(i).Uint())))
					return buf.Bytes(), err
				}
			}
		}
		switch value.Field(i).Type().Kind() {
		case reflect.Ptr:
			b, err := StructToBytes(value.Field(i).Elem())
			if err != nil {
				buf.Write(b)
			}
		case reflect.Slice:
			if valueType.Field(i).Name == "Data" && len(value.Field(i).Bytes()) == 0 {
				continue
			}
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
	// log.Printf("struct to bytes by struct: %+v\n", s)
	log.Printf("struct to bytes generate payload: %+v\n", result)
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
	case int64:
		var buf = make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(v))
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
	totalLength := len(payload)
	log.Printf("DecodePayload params: payload bytes length : %d\n", len(payload))
	log.Printf("DecodePayload params: payload bytes : %+v\n", payload)
	leftIndex := 0
	forkNum := 0
	for i := 0; i < resultValue.NumField(); i++ {
		// leng := resultType.Field(i).Tag.Get("len")
		leng, err := strconv.Atoi(resultType.Field(i).Tag.Get("len"))
		if err != nil {
			return err
		}
		log.Printf("result get : %+v\n", result)
		if resultValue.Field(i).CanSet() {
			switch resultValue.Field(i).Type().Kind() {
			case reflect.String:
				if leng > 0 {
					if (leftIndex + leng) > totalLength {
						return errors.New("slice bounds out of range")
					}
					resultValue.Field(i).SetString(string(payload[leftIndex:(leftIndex + leng)]))
				} else if resultType.Field(i).Name == "TxData" {
					allLength := len(payload)
					length, err := getAllLength(result)
					if err != nil {
						return err
					}
					leng = allLength - length
					if (leftIndex + leng) > totalLength {
						return errors.New("slice bounds out of range")
					}
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
					if resultType.Field(i).Name == "ServSignature" {
						resultValue.Field(i).SetBytes(payload[leftIndex:])
					} else {
						// log.Printf("result get : %+v\n", payload[leftIndex:(leftIndex+leng)])
						if (leftIndex + leng) > totalLength {
							return errors.New("slice bounds out of range")
						}
						resultValue.Field(i).SetBytes(payload[leftIndex:(leftIndex + leng)])
					}
				} else if resultType.Field(i).Name == "TxData" {
					allLength := len(payload)
					length, err := getAllLength(result)
					if err != nil {
						return err
					}
					leng = allLength - length
					if (leftIndex + leng) > totalLength {
						return errors.New("slice bounds out of range")
					}
					resultValue.Field(i).SetBytes(payload[leftIndex:(leftIndex + leng)])
				} else if resultType.Field(i).Name == "ForkList" {
					leng = forkNum * 32
					if (leftIndex + leng) > totalLength {
						return errors.New("slice bounds out of range")
					}
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
				if (leftIndex + leng) > totalLength {
					return errors.New("slice bounds out of range")
				}
				resultValue.Field(i).Set(
					reflect.ValueOf(BytesToInt(payload[leftIndex:(leftIndex + leng)]).(uint8)))
				if resultType.Field(i).Name == "ForkNum" {
					forkNum = int(BytesToInt(payload[leftIndex:(leftIndex + leng)]).(uint8))
				}
			case reflect.Uint16:
				if (leftIndex + leng) > totalLength {
					return errors.New("slice bounds out of range")
				}
				resultValue.Field(i).Set(
					reflect.ValueOf(BytesToInt(payload[leftIndex:(leftIndex + leng)]).(uint16)))
			case reflect.Uint32:
				if (leftIndex + leng) > totalLength {
					return errors.New("slice bounds out of range")
				}
				resultValue.Field(i).Set(
					reflect.ValueOf(BytesToInt(payload[leftIndex:(leftIndex + leng)]).(uint32)))
			case reflect.Uint64:
				if (leftIndex + leng) > totalLength {
					return errors.New("slice bounds out of range")
				}
				resultValue.Field(i).Set(
					reflect.ValueOf(BytesToInt(payload[leftIndex:(leftIndex + leng)]).(uint64)))
			default:
				err = errors.New("unsuport type")
			}
		} else {
			err = errors.New("decode payload get err: can not set value!")
		}
		leftIndex = (leftIndex + leng)
	}
	log.Printf("DecodePayload get result : %+v\n", result)
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

// reverse Bytes
func reverseBytes(src []byte) []byte {
	srcTemp := make([]byte, len(src))
	copy(srcTemp[:], src)
	for i := len(srcTemp)/2 - 1; i >= 0; i-- {
		opp := len(srcTemp) - 1 - i
		srcTemp[i], srcTemp[opp] = srcTemp[opp], srcTemp[i]
	}
	return srcTemp
}

func reverseString(s string) string {
	runes := []rune(s)

	// runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	for i := 0; i < len(s); i += 2 {
		runes[i], runes[i+1] = runes[i+1], runes[i]
	}
	return string(runes)
}
