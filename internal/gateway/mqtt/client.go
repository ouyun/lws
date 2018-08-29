package mqtt

import (
	"os"
	"log"
	"os/signal"
	"fmt"
	"time"
	// "errors"

	"github.com/eclipse/paho.mqtt.golang"
)

var (
	clientHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		// DecodePayload(msg.Payload())
	}
	lwsHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("TOPIC: %s\n", msg.Topic())
		if token := client.Publish("DEVICE01/fnfn/ServiceReply", 0, false, GenerateReply("ServiceReply")); token.Wait() && token.Error() != nil {
			token.Wait()
		}
		// DecodePayload(msg.Payload())
	}
)

type Service interface {
    Init() error
    Start() error
		Stop() error
		Publish(string, []byte) error
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

func Interrupt(){
    msgChan<-os.Interrupt
}

type Program struct {
	Id string
	Client mqtt.Client
	isLws  bool
}

func (p *Program) Start() error  {
	fmt.Printf("client %+v start\n", p.Id)
	if token := p.Client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if p.isLws {
		if token := p.Client.Subscribe("LWS/lws/ServiceReq", 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
		if token := p.Client.Subscribe("LWS/lws/SyncReq", 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
		if token := p.Client.Subscribe("LWS/lws/UTXOAbort", 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
		if token := p.Client.Subscribe("LWS/lws/SendTxReq", 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
	} else {
		if token := p.Client.Subscribe("DEVICE01/fnfn/ServiceReply", 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
		if token := p.Client.Subscribe("DEVICE01/fnfn/SyncReply", 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
		if token := p.Client.Subscribe("DEVICE01/fnfn/UTXOUpdate", 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
		if token := p.Client.Subscribe("DEVICE01/fnfn/SendTxReply", 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			os.Exit(1)
		}
	}
	return nil
}

func (p *Program) Init() error {
	mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker("tcp://127.0.0.1:1883").SetClientID("gotrivial")
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
	fmt.Println("application is end.")
	if token := p.Client.Unsubscribe("DEVICE01/fnfn/ServiceReply"); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	p.Client.Disconnect(250)
	return nil
}

func (p *Program) Publish(topic string, msg []byte) error {
	// msg, err := GeneratePayload("ServicePayload")
	// if err != nil {
	// 	return errors.New("GeneratePayload fail")
	// }
	if token :=  p.Client.Publish(topic, 0, false, msg); token.Wait() && token.Error() != nil {
		token.Wait()
	}
	return nil
}
