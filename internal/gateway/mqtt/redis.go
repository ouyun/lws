package mqtt

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

type CliMap struct {
	Address     []byte `redis:"Address"`
	AddressId   uint32 `redis:"AddressId"`
	ApiKey      []byte `redis:"ApiKey"`
	TopicPrefix string `redis:"TopicPrefix"`
	ForkNum     uint8  `redis:"ForkNum"`
	ForkList    string `redis:"ForkList"`
	ReplyUTXON  uint16 `redis:"ReplyUTXON"`
}

func NewRedisPool() *redis.Pool {

	address := "127.0.0.1:6379"
	dbOption := redis.DialDatabase(0)
	// pwOption := redis.DialPassword()
	readTimeout := redis.DialReadTimeout(time.Second)
	// writeTimeout := redis.DialWriteTimeout(time.Second * )
	// conTimeout := redis.DialConnectTimeout(time.Second * time.Duration(redisConf.ConTimeout))

	redisPool := &redis.Pool{
		MaxIdle:     100,
		MaxActive:   50,
		Wait:        true,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address, dbOption,
				readTimeout)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
	return redisPool
}
