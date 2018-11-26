package coreclient

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/dbp"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

// Pack business struct to serialized BaseMsg bytes
func PackMsg(msg proto.Message, id string) ([]byte, error) {
	var msgType dbp.Msg
	switch interface{}(msg).(type) {
	case *dbp.Connect:
		fmt.Println("[INFO] dbp.connect detect")
		msgType = dbp.Msg_CONNECT
	case *dbp.Connected:
		msgType = dbp.Msg_CONNECTED
	case *dbp.Failed:
		msgType = dbp.Msg_FAILED
	case *dbp.Ping:
		msgType = dbp.Msg_PING
	case *dbp.Pong:
		msgType = dbp.Msg_PONG
	case *dbp.Sub:
		msgType = dbp.Msg_SUB
	case *dbp.Unsub:
		msgType = dbp.Msg_UNSUB
	case *dbp.Nosub:
		msgType = dbp.Msg_NOSUB
	case *dbp.Ready:
		msgType = dbp.Msg_READY
	case *dbp.Added:
		msgType = dbp.Msg_ADDED
	case *dbp.Changed:
		msgType = dbp.Msg_CHANGED
	case *dbp.Removed:
		msgType = dbp.Msg_REMOVED
	case *dbp.Method:
		msgType = dbp.Msg_METHOD
	case *dbp.Result:
		msgType = dbp.Msg_RESULT
	case *dbp.Error:
		msgType = dbp.Msg_ERROR
	default:
		fmt.Println("===i dont know")
		msgType = UNKNOWN_MSG_TYPE
	}

	if msgType == UNKNOWN_MSG_TYPE {
		log.Println("[WARN] can not match msgType", msg)
		return nil, errors.New("can not match msgType")
	}

	msgValues := reflect.ValueOf(msg)
	if field := msgValues.Elem().FieldByName("Id"); field.IsValid() {
		field.SetString(id)
	}

	serializedObject, err := ptypes.MarshalAny(msg)
	if err != nil {
		log.Fatal("[ERROR] could not serialize any field")
	}

	baseMsg := &dbp.Base{
		Msg:    msgType,
		Object: serializedObject,
	}

	serializedBaseMsg, err := proto.Marshal(baseMsg)
	if err != nil {
		log.Fatal("[ERROR] could not serialize msg")
	}

	// log.Printf("pack baseMsg.Msg = %+v\n", baseMsg.Msg)

	return serializedBaseMsg, nil
}

// Pack Frame
// | msg-length(4 bytes) | msg |
func Pack(msg []byte) ([]byte, error) {
	buflen := len(msg) + 4
	// fmt.Printf("buflen = %+v\n", buflen)

	msglen := len(msg)
	// fmt.Printf("msglen = %+v\n", msglen)

	buf := make([]byte, buflen)
	binary.BigEndian.PutUint32(buf, uint32(msglen))

	// fmt.Printf("before content buf = %+v\n", buf)
	copy(buf[4:], msg)

	// fmt.Printf("buf = %+v\n", buf)
	return buf, nil
}

func Unpack(bytes []byte) (dbp.Msg, proto.Message) {
	var err error
	baseMsg := &dbp.Base{}

	if err = proto.Unmarshal(bytes, baseMsg); err != nil {
		log.Fatal("[ERROR] unkonwn message received", bytes, err)
	}

	var object proto.Message

	switch baseMsg.Msg {
	case dbp.Msg_CONNECT:
		object = &dbp.Connect{}
	case dbp.Msg_CONNECTED:
		object = &dbp.Connected{}
	case dbp.Msg_FAILED:
		object = &dbp.Failed{}
	case dbp.Msg_PING:
		object = &dbp.Ping{}
	case dbp.Msg_PONG:
		object = &dbp.Pong{}
	case dbp.Msg_SUB:
		object = &dbp.Sub{}
	case dbp.Msg_UNSUB:
		object = &dbp.Unsub{}
	case dbp.Msg_NOSUB:
		object = &dbp.Nosub{}
	case dbp.Msg_READY:
		object = &dbp.Ready{}
	case dbp.Msg_ADDED:
		object = &dbp.Added{}
	case dbp.Msg_CHANGED:
		object = &dbp.Changed{}
	case dbp.Msg_REMOVED:
		object = &dbp.Removed{}
	case dbp.Msg_METHOD:
		object = &dbp.Method{}
	case dbp.Msg_RESULT:
		object = &dbp.Result{}
	case dbp.Msg_ERROR:
		object = &dbp.Error{}
	}

	// log.Println("[DEBUG] received baseMsg: ", baseMsg.Msg)

	err = ptypes.UnmarshalAny(baseMsg.Object, object)
	if err != nil {
		log.Println("[ERROR] packager: unpack Object failed", err)
	}
	return baseMsg.Msg, object
}
