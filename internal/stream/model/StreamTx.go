package model

import (
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
)

type StreamTx struct {
	*lws.Transaction
	Sender []byte
}
