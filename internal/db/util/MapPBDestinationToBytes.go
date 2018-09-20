package util

import (
	"bytes"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
)

func MapPBDestinationToBytes(dest *lws.Transaction_CDestination) []byte {
	buffer := bytes.NewBuffer(make([]byte, 0))
	buffer.Grow(33)
	buffer.WriteByte(byte(uint8(dest.Prefix)))
	buffer.Write(dest.Data)
	return buffer.Bytes()
}
