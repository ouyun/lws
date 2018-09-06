package coreclient

import (
	"fmt"
	"io"
	"log"
	"net"
	"testing"
	"time"

	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"
)

func sub(c *Client, name string, t *testing.T) {
	sub := &dbp.Sub{
		Name: "block",
	}

	subscription, msg, err := c.Subscribe(sub)
	if err != nil {
		t.Fatalf("CallAsync error [%s]", err)
	}

	go (func(closeChan *chan struct{}, notificationChan *chan *Notification) {
		for {
			select {
			case <-*closeChan:
				log.Printf("[%s]: client handle close chan", name)
				return
			case noti := <-*notificationChan:
				log.Printf("[%s]: recevied notification [%s]", name, noti)
			}
		}
	})(&subscription.CloseChan, &subscription.NotificationChan)

	var subId string

	switch msg.(type) {
	case *dbp.Ready:
		ready := msg.(*dbp.Ready)
		subId = ready.Id
	case *dbp.Nosub:
		nosub := msg.(*dbp.Nosub)
		c.deleteSubscription(nosub.Id)
	default:
		t.Fatalf("unexpected response type [%s]", msg)
	}

	select {
	case <-subscription.CloseChan:
		fmt.Printf("[%s]: close sub", name)
	case <-time.After(time.Second * 3):
		if subId != "" {
			log.Printf("[%s]: time's up, delete subscription", name)
			c.deleteSubscription(subId)
			// in real world, we need to send unsub msg before calling deleteSubscription
		}
	}
}

func TestSubscribe(t *testing.T) {
	serverConn, clientConn := net.Pipe()

	go (func(conn io.ReadWriteCloser) {
		var wr wireResponse
		decoder := newMessageDecoder(conn, 1024)
		if err := decoder.ReadMsg(&wr); err != nil {
			t.Fatalf("ReadMsg failed[%s]", err)
		}
		fmt.Println("Received MsgType: ", wr.MsgType)
		fmt.Println("Received client request: ", wr.Response)

		response := wr.Response

		subMsg, ok := response.(*dbp.Sub)
		if !ok {
			t.Fatalf("received non-ping message type:[%s] content:[%s] ", wr.MsgType, wr.Response)
		}

		var wreq wireRequest
		encoder := newMessageEncoder(conn, 1024)

		// send ready
		wreq.ID = subMsg.Id
		wreq.Request = &dbp.Ready{}
		if err := encoder.WriteMsg(&wreq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}
		// if err := encoder.Flush(); err != nil {
		// 	t.Fatalf("Write flush failed: [%s]", err)
		// }

		// send added twice
		wreq.ID = subMsg.Id
		wreq.Request = &dbp.Added{
			Name: "test1",
		}
		if err := encoder.WriteMsg(&wreq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}

		// time.Sleep(time.Second)
		wreq.ID = subMsg.Id
		wreq.Request = &dbp.Added{
			Name: "test1",
		}
		if err := encoder.WriteMsg(&wreq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}

		if err := encoder.Flush(); err != nil {
			t.Fatalf("Write flush failed: [%s]", err)
		}
	})(serverConn)

	c := &Client{
		Addr: "whatever",
		Dial: func(addr string) (conn io.ReadWriteCloser, err error) {
			return clientConn, nil
		},
		LogDebug: log.Printf,
		// OnConnect: onConnectNegotiation,
	}

	c.Start()

	sub(c, "single block", t)
}

func TestMultiSubscribe(t *testing.T) {
	serverConn, clientConn := net.Pipe()

	go (func(conn io.ReadWriteCloser) {
		var wr wireResponse

		decoder := newMessageDecoder(conn, 1024)
		if err := decoder.ReadMsg(&wr); err != nil {
			t.Fatalf("ReadMsg failed[%s]", err)
		}
		fmt.Println("Received MsgType: ", wr.MsgType)
		fmt.Println("Received client request: ", wr.Response)

		// response := wr.Response

		subMsg, ok := wr.Response.(*dbp.Sub)
		if !ok {
			t.Fatalf("received non-ping message type:[%s] content:[%s] ", wr.MsgType, wr.Response)
		}

		if err := decoder.ReadMsg(&wr); err != nil {
			t.Fatalf("ReadMsg failed[%s]", err)
		}
		fmt.Println("Received MsgType: ", wr.MsgType)
		fmt.Println("Received client request: ", wr.Response)

		// response2 := wr.Response

		subMsg2, ok := wr.Response.(*dbp.Sub)
		if !ok {
			t.Fatalf("received non-ping message type:[%s] content:[%s] ", wr.MsgType, wr.Response)
		}

		var wreq wireRequest
		encoder := newMessageEncoder(conn, 1024)

		// send ready 1
		wreq.ID = subMsg.Id
		wreq.Request = &dbp.Ready{}
		if err := encoder.WriteMsg(&wreq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}
		if err := encoder.Flush(); err != nil {
			t.Fatalf("Write flush failed: [%s]", err)
		}

		// send ready 2
		wreq.ID = subMsg2.Id
		wreq.Request = &dbp.Ready{}
		if err := encoder.WriteMsg(&wreq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}
		if err := encoder.Flush(); err != nil {
			t.Fatalf("Write flush failed: [%s]", err)
		}

		// send added
		wreq.ID = subMsg.Id
		wreq.Request = &dbp.Added{
			Name: "test1",
		}
		if err := encoder.WriteMsg(&wreq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}

		if err := encoder.Flush(); err != nil {
			t.Fatalf("Write flush failed: [%s]", err)
		}
		time.Sleep(time.Microsecond * 500)

		// send added msg2
		wreq.ID = subMsg2.Id
		wreq.Request = &dbp.Added{
			Name: "msg2 test2",
		}
		if err := encoder.WriteMsg(&wreq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}

		// send added
		// time.Sleep(time.Second)
		wreq.ID = subMsg.Id
		wreq.Request = &dbp.Added{
			Name: "test2",
		}
		if err := encoder.WriteMsg(&wreq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}

		if err := encoder.Flush(); err != nil {
			t.Fatalf("Write flush failed: [%s]", err)
		}
	})(serverConn)

	c := &Client{
		Addr: "whatever",
		Dial: func(addr string) (conn io.ReadWriteCloser, err error) {
			return clientConn, nil
		},
		// OnConnect: onConnectNegotiation,
	}

	c.Start()

	time.Sleep(time.Second)

	go sub(c, "block1", t)
	go sub(c, "block2", t)

	time.Sleep(time.Second * 4)
}
