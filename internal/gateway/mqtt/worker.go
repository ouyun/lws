package mqtt

import (
	"context"
	"github.com/eclipse/paho.mqtt.golang"
)

type ClientMsg struct {
	client *mqtt.Client
	msg    *mqtt.Message
}

type WorkerPool struct {
	ServiceReqChan chan *ClientMsg
	SyncReqChan    chan *ClientMsg
	UtxoAbortChan  chan *ClientMsg
	SendTxReqChan  chan *ClientMsg

	cancel context.CancelFunc
}

var workerPool *WorkerPool

func NewWorkerPool() {
	workerPool = &WorkerPool{}
	workerPool.Start()
}

func (this *WorkerPool) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	this.cancel = cancel
	this.SendTxReqChan = make(chan *ClientMsg)

	maxWorkers := 20

	for i := 1; i <= maxWorkers; i++ {
		go func(msgs chan *ClientMsg) {
			for {
				select {
				case clientMsg := <-msgs:
					sendTxReqWorkerHandler(clientMsg)
				case <-ctx.Done():
					return
				}
			}
		}(this.SendTxReqChan)
	}
}

func (this *WorkerPool) Stop() {
	this.cancel()
}
