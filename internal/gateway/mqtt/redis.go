package mqtt

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
)

type CliMap struct {
	Address     []byte `redis:"Address"`
	AddressId   uint32 `redis:"AddressId"`
	ApiKey      []byte `redis:"ApiKey"`
	TopicPrefix string `redis:"TopicPrefix"`
	ForkNum     uint8  `redis:"ForkNum"`
	ForkList    string `redis:"ForkList"`
	ReplyUTXON  uint16 `redis:"ReplyUTXON"`
	Nonce       uint16 `redis:"Nonce"`
}

func NewRedisPool() *redis.Pool {
	if err := godotenv.Load(os.ExpandEnv("$GOPATH/src/github.com/lomocoin/lws/.env")); err != nil {
		log.Println("no .env file found, will try to use native environment variables")
	}

	address := os.Getenv("REDIS_URL")
	log.Printf("address: %+s", address)
	dbOption := redis.DialDatabase(0)
	// pwOption := redis.DialPassword()
	// readTimeout := redis.DialReadTimeout(time.Second)
	// writeTimeout := redis.DialWriteTimeout(time.Second * )
	// conTimeout := redis.DialConnectTimeout(time.Second * time.Duration(redisConf.ConTimeout))
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
