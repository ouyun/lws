package coreclient

import (
	"fmt"
	"io"
	"time"

	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"
)

const VERSION = 1
const CLIENT = "lws"

const NEGOTIATION_BUF_SIZE = 1024 * 10

func onConnectNegotiation(remoteAddr string, rwc io.ReadWriteCloser) (io.ReadWriteCloser, error) {
	var err error
	successChan := make(chan bool, 1)
	failChan := make(chan bool, 1)

	// wait for CONNECT/FAIL message
	go (func(successChan chan bool, failChan chan bool) {
		var wres wireResponse
		decoder := newMessageDecoder(rwc, NEGOTIATION_BUF_SIZE)

		if err = decoder.ReadMsg(&wres); err != nil {
			fmt.Println("read message failed ", err)
			failChan <- true
			return
		}

		if wres.MsgType == dbp.Msg_CONNECTED {
			fmt.Println("received CONNECTED message", wres.Response)
			successChan <- true
			return
		} else if wres.MsgType == dbp.Msg_FAILED {
			fmt.Println("received FAILED message", wres.Response)
			// fmt.Printf("received Failed message [%s]
			failChan <- true
			return
		}

		fmt.Println("received non-CONNECT/FAILED message", wres.Response)
		failChan <- true
		return
	})(successChan, failChan)

	encoder := newMessageEncoder(rwc, NEGOTIATION_BUF_SIZE)
	var wreq wireRequest
	wreq.Request = &dbp.Connect{
		Version: VERSION,
		Client:  CLIENT,
		Session: "",
	}

	if err = encoder.WriteMsg(&wreq); err != nil {
		err = fmt.Errorf("Write CONNECT message failed: [%s]", err)
		return rwc, err
	}
	if err = encoder.Flush(); err != nil {
		err = fmt.Errorf("Write flush failed: [%s]", err)
		return rwc, err
	}

	select {
	case <-successChan:
		fmt.Println("coreclient: negotiate successChan")
		err = nil
	case <-failChan:
		err = fmt.Errorf("coreclient: negotiate failed")
	case <-time.After(time.Second * 20):
		err = fmt.Errorf("coreclient: negotiate timeout")
	}

	return rwc, err
}
