package coreclient

import (
	"fmt"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/dbp"
	"io"
	"net"
	"testing"
)

func TestOnConnectSuccess(t *testing.T) {
	serverConn, clientConn := net.Pipe()

	go (func(conn io.ReadWriteCloser) {
		var wr wireResponse
		decoder := newMessageDecoder(conn, 1024)
		if err := decoder.ReadMsg(&wr); err != nil {
			t.Fatalf("ReadMsg failed[%s]", err)
		}
		fmt.Println("Received MsgType: ", wr.MsgType)
		fmt.Println("Received Response: ", wr.Response)

		var wreq wireRequest
		wreq.Request = &dbp.Connected{
			Session: "hahaha",
		}
		encoder := newMessageEncoder(conn, 1024)
		if err := encoder.WriteMsg(&wreq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}
		if err := encoder.Flush(); err != nil {
			t.Fatalf("Write flush failed: [%s]", err)
		}
	})(serverConn)

	_, err := onConnectNegotiation("", clientConn)
	if err != nil {
		t.Fatalf("negotiation failed: [%s]", err)
	}
}

func TestOnConnectFailed(t *testing.T) {
	serverConn, clientConn := net.Pipe()

	go (func(conn io.ReadWriteCloser) {
		var wr wireResponse
		decoder := newMessageDecoder(conn, 1024)
		if err := decoder.ReadMsg(&wr); err != nil {
			t.Fatalf("ReadMsg failed[%s]", err)
		}
		fmt.Println("Received MsgType: ", wr.MsgType)
		fmt.Println("Received Response: ", wr.Response)

		var wreq wireRequest
		wreq.Request = &dbp.Failed{
			Version: []int32{3, 4, 5},
		}
		encoder := newMessageEncoder(conn, 1024)
		if err := encoder.WriteMsg(&wreq); err != nil {
			t.Fatalf("WriteMsg failed[%s]", err)
		}
		if err := encoder.Flush(); err != nil {
			t.Fatalf("Write flush failed: [%s]", err)
		}
	})(serverConn)

	_, err := onConnectNegotiation("", clientConn)
	if err != nil {
		if err.Error() != "coreclient: negotiate failed" {
			t.Fatalf("negotiation non-failed message : [%s]", err)
		}
	} else {
		t.Fatal("negotaion should be failed")
	}
}
