package coreclient

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

const VERSION = 1
const CLIENT = "lws"

const NEGOTIATION_BUF_SIZE = 1024 * 10
const REASON_SESSION_INVALID = "002"

func (c *Client) negotiate(rwc io.ReadWriteCloser) error {
	var err error
	successChan := make(chan string, 1)
	failChan := make(chan string, 1)

	// wait for CONNECT/FAIL message
	go (func(successChan chan string, failChan chan string) {
		var wres wireResponse
		decoder := newMessageDecoder(rwc, NEGOTIATION_BUF_SIZE)

		if err = decoder.ReadMsg(&wres); err != nil {
			log.Println("read message failed ", err)
			failChan <- ""
			return
		}

		if wres.MsgType == dbp.Msg_CONNECTED {
			if connected, ok := wres.Response.(*dbp.Connected); ok {
				log.Printf("received connected session: [%s]", connected.Session)
				successChan <- connected.Session
			}
			return
		} else if wres.MsgType == dbp.Msg_FAILED {
			if failed, ok := wres.Response.(*dbp.Failed); ok {
				// fmt.Println("received FAILED message", failed)
				log.Printf("received connected reason: [%s]", failed.Reason)
				failChan <- failed.Reason
			}
			return
		}

		fmt.Println("received non-CONNECT/FAILED message", wres.Response)
		failChan <- ""
		return
	})(successChan, failChan)

	encoder := newMessageEncoder(rwc, NEGOTIATION_BUF_SIZE)
	var wreq wireRequest

	if c.session == "" {
		wreq.Request = &dbp.Connect{
			Version: VERSION,
			Client:  CLIENT,
			Session: "",
			Udata:   generateUdata(),
		}
	} else {
		wreq.Request = &dbp.Connect{
			Session: c.session,
		}
	}
	log.Printf("requrest connected session [%s]", c.session)

	if err = encoder.WriteMsg(&wreq); err != nil {
		err = fmt.Errorf("Write CONNECT message failed: [%s]", err)
		return err
	}
	if err = encoder.Flush(); err != nil {
		err = fmt.Errorf("Write flush failed: [%s]", err)
		return err
	}

	select {
	case newSession := <-successChan:
		fmt.Println("coreclient: negotiate successChan")
		if c.session != newSession {
			// re-subscribe
			log.Printf("coreclient: session changed [%s] -> [%s]", c.session, newSession)
			c.session = newSession
			go c.Resubscribe()
		}
		err = nil
	case reason := <-failChan:
		err = fmt.Errorf("coreclient: negotiate failed")
		if reason == REASON_SESSION_INVALID {
			c.session = ""
			err = c.negotiate(rwc)
		}
	case <-time.After(time.Second * 20):
		err = fmt.Errorf("coreclient: negotiate timeout")
	}

	return err
}

/* udata { forkid, } */
func generateUdata() map[string]*any.Any {
	udata := make(map[string]*any.Any)
	forkid := &lws.ForkID{
		Ids: []string{os.Getenv("FORK_ID")},
	}
	forkidAny, err := ptypes.MarshalAny(forkid)
	if err != nil {
		log.Fatalf("pack forkid any failed [%s]", err)
	}
	udata["forkid"] = forkidAny
	return udata
}
