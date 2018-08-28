package coreclient

import (
	"fmt"
	"log"
)

const NEGOTIATION_BUF_SIZE = 1024 * 10

func onConnectNegotiation(remoteAddr string, rwc io.ReadWriteCloser) (io.ReadWriteCloser, error) {

	// encoder := newMessageEncoder(wrc, NEGOTIATION_BUF_SIZE)
	// var wr wireRequest

	return wrc, nil
}
