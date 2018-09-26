package mqtt

import (
	"encoding/hex"
	"errors"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/gateway/crypto"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
)

var clientHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("TOPIC: %s\n", msg.Topic())
	// DecodePayload(msg.Payload())
}

type Service interface {
	Init()
	Start() error
	Stop() error
	Publish(string, byte, bool, []byte) error
	Subscribe(string, byte, mqtt.MessageHandler) error
}

var msgChan = make(chan os.Signal, 1)

func Run(service Service) error {
	service.Init()
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
	if token := p.Client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("conn mqtt broker failed : %+v \n", token.Error())
		err := errors.New("conn mqtt broker fail")
		return err
	}
	if p.isLws {
		p.Subscribe("LWS/lws/ServiceReq", 0, serviceReqHandler)
		p.Subscribe("LWS/lws/SyncReq", 1, syncReqHandler)
		p.Subscribe("LWS/lws/UTXOAbort", 1, uTXOAbortReqHandler)
		p.Subscribe("LWS/lws/SendTxReq", 1, sendTxReqReqHandler)
	}
	log.Printf("client start successed!")
	return nil
}

// init client
func (p *Program) Init() {
	// mqtt.DEBUG = log.New(os.Stdout, "", 0)
	// mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker(os.Getenv("MQTT_URL")).SetClientID(p.Id)
	opts.SetKeepAlive(2 * time.Second)
	if p.isLws {
		opts.SetDefaultPublishHandler(serviceReqHandler)
	} else {
		opts.SetDefaultPublishHandler(clientHandler)
	}
	opts.SetConnectTimeout(10 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	p.Client = mqtt.NewClient(opts)
}

// Stop clients
func (p *Program) Stop() error {
	if p.Client.IsConnected() {
		p.Client.Disconnect(250)
		return nil
	}
	return errors.New("client did not conn broker")
}

// publish topic
func (p *Program) Publish(topic string, qos byte, retained bool, msg []byte) error {
	token := p.Client.Publish(topic, qos, retained, msg)
	if token.Wait() && token.Error() != nil {
		log.Printf("publish get err : %s", token.Error())
	}
	return token.Error()
}

// subscribe topic
func (p *Program) Subscribe(topic string, qos byte, handler mqtt.MessageHandler) error {
	token := p.Client.Subscribe(topic, qos, handler)
	if token.Wait() && token.Error() != nil {
		log.Printf("subscribe err : %s", token.Error())
	}
	return token.Error()
}

// create user
func SaveUser(conn *gorm.DB, user *model.User) (err error) {
	conn.NewRecord(user)
	conn.Create(user)
	if conn.NewRecord(user) {
		// fail
		err = errors.New("save user failed")
	}
	return err
}

// update redis
func updateRedis(conn *redis.Conn, cliMap *CliMap, field string, value interface{}) (err error) {
	// save struct
	_, err = (*conn).Do("HSET", strconv.FormatUint(uint64(cliMap.AddressId), 10), field, value)
	return err
}

// check is addressId exist
func CheckAddressId(addressId uint32, conn *gorm.DB, redisConn *redis.Conn, user *model.User, cliMap *CliMap) (inRedis bool, inDB bool, err error) {
	addressIdStr := strconv.FormatUint(uint64(addressId), 10)
	exists, err := redis.Bool((*redisConn).Do("EXISTS", addressIdStr))
	if err != nil {
		return false, false, err
	}
	if exists {
		value, err := redis.Values((*redisConn).Do("hgetall", addressIdStr))
		redis.ScanStruct(value, cliMap)
		return true, false, err
	} else {
		// check from db
		found := (*conn).Where("address_id = ?", addressId).First(&user).RecordNotFound()
		if found {
			return false, false, err
		}
		copyUserToCliMap(user, cliMap)
		return false, true, err
	}
}

func GetUserByAddress(address []byte, conn *gorm.DB, redisConn *redis.Conn, user *model.User, cliMap *CliMap) error {
	addressStr := hex.EncodeToString(address)
	exists, err := redis.Bool((*redisConn).Do("EXISTS", addressStr))
	if err != nil {
		return err
	}
	if exists {
		addrId, err := redis.Uint64((*redisConn).Do("GET", addressStr))
		if err != nil {
			return err
		}
		addressIdStr := strconv.FormatUint(uint64(addrId), 10)
		value, err := redis.Values((*redisConn).Do("hgetall", addressIdStr))
		if err != nil {
			return err
		}
		err = redis.ScanStruct(value, cliMap)
		if err != nil {
			return err
		}
	} else {
		// get from db
		found := (*conn).Where("address = ?", address).First(&user).RecordNotFound()
		if found {
			return errors.New("user not found")
		}
		copyUserToCliMap(user, cliMap)
	}
	return err
}

func PayloadToUser(user *model.User, s *ServicePayload) []byte {
	user.Address = s.Address
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

func copyUserToCliMap(user *model.User, cliMap *CliMap) {
	cliMap.Address = user.Address
	cliMap.AddressId = user.AddressId
	cliMap.ApiKey = user.ApiKey
	cliMap.Nonce = user.Nonce
	cliMap.TopicPrefix = user.TopicPrefix
	cliMap.ReplyUTXON = user.ReplyUTXON
}
