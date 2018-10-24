package mqtt

import (
	// "encoding/hex"
	// "flag"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"log"
	"reflect"
	"strconv"
	"strings"
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

func TestGetUserByAddress(t *testing.T) {

	code1 := "1 21 31 223 54 184 6 54 94 213 182 216 67 173 88 68 174 66 80 132 46 27 115 69 27 63 250 22 209 215 234 10 178"
	codeArr1 := strings.Split(code1, " ")
	log.Printf("len : %+v", len(codeArr1))
	codeAdd1 := make([]byte, 33)
	log.Printf("len : %+v", codeArr1)
	for index := 0; index < len(codeArr1); index++ {
		value, _ := strconv.Atoi(codeArr1[index])
		codeAdd1[index] = byte(value)
	}
	pool := GetRedisPool()
	redisConn := pool.Get()
	connection := db.GetConnection()

	user := model.User{}
	cliMap := CliMap{}

	defer redisConn.Close()
	err := GetUserByAddress(codeAdd1, connection, &redisConn, &user, &cliMap)
	if err != nil {
		log.Printf("get err : %+v", err)
	}
	log.Printf("get user : %+v", user)
	log.Printf("get cliMap : %+v", cliMap)
}

func TestStructField(t *testing.T) {
	var utxoList UTXO
	log.Printf("utxoList : %+v", utxoList)
}
