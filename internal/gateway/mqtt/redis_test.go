package mqtt

import (
	"encoding/hex"
	// "log"
	"reflect"
	"strconv"
	"testing"

	"github.com/gomodule/redigo/redis"
)

func TestNewRedisPool(t *testing.T) {
	pool := GetRedisPool()
	red := pool.Get()
	err := red.Err()
	if err != nil {
		t.Error("conn redis fail")
	}
	defer red.Close()
}

func TestRedis(t *testing.T) {
	pool := GetRedisPool()
	red := pool.Get()
	err := red.Err()
	if err != nil {
		t.Error("conn redis fail")
	}
	defer red.Close()
	cli := CliMap{}
	address, _ := hex.DecodeString("7f937c2f5944f5da2a118cebb067cd2c9c92c75955ce05aa05158a1af28e1607")

	cli.Address = address
	cli.AddressId = uint32(2323)
	cli.ApiKey = address
	cli.ReplyUTXON = uint16(160)
	cli.TopicPrefix = "wawawawaw"
	// _, err = red.Do("SET", "jojo", "COCO")
	// _, err = red.Do("HMSET", redis.Args{}.Add("tsau").AddFlat(cli)...)
	err = SaveToRedis(&red, &cli)
	if err != nil {
		t.Error("save to redis fail")
	}
	cliMap := CliMap{}
	value, err := redis.Values(red.Do("hgetall", strconv.FormatUint(uint64(cli.AddressId), 10)))
	// log.Printf("value %+v! \n", value)
	redis.ScanStruct(value, &cliMap)
	if err != nil {
		t.Error("get redis failed")
	}
	// log.Printf("cli %+v! \n", cli)
	// log.Printf("climap %+v! \n", cliMap)
	if !reflect.DeepEqual(cliMap, cli) {
		t.Error("save redis  failed")
	}
	_, err = redis.String(red.Do("get", hex.EncodeToString(cli.Address)))
	if err != nil {
		t.Error("get address string failed")
	}
	// log.Printf("value %+v! \n", values)
}

// func TestRedis(t *testing.T) {
// 	for index := 0; index < 50; index++ {
// 		go RunRedis(t, index)
// 	}
// 	RunRedis(t, 10000)
// 	// defer redisC.Close()
// }

// func RunRedis(t *testing.T, c int) {
// 	pool := GetRedisPool()
// 	redisC := pool.Get()
// 	err := redisC.Err()
// 	if err != nil {
// 		t.Error("conn redis fail")
// 	}
// 	for i := 0; i < 100; i++ {
// 		_, err := redisC.Do("SET", c, i)
// 		if err != nil {
// 			log.Print("set to err: ", err)
// 			t.Error("set redis failed: ", c)
// 		}
// 		log.Print("set to redis", c)
// 	}
// 	redisC.Do("DEL", "x")
// 	defer redisC.Close()
// }
