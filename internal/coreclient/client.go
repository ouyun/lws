package coreclient

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	dbp "github.com/lomocoin/lws/internal/coreclient/DBPMsg/go"
)

const version = 1

type CoreClient struct {
	Conn         net.Conn
	w            *bufio.Writer
	r            *bufio.Reader
	netloc       string // for reconnect
	session      string
	isNegotiated bool
}

func (c *CoreClient) Connect(netloc string) error {
	conn, err := net.Dial("tcp", netloc)
	if err != nil {
		log.Println("connect to core serve failed", err)
		return err
	}
	c.Conn = conn
	c.w = bufio.NewWriter(conn)
	c.r = bufio.NewReader(conn)

	return nil
}

func (c *CoreClient) SendMethod(msgType dbp.Msg, msg *proto.Message) error {
	panic("Unimplenmented")
	return nil
}

func (c *CoreClient) generateMsgId() string {
	return "123"
}

func (c *CoreClient) negotiate() error {
	panic("Unimplenmented")
	// 1. package CONNECT pb

	return nil
}

func (c *CoreClient) Send(msg *proto.Message) error {
	panic("Unimplenmented")
	// 1. package CONNECT pb

	return nil
}

func (c *CoreClient) Pack(msg proto.Message) ([]byte, error) {
	var msgType dbp.Msg
	switch interface{}(msg).(type) {
	case *dbp.Connect:
		fmt.Println("dbp.connect detect")
		msgType = dbp.Msg_CONNECT
	case *dbp.Ping:
		msgType = dbp.Msg_PING
	default:
		fmt.Println("===i dont know")
		msgType = -1
	}

	if msgType == -1 {
		log.Println("can not match msgType", msg)
		return nil, errors.New("can not match msgType")
	}

	msgValues := reflect.ValueOf(msg)
	if field := msgValues.Elem().FieldByName("Id"); field.IsValid() {
		field.SetString(c.generateMsgId())
	}

	serializedObject, err := ptypes.MarshalAny(msg)
	if err != nil {
		log.Fatal("could not serialize any field")
	}

	baseMsg := &dbp.Base{
		Msg:    msgType,
		Object: serializedObject,
	}

	serializedBaseMsg, err := proto.Marshal(baseMsg)
	if err != nil {
		log.Fatal("could not serialize msg")
	}

	fmt.Printf("serializedBaseMsg = %+v\n", serializedBaseMsg)

	return serializedBaseMsg, nil
}

func (c *CoreClient) PrependLengthSend(msg []byte) error {
	buf := make([]byte, 4)
	length := len(msg)
	binary.LittleEndian.PutUint32(buf, uint32(length))

	if _, err := c.w.Write(buf); err != nil {
		log.Println("Send failed", err)
		return err
	}

	if _, err := c.w.Write(msg); err != nil {
		log.Println("Send failed", err)
		return err
	}

	c.w.Flush()

	return nil
}
