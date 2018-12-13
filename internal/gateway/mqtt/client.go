package mqtt

import (
	"encoding/hex"
	"errors"
	"log"
	"os"
	// "os/signal"
	"strconv"
	"time"

	"github.com/FissionAndFusion/lws/internal/coreclient"
	cclientModule "github.com/FissionAndFusion/lws/internal/coreclient/instance"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/gateway/crypto"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
)

type Service interface {
	Init()
	Start() error
	Stop() error
	Publish(string, byte, bool, []byte) error
	Subscribe(string, byte, mqtt.MessageHandler) error
}

type Program struct {
	Id     string
	Topic  string
	Client mqtt.Client
	IsLws  bool
	subs   []string
}

var programInstance *Program

var clientHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("TOPIC: %s\n", msg.Topic())
}

var msgChan = make(chan os.Signal, 1)

func StartCoreClient() *coreclient.Client {
	return cclientModule.StartCoreClient()
}

func Run(service *Program) error {
	service.Init()
	if err := service.Start(); err != nil {
		return err
	}
	programInstance = service
	return nil
}

func GetProgram() *Program {
	return programInstance
}

// start client
func (p *Program) Start() error {
	if token := p.Client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("conn mqtt broker failed : %+v \n", token.Error())
		err := errors.New("conn mqtt broker failed")
		return err
	}
	log.Printf("[INFO] mqtt client: %s started!", p.Id)
	return nil
}

// init client
func (p *Program) Init() {
	// mqtt.DEBUG = log.New(os.Stdout, "", 20)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker(os.Getenv("MQTT_URL")).SetClientID(p.Id)

	username := os.Getenv("MQTT_USERNAME")
	if username != "" {
		opts.SetUsername(username)
	}
	password := os.Getenv("MQTT_PASSWORD")
	if password != "" {
		opts.SetPassword(password)
	}

	opts.SetKeepAlive(10 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(false)
	// opts.SetDefaultPublishHandler(clientHandler)
	if p.IsLws {
		opts.SetOnConnectHandler(p.GetOnConnectHandler())
	}
	opts.SetConnectTimeout(30 * time.Second)
	opts.SetPingTimeout(3 * time.Second)
	p.Client = mqtt.NewClient(opts)
}

// stop client
func (p *Program) Stop() error {
	if p.Client.IsConnected() {
		p.Client.Disconnect(250)
		return nil
	}
	return errors.New("client did not conn broker!")
}

// publish topic
func (p *Program) Publish(topic string, qos byte, retained bool, msg []byte) error {
	token := p.Client.Publish(topic, qos, retained, msg)
	if token.Wait() && token.Error() != nil {
		log.Printf("publish get err: %s \n", token.Error())
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

func (p *Program) GetOnConnectHandler() mqtt.OnConnectHandler {
	log.Printf("[INFO] mqtt connect handler %s", p.Topic)
	var conn mqtt.OnConnectHandler = func(client mqtt.Client) {
		// log.Printf("program %s", p.Topic)
		topicSuffix := os.Getenv("MQTT_LWS_TOPIC_SUFFIX")
		client.Subscribe(p.Topic+"/lws/ServiceReq"+topicSuffix, byte(0), serviceReqHandler)
		client.Subscribe(p.Topic+"/lws/SyncReq"+topicSuffix, byte(1), syncReqHandler)
		client.Subscribe(p.Topic+"/lws/UTXOAbort"+topicSuffix, byte(1), uTXOAbortReqHandler)
		client.Subscribe(p.Topic+"/lws/SendTxReq"+topicSuffix, byte(1), sendTxReqReqHandler)
	}
	return conn
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

func GetUserByAddress(address []byte, redisConn *redis.Conn) (*CliMap, error) {
	cliMap := &CliMap{}
	addressStr := hex.EncodeToString(address)
	exists, err := redis.Bool((*redisConn).Do("EXISTS", addressStr))
	if err != nil {
		return nil, err
	}
	if exists {
		addrId, err := redis.Uint64((*redisConn).Do("GET", addressStr))
		if err != nil {
			return nil, err
		}
		addressIdStr := strconv.FormatUint(uint64(addrId), 10)
		value, err := redis.Values((*redisConn).Do("hgetall", addressIdStr))
		if err != nil {
			return nil, err
		}
		err = redis.ScanStruct(value, cliMap)
		if err != nil {
			return nil, err
		}
	}
	return cliMap, nil
}

func GetUserByAddressId(addrId uint32, redisConnArg *redis.Conn) (*CliMap, error) {
	cliMap := &CliMap{}
	var redisConn redis.Conn

	// get redisConn if not provided
	if redisConnArg == nil {
		pool := GetRedisPool()
		redisConn = pool.Get()
		defer redisConn.Close()
	} else {
		redisConn = *redisConnArg
	}

	addressIdStr := strconv.FormatUint(uint64(addrId), 10)
	value, err := redis.Values(redisConn.Do("hgetall", addressIdStr))
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}
	err = redis.ScanStruct(value, cliMap)
	if err != nil {
		return nil, err
	}
	return cliMap, nil
}

func PayloadToUser(user *model.User, s *ServicePayload) []byte {
	user.Address = s.Address
	user.ForkList = s.ForkList
	user.ForkNum = s.ForkNum
	user.TopicPrefix = s.TopicPrefix
	user.TimeStamp = s.TimeStamp
	user.ReplyUTXON = s.ReplyUTXON

	pubKey, privKey, privSignKey := crypto.GenerateKeyPair(nil)

	var address crypto.PublicKey

	copy(address[:], user.Address[1:])
	apiKey := crypto.GenerateApiKey(&privKey, &address)
	user.ApiKey = apiKey[:]
	log.Printf("[DEBUG] pubKey: %+v", hex.EncodeToString(pubKey[:]))
	log.Printf("[DEBUG] privKey: %+v", hex.EncodeToString(privKey[:]))
	log.Printf("[DEBUG] privSignKey: %+v", hex.EncodeToString(privSignKey[:]))
	log.Printf("[DEBUG] apiKey: %+v", hex.EncodeToString(apiKey[:]))
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
