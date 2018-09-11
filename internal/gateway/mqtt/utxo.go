package mqtt

import (
	"bytes"

	"golang.org/x/crypto/blake2b"
)

type UTXO struct {
	TXID        []byte `len:"32"`
	Out         uint8  `len:"1"`
	BlockHeight uint32 `len:"4"`
	Type        uint16 `len:"2"`
	Amount      uint64 `len:"8"`
	Sender      []byte `len:"33"`
	LockUntil   string `len:"4"`
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
	hash := blake2b.Sum512(buf.Bytes())
	return hash[:31]
}

// get utxo index bytes
func GetIndex(u *UTXO) []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.Write(u.TXID)
	buf.Write(IntToBytes(u.Out))
	return buf.Bytes()
}
