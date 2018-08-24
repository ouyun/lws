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
	"github.com/satori/go.uuid"
)

const VERSION = 1
const CLIENT = "lws"
const UNKNOWN_MSG_TYPE = -1

type CoreClient struct {
	Conn         net.Conn
	w            *bufio.Writer
	r            *bufio.Reader
	netloc       string // for reconnect
	session      string
	isNegotiated bool
}

// set Conn for CoreClient
func (c *CoreClient) SetConn(conn *net.Conn) error {
	c.Conn = *conn
	c.w = bufio.NewWriter(c.Conn)
	c.r = bufio.NewReader(c.Conn)

	return nil
}

// Connect to Core Wallet Server, dev - not to use the one
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
	id, err := uuid.NewV4()
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		return ""
	}
	return id.String()
}

// Negotiate version, send CONNECT DBP
func (c *CoreClient) negotiate() error {
	// 1. package CONNECT pb
	connect := &dbp.Connect{
		Version: VERSION,
		Session: c.session,
		Client:  CLIENT,
	}

	return c.Send(connect, "")
}

func (c *CoreClient) Send(msg proto.Message, id string) error {
	bytes, err := c.Pack(msg, id)
	if err != nil {
		return err
	}

	return c.PrependLengthSend(bytes)
}

// Pack business struct to serialized BaseMsg bytes
func (c *CoreClient) Pack(msg proto.Message, id string) ([]byte, error) {
	var msgType dbp.Msg
	switch interface{}(msg).(type) {
	case *dbp.Connect:
		fmt.Println("dbp.connect detect")
		msgType = dbp.Msg_CONNECT
	case *dbp.Ping:
		msgType = dbp.Msg_PING
	default:
		fmt.Println("===i dont know")
		msgType = UNKNOWN_MSG_TYPE
	}

	if msgType == UNKNOWN_MSG_TYPE {
		log.Println("can not match msgType", msg)
		return nil, errors.New("can not match msgType")
	}

	msgValues := reflect.ValueOf(msg)
	if field := msgValues.Elem().FieldByName("Id"); field.IsValid() {
		field.SetString(id)
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

func (c *CoreClient) Unpack(bytes []byte) (dbp.Msg, proto.Message) {
	var err error
	baseMsg := &dbp.Base{}

	if err = proto.Unmarshal(bytes, baseMsg); err != nil {
		log.Fatal("unkonwn message received", bytes, err)
	}

	var object proto.Message

	switch baseMsg.Msg {
	case dbp.Msg_CONNECTED:
		object = &dbp.Connect{}
	case dbp.Msg_FAILED:
		object = &dbp.Failed{}
	}

	log.Println("received baseMsg: ", baseMsg)

	err = ptypes.UnmarshalAny(baseMsg.Object, object)
	if err != nil {
		log.Println("unpack Object failed", err)
	}
	return baseMsg.Msg, object
}

func (c *CoreClient) HandleReceivedMsg(bytes []byte) {
	msgType, object := c.Unpack(bytes)
	switch msgType {
	case dbp.Msg_CONNECTED:
	case dbp.Msg_FAILED:
		log.Fatalln("negotiate failed")
	case dbp.Msg_ADDED, dbp.Msg_REMOVED, dbp.Msg_CHANGED:
		// TODO subscribe map
	}
}

func (c *CoreClient) Start() {
	// for {
	// }
}
