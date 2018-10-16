package mqtt

import (
	// "encoding/hex"
	// "flag"
	// "log"
	"reflect"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := GetProgram()
	// if err != nil {
	// 	t.Errorf("new  client failed")
	// }
	if client == nil {
		t.Errorf("new  client failed")
	}
	client2 := GetProgram()
	// if err != nil {
	// 	t.Errorf("new  client failed")
	// }
	if client2 == nil {
		t.Errorf("new  client failed")
	}
	if !reflect.DeepEqual(client, client2) {
		t.Errorf("get two diffrent client")
	}
	client2.Stop()
}
