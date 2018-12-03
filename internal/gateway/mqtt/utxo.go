package mqtt

import (
	"bytes"
	"log"

	"golang.org/x/crypto/blake2b"
)

type UTXO struct {
	TXID        []byte `len:"32"`
	Out         uint8  `len:"1"`
	BlockHeight uint32 `len:"4"`
	Type        uint16 `len:"2"`
	Amount      int64  `len:"8"`
	Sender      []byte `len:"33"`
	LockUntil   uint32 `len:"4"`
	DataSize    uint16 `len:"2"`
	Data        []byte `len:"0"`
}

type UTXOUpdate struct {
	OpType      uint8  `len:"1"`
	UTXOIndex   []byte `len:"33"`
	BlockHeight uint32 `len:"4"`
	UTXO        *UTXO  `len:"0"`
}

// get utxo hash
func UTXOHash(u *[]UTXO) []byte {
	buf := bytes.NewBuffer([]byte{})
	for _, value := range *u {
		buf.Write(GetIndex(&value))
		buf.Write(IntToBytes(value.BlockHeight))
	}
	hash := blake2b.Sum256(buf.Bytes())
	return hash[:32]
}

// get utxo index bytes
func GetIndex(u *UTXO) []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.Write(u.TXID)
	buf.Write(IntToBytes(u.Out))
	return buf.Bytes()
}

// utxo list to bytes
func UTXOListToByte(u *[]UTXO) (result []byte, err error) {
	buf := bytes.NewBuffer([]byte{})
	for index := 0; index < len(*u); index++ {
		(*u)[index].DataSize = uint16(len((*u)[index].Data))
		if len((*u)[index].Sender) == 0 {
			(*u)[index].Sender = make([]byte, 33)
		}
		utxoByte, err := StructToBytes((*u)[index])
		if err != nil {
			log.Printf("StructToBytes err: %+v", err)
			return result, err
		}
		buf.Write(utxoByte)
	}
	result = buf.Bytes()
	return result, err
}

func UTXOUpdateListToByte(u *[]UTXOUpdate) (result []byte, err error) {
	buf := bytes.NewBuffer([]byte{})
	for index := 0; index < len(*u); index++ {
		buf.Write([]byte{byte((*u)[index].OpType)})
		switch (*u)[index].OpType {
		case 0:
		case 1:
			buf.Write((*u)[index].UTXOIndex)
		case 2:
			buf.Write((*u)[index].UTXOIndex)
			buf.Write(IntToBytes((*u)[index].BlockHeight))
		case 3:
			(*u)[index].UTXO.DataSize = uint16(len((*u)[index].UTXO.Data))
			log.Printf("utxo %d: %+v", index, *((*u)[index].UTXO))
			if len((*u)[index].UTXO.Sender) == 0 {
				(*u)[index].UTXO.Sender = make([]byte, 33)
			}
			utxoByte, err := StructToBytes(*((*u)[index].UTXO))
			if err != nil {
				log.Printf("StructToBytes err: %+v", err)
				return result, err
			}
			buf.Write(utxoByte)
		}
	}
	result = buf.Bytes()
	return result, err
}
