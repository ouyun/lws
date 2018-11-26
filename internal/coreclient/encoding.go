package coreclient

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"reflect"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/golang/protobuf/proto"
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
		log.Println("[ERROR] pack msg error", err)
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
	return PackMsg(msg, id)
}

// Pack Frame
// | msg-length(4 bytes) | msg |
func (c *messageEncoder) Pack(msg []byte) ([]byte, error) {
	return Pack(msg)
}

type messageDecoder struct {
	io.Reader
}

func newMessageDecoder(r io.Reader, bufferSize int) *messageDecoder {
	br := bufio.NewReaderSize(r, bufferSize)
	return &messageDecoder{br}
}

func (dec *messageDecoder) Unpack(bytes []byte) (dbp.Msg, proto.Message) {
	return Unpack(bytes)
}

func (dec *messageDecoder) ReadMsg(wr *wireResponse) error {
	var err error
	lenBuf := make([]byte, 4)
	_, err = dec.Read(lenBuf)
	if err != nil {
		log.Println("[ERROR] decoder ReadMsg failed", err)
		return err
	}

	msgLen := binary.BigEndian.Uint32(lenBuf)
	// log.Printf("[DEBUG] received buf len [%d]", msgLen)

	buf := make([]byte, msgLen)
	_, err = io.ReadFull(dec, buf)
	if err != nil {
		log.Println("[ERROR] read full bytes error", err)
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
		dbp.Msg_METHOD,
		dbp.Msg_ERROR:
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
