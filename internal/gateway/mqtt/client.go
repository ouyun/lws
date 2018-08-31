package mqtt

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
	// "errors"

	"github.com/eclipse/paho.mqtt.golang"
)

var (
	clientHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("TOPIC: %s\n", msg.Topic())
		// DecodePayload(msg.Payload())
	}
	lwsHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("TOPIC: %s\n", msg.Topic())
		log.Printf("收到message\n")
		if msg.Topic() == "LWS/lws/ServiceReq" {
			s := ServicePayload{}
			_, err := DecodePayload(msg.Payload(), &s)
			if err != nil {
				log.Printf("message: %+v\n", err)
			}
		}
		// if token := client.Publish("DEVICE01/fnfn/ServiceReply", 0, false, GenerateReply("ServiceReply")); token.Wait() && token.Error() != nil {
		// 	token.Wait()
		// }
		// DecodePayload(msg.Payload())
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

func (p *Program) Start() error {
	fmt.Printf("client %+v start\n", p.Id)
	if token := p.Client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if len(p.subs) > 0 {
		for i, _ := range p.subs {
			if strings.Contains(p.subs[i], "lws/ServiceReq") {
				p.Subscribe(p.subs[i], 0, lwsHandler)
			} else {
				p.Subscribe(p.subs[i], 1, lwsHandler)
			}
		}
	}
	return nil
}

func (p *Program) Init() error {
	// mqtt.DEBUG = log.New(os.Stdout, "", 0)
	// mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker("tcp://127.0.0.1:1883").SetClientID("lws")
	opts.SetKeepAlive(2 * time.Second)
	if p.isLws {
		opts.SetDefaultPublishHandler(lwsHandler)
	} else {
		opts.SetDefaultPublishHandler(clientHandler)
	}

	opts.SetPingTimeout(1 * time.Second)
	p.Client = mqtt.NewClient(opts)
	return nil
}

func (p *Program) Stop() error {
	// fmt.Println("application is end.")
	// if token := p.Client.Unsubscribe("DEVICE01/fnfn/ServiceReply"); token.Wait() && token.Error() != nil {
	// 	fmt.Println(token.Error())
	// 	os.Exit(1)
	// }
	p.Client.Disconnect(250)
	return nil
}

func (p *Program) Publish(topic string, qos byte, retained bool, msg []byte) error {
	if token := p.Client.Publish(topic, qos, retained, msg); token.Wait() && token.Error() != nil {
		token.Wait()
	}
	return nil
}

func (p *Program) Subscribe(topic string, qos byte, handler mqtt.MessageHandler) error {
	if token := p.Client.Subscribe(topic, qos, handler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}
	return nil
}
