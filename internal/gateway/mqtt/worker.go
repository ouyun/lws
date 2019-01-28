package mqtt

import (
	"context"
	"log"

	"github.com/eclipse/paho.mqtt.golang"
)

type Work struct {
	Client *mqtt.Client
	Msg    *mqtt.Message
	Handle mqtt.MessageHandler
}

// type WorkerPool struct {
// 	ServiceReqChan chan *ClientMsg
// 	SyncReqChan    chan *ClientMsg
// 	UtxoAbortChan  chan *ClientMsg
// 	SendTxReqChan  chan *ClientMsg

// 	cancel context.CancelFunc
// }

// var workerPool *WorkerPool

// func NewWorkerPool() {
// 	workerPool = &WorkerPool{}
// 	workerPool.Start()
// }

// func (this *WorkerPool) Start() {
// 	ctx, cancel := context.WithCancel(context.Background())

// 	this.cancel = cancel
// 	this.SendTxReqChan = make(chan *ClientMsg)

// 	maxWorkers := 20

// 	for i := 1; i <= maxWorkers; i++ {
// 		go func(msgs chan *ClientMsg) {
// 			for {
// 				select {
// 				case clientMsg := <-msgs:
// 					sendTxReqWorkerHandler(clientMsg)
// 				case <-ctx.Done():
// 					return
// 				}
// 			}
// 		}(this.SendTxReqChan)
// 	}
// }

// func (this *WorkerPool) Stop() {
// 	this.cancel()
// }

type Worker struct {
	WorkerChannel chan chan *Work // used to communicate between dispatcher and workers
	Channel       chan *Work
	Ctx           context.Context
}

// start worker
func (w *Worker) Start() {
	go func() {
		for {
			w.WorkerChannel <- w.Channel // when the worker is available place channel in queue
			select {
			case job := <-w.Channel: // worker has received job
				job.Handle(*job.Client, *job.Msg) // do work
			case <-w.Ctx.Done():
				return
			}
		}
	}()
}

var WorkerChannel = make(chan chan *Work)

type Dispatcher struct {
	WorkChan chan *Work // receives jobs to send to workers
	Cancel   context.CancelFunc
}

var dispatcher *Dispatcher

func StartDispatcher(workerCount int) *Dispatcher {
	ctx, cancel := context.WithCancel(context.Background())

	workChannel := make(chan *Work) // channel to recieve work

	dispatcher = &Dispatcher{
		WorkChan: workChannel,
		Cancel:   cancel,
	}

	for i := 0; i < workerCount; i++ {
		log.Println("starting worker: ", i)
		worker := Worker{
			Channel:       make(chan *Work),
			WorkerChannel: WorkerChannel,
			Ctx:           ctx,
		}
		worker.Start()
	}

	// start collector
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case work := <-workChannel:
				worker := <-WorkerChannel // wait for available channel
				worker <- work            // dispatch work to worker
			}
		}
	}()

	return dispatcher
}

func (this *Dispatcher) NewWorkCallback(handler mqtt.MessageHandler) mqtt.MessageHandler {
	var res mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("[DEBUG] mqtt callback msgId[%d]", msg.MessageID())
		work := &Work{
			Client: &client,
			Msg:    &msg,
			Handle: handler,
		}
		this.WorkChan <- work
	}
	return res
}
