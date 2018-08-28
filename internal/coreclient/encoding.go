package coreclient

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	dbp "github.com/lomocoin/lws/internal/coreclient/DBPMsg/go"
)

const UNKNOWN_MSG_TYPE = -1

type wireRequest struct {
	ID      string
	Request interface{}
}

type wireResponse struct {
	ID       string
	Response interface{}
	Error    string
	MsgType  dbp.Msg
}

type messageEncoder struct {
	*bufio.Writer
}

func newMessageEncoder(w io.Writer, bufferSize int) *messageEncoder {
	bw := bufio.NewWriterSize(w, bufferSize)
	return &messageEncoder{bw}
}

// WriteMsg
func (enc *messageEncoder) WriteMsg(wr *wireRequest) error {
	Request := wr.Request
	msg, ok := Request.(proto.Message)
	if !ok {
		err := fmt.Errorf("coreclient bad request %s", wr.Request)
		return err
	}
	msgBuf, err := enc.PackMsg(msg, wr.ID)
	if err != nil {
		log.Println("pack msg error", err)
		return err
	}

	buf, err := enc.Pack(msgBuf)
	if err != nil {
		log.Println("pack error", err)
		return err
	}

	_, err = enc.Write(buf)
	return err
}

// Pack business struct to serialized BaseMsg bytes
func (enc *messageEncoder) PackMsg(msg proto.Message, id string) ([]byte, error) {
	var msgType dbp.Msg
	switch interface{}(msg).(type) {
	case *dbp.Connect:
		fmt.Println("dbp.connect detect")
		msgType = dbp.Msg_CONNECT
	case *dbp.Ping:
		msgType = dbp.Msg_PING
	case *dbp.Connected:
		msgType = dbp.Msg_CONNECTED
	case *dbp.Failed:
		msgType = dbp.Msg_FAILED
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

// Pack Frame
// | msg-length(4 bytes) | msg |
func (c *messageEncoder) Pack(msg []byte) ([]byte, error) {
	buflen := len(msg) + 4
	fmt.Printf("buflen = %+v\n", buflen)

	msglen := len(msg)
	fmt.Printf("msglen = %+v\n", msglen)

	buf := make([]byte, buflen)
	binary.LittleEndian.PutUint32(buf, uint32(msglen))

	// fmt.Printf("before content buf = %+v\n", buf)
	copy(buf[4:], msg)

	fmt.Printf("buf = %+v\n", buf)
	return buf, nil
}

type messageDecoder struct {
	io.Reader
}

func newMessageDecoder(r io.Reader, bufferSize int) *messageDecoder {
	br := bufio.NewReaderSize(r, bufferSize)
	return &messageDecoder{br}
}

func (dec *messageDecoder) Unpack(bytes []byte) (dbp.Msg, proto.Message) {
	var err error
	baseMsg := &dbp.Base{}

	if err = proto.Unmarshal(bytes, baseMsg); err != nil {
		log.Fatal("unkonwn message received", bytes, err)
	}

	var object proto.Message

	// TODO mark empty_ID messages
	switch baseMsg.Msg {
	case dbp.Msg_CONNECT:
		object = &dbp.Connect{}
	case dbp.Msg_CONNECTED:
		object = &dbp.Connected{}
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

func (dec *messageDecoder) ReadMsg(wr *wireResponse) error {
	var err error
	lenBuf := make([]byte, 4)
	_, err = dec.Read(lenBuf)
	if err != nil {
		log.Println("dec ReadMsg failed", err)
		return err
	}

	msgLen := binary.LittleEndian.Uint32(lenBuf)

	buf := make([]byte, msgLen)
	_, err = io.ReadFull(dec, buf)
	if err != nil {
		log.Println("read full bytes error", err)
		return err
	}

	msgType, msg := dec.Unpack(buf)
	wr.MsgType = msgType
	wr.Response = msg

	// OR WE CAN USE type.assetion, but we need to map all the types manually
	// if rMsg, ok := msg.(*dbp.Ping); ok {
	// 	wr.Id = rMsg.Id
	// }
	switch msgType {
	// dbp.Msg_ERROR
	case dbp.Msg_PING,
		dbp.Msg_RESULT,
		dbp.Msg_PONG,
		dbp.Msg_SUB,
		dbp.Msg_UNSUB,
		dbp.Msg_NOSUB,
		dbp.Msg_READY,
		dbp.Msg_ADDED,
		dbp.Msg_CHANGED,
		dbp.Msg_REMOVED,
		dbp.Msg_METHOD:
		// for those msg types who has Id field
		msgValues := reflect.ValueOf(msg)
		if field := msgValues.Elem().FieldByName("Id"); field.IsValid() {
			wr.ID = field.String()
		}
	default:
	}
	return nil
}

// // ReadN reads n bytes
// func ReadN(r io.Reader, n int) ([]byte, error) {
// 	buf := make([]byte, n)
// 	_, err := io.ReadFull(r, buf)
// 	return buf, err
// }
