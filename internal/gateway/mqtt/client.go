package mqtt

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/internal/db/model"
	"github.com/lomocoin/lws/internal/gateway/crypto"
)

var (
	clientHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("TOPIC: %s\n", msg.Topic())
		// DecodePayload(msg.Payload())

	}
	serviceReqHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		// TODO ：
		s := ServicePayload{}
		cliMap := CliMap{}
		user := model.User{}
		var pubKey []byte
		err := DecodePayload(msg.Payload(), &s)
		if err != nil {
			log.Printf("err: %+v\n", err)
			ReplyServiceReq(&client, 16, &s, &user, pubKey)
		}
		pubKey = PayloadToUser(&user, &s)

		// 连接 redis && db
		pool := NewRedisPool()
		redisConn := pool.Get()
		defer redisConn.Close()
		connection := db.GetConnection()
		defer connection.Close()
		if err != nil {
			log.Printf("err: %+v\n", err)
			ReplyServiceReq(&client, 16, &s, &user, pubKey[:])
		}

		// 查询 redis
		// 没有 -> 查询 数据库
		found := connection.Where("address = ?", []byte(s.Address)).First(&user).RecordNotFound()
		cliMap.Address = user.Address
		cliMap.ForkList = user.ForkList
		cliMap.ForkNum = user.ForkNum
		cliMap.TopicPrefix = user.TopicPrefix
		cliMap.AddressId = user.AddressId
		cliMap.ApiKey = user.ApiKey
		cliMap.ReplyUTXON = user.ReplyUTXON
		if !found {
			// update user
			err = connection.Save(&user).Error // update user
			if err != nil {
				// fail
				log.Printf("err: %+v\n", err)
				ReplyServiceReq(&client, 16, &s, &user, pubKey[:])
				return
			}
		} else {
			// save user
			err = SaveUser(connection, &user)
			if err != nil {
				// fail
				log.Printf("err: %+v\n", err)
				ReplyServiceReq(&client, 16, &s, &user, pubKey[:])
				return
			}
		}
		// 保存到 redis
		err = SaveToRedis(&redisConn, &cliMap)
		if err != nil {
			// fail
			log.Printf("err: %+v\n", err)
			ReplyServiceReq(&client, 16, &s, &user, pubKey[:])
			return
		}
		//TODO: 检查分支
		ReplyServiceReq(&client, 0, &s, &user, pubKey)
	}

	syncReqHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		// TODO ：
		s := SyncPayload{}
		err := DecodePayload(msg.Payload(), &s)
		log.Printf("SyncPayloads: %+v\n", s)
		if err != nil {
			log.Printf("err: %+v\n", err)
		}
		// 连接 redis
		pool := NewRedisPool()
		redisConn := pool.Get()
		defer redisConn.Close()
		exists, err := redis.Bool(redisConn.Do("EXISTS", s.AddressId))
		if err != nil {
			return
		}
		if exists {
			// 检查分支
		}
		// TODO：UTXO相关

	}
	uTXOAbortReqHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		a := AbortPayload{}
		err := DecodePayload(msg.Payload(), &a)
		log.Printf("AbortPayload: %+v\n", a)
		if err != nil {
			log.Printf("err: %+v\n", err)
		}
		// 连接 redis
		pool := NewRedisPool()
		redisConn := pool.Get()
		defer redisConn.Close()
		exists, err := redis.Bool(redisConn.Do("EXISTS", a.AddressId))
		if err != nil {
			// fail
			return
		}
		if exists {
			//取消utxo update
		}
		// TODO：UTXO相关

	}
	sendTxReqReqHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		s := SendTxPayload{}
		err := DecodePayload(msg.Payload(), &s)
		log.Printf("AbortPayload: %+v\n", s)
		if err != nil {
			log.Printf("err: %+v\n", err)
		}
		// 连接 redis
		pool := NewRedisPool()
		redisConn := pool.Get()
		defer redisConn.Close()
		exists, err := redis.Bool(redisConn.Do("EXISTS", s.AddressId))
		if err != nil {
			// fail
			return
		}
		if exists {
			//取消utxo update
		}
	}
)

type Service interface {
	Init() error
	Start() error
	Stop() error
	Publish(string, byte, bool, []byte) error
	Subscribe(string, byte, mqtt.MessageHandler) error
}

type message struct {
	qos      byte
	retained bool
	topic    string
	payload  []byte
}

