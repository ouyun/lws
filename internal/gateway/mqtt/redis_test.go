package mqtt

import (
	// "log"
	"testing"
	// "github.com/gomodule/redigo/redis"
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
