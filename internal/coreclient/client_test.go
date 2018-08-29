package coreclient

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"
)

func TestPing(t *testing.T) {
	serverConn, clientConn := net.Pipe()

	go (func(conn io.ReadWriteCloser) {
		var wr wireResponse
		decoder := newMessageDecoder(conn, 1024)
		if err := decoder.ReadMsg(&wr); err != nil {
			t.Fatalf("ReadMsg failed[%s]", err)
		}
		fmt.Println("Received MsgType: ", wr.MsgType)
		fmt.Println("Received Response: ", wr.Response)

		response := wr.Response

		ping, ok := response.(*dbp.Ping)
		if !ok {
			t.Fatalf("received non-ping message type:[%s] content:[%s] ", wr.MsgType, wr.Response)
		}

		var wreq wireRequest
		wreq.ID = ping.Id
		wreq.Request = &dbp.Pong{}
		encoder := newMessageEncoder(conn, 1024)
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

	m, err := c.CallAsync(&dbp.Ping{})
	if err != nil {
		t.Fatalf("CallAsync error [%s]", err)
	}

	select {
	case <-time.After(c.RequestTimeout):
		t.Fatal("call async timeout")
	case <-m.Done:
		mRes := m.Response
		_, ok := mRes.(*dbp.Pong)
		if !ok {
			t.Fatal("receved non-pong", mRes)
		}
	}
}

func TestConnectAndPing(t *testing.T) {
	serverConn, clientConn := net.Pipe()

	go (func(conn io.ReadWriteCloser) {
		var wr wireResponse
		var wreq wireRequest
		decoder := newMessageDecoder(conn, 1024)
		encoder := newMessageEncoder(conn, 1024)

		// receive CONNECT
		if err := decoder.ReadMsg(&wr); err != nil {
			t.Fatalf("ReadMsg failed[%s]", err)
		}
		fmt.Println("Received MsgType: ", wr.MsgType)
		fmt.Println("Received Response: ", wr.Response)

		// write CONNECTED
		wreq.Request = &dbp.Connected{
			Session: "hahaha",
		}
		if err := encoder.WriteMsg(&wreq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}
		if err := encoder.Flush(); err != nil {
			t.Fatalf("Write flush failed: [%s]", err)
		}

		// receive PING
		if err := decoder.ReadMsg(&wr); err != nil {
			t.Fatalf("ReadMsg failed[%s]", err)
		}
		fmt.Println("Received MsgType: ", wr.MsgType)
		fmt.Println("Received Response: ", wr.Response)

		response := wr.Response

		ping, ok := response.(*dbp.Ping)
		if !ok {
			t.Fatalf("received non-ping message type:[%s] content:[%s] ", wr.MsgType, wr.Response)
		}

		wreq.ID = ping.Id
		wreq.Request = &dbp.Pong{}
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
		OnConnect: onConnectNegotiation,
	}

	c.Start()

	m, err := c.CallAsync(&dbp.Ping{})
	if err != nil {
		t.Fatalf("CallAsync error [%s]", err)
	}

	select {
	case <-time.After(c.RequestTimeout):
		t.Fatal("call async timeout")
	case <-m.Done:
		mRes := m.Response
		_, ok := mRes.(*dbp.Pong)
		if !ok {
			t.Fatal("receved non-pong", mRes)
		}
	}
}