var msgChan = make(chan os.Signal, 1)

func Run(service Service) error {
	if err := service.Init(); err != nil {
		return err
	}
	if err := service.Start(); err != nil {
		return err
	}
	signal.Notify(msgChan, os.Interrupt, os.Kill)
	<-msgChan
	return service.Stop()
}

func Interrupt() {
	msgChan <- os.Interrupt
}

type Program struct {
	Id     string
	Client mqtt.Client
	isLws  bool
	subs   []string
}

// start client
func (p *Program) Start() error {
	fmt.Printf("client %+v start\n", p.Id)
	if token := p.Client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if p.isLws {
		p.Subscribe("LWS/lws/ServiceReq", 0, serviceReqHandler)
		p.Subscribe("LWS/lws/SyncReq", 1, syncReqHandler)
		p.Subscribe("LWS/lws/UTXOAbort", 1, uTXOAbortReqHandler)
		p.Subscribe("LWS/lws/SendTxReq", 1, sendTxReqReqHandler)
	}
	return nil
}

// init client
func (p *Program) Init() error {
	// mqtt.DEBUG = log.New(os.Stdout, "", 0)
	// mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker(os.Getenv("MQTT_URL")).SetClientID("lws")
	opts.SetKeepAlive(2 * time.Second)
	if p.isLws {
		opts.SetDefaultPublishHandler(serviceReqHandler)
	} else {
		opts.SetDefaultPublishHandler(clientHandler)
	}

	opts.SetPingTimeout(1 * time.Second)
	p.Client = mqtt.NewClient(opts)
	return nil
}

// Stop clients
func (p *Program) Stop() error {
	// fmt.Println("application is end.")
	// if token := p.Client.Unsubscribe("DEVICE01/fnfn/ServiceReply"); token.Wait() && token.Error() != nil {
	// 	fmt.Println(token.Error())
	// 	os.Exit(1)
	// }
	p.Client.Disconnect(250)
	return nil
}

// publish topic
func (p *Program) Publish(topic string, qos byte, retained bool, msg []byte) error {
	if token := p.Client.Publish(topic, qos, retained, msg); token.Wait() && token.Error() != nil {
		token.Wait()
	}
	return nil
}

// subscribe topic
func (p *Program) Subscribe(topic string, qos byte, handler mqtt.MessageHandler) error {
	if token := p.Client.Subscribe(topic, qos, handler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}
	return nil
}

func SaveUser(conn *gorm.DB, user *model.User) (err error) {
	conn.NewRecord(user)
	conn.Create(user)
	if conn.NewRecord(user) {
		// fail
		err = errors.New("save user fail")
		return err
	}
	return err
}

func SaveToRedis(conn *redis.Conn, cliMap *CliMap) (err error) {
	_, err = (*conn).Do("HMSET", redis.Args{}.Add(
		hex.EncodeToString(cliMap.Address)).AddFlat(cliMap)...)
	if err != nil {
		return err
	}
	_, err = (*conn).Do("SET", cliMap.AddressId, cliMap.Address)
	return err
}

func ReplyServiceReq(client *mqtt.Client, err int, s *ServicePayload, user *model.User, pubKey []byte) {
	reply := ServiceReply{}
	reply.Nonce = s.Nonce
	reply.Version = s.Version
	reply.Error = byte(uint8(err))
	if err == 0 {
		reply.AddressId = user.AddressId
		reply.ForkBitmap = uint64(2553)
		reply.ApiKeySeed = string(pubKey[:])
	}
	result, errs := GenerateService(reply)
	if errs != nil {
		log.Printf("err: %+v\n", err)
	}
	t := s.TopicPrefix + "/fnfn/ServiceReply"
	(*client).Publish(t, 0, false, result)
}

func PayloadToUser(user *model.User, s *ServicePayload) []byte {
	user.Address = []byte(s.Address)
	user.ForkList = s.ForkList
	user.ForkNum = s.ForkNum
	user.TopicPrefix = s.TopicPrefix
	user.TimeStamp = s.TimeStamp
	user.ReplyUTXON = s.ReplyUTXON

	pubKey, privKey, _ := crypto.GenerateKeyPair(nil)
	var address crypto.PublicKey
	copy(address[:], []byte(user.Address))
	apiKey := crypto.GenerateApiKey(&privKey, &address)
	user.ApiKey = apiKey[:]
	return pubKey[:]
}
