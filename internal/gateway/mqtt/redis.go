package mqtt

import (
	"os"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

type CliMap struct {
	Address     []byte `redis:"Address"`
	AddressId   uint32 `redis:"AddressId"`
	ApiKey      []byte `redis:"ApiKey"`
	TopicPrefix string `redis:"TopicPrefix"`
	ReplyUTXON  uint16 `redis:"ReplyUTXON"`
	Nonce       uint16 `redis:"Nonce"`
}

func NewRedisPool() *redis.Pool {
	address := os.Getenv("REDIS_URL")
	dbOption := redis.DialDatabase(0)
	// pwOption := redis.DialPassword()
	REDIS_MAXIDLE, _ := strconv.Atoi(os.Getenv("REDIS_MAXIDLE"))
	REDIS_MAXACTIVE, _ := strconv.Atoi(os.Getenv("REDIS_MAXACTIVE"))
	redisPool := &redis.Pool{
		MaxIdle:     REDIS_MAXIDLE,
		MaxActive:   REDIS_MAXACTIVE,
		Wait:        true,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address, dbOption)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
	return redisPool
}
