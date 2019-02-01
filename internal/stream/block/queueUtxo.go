package block

import (
	"context"
	"log"
	"sync"

	"github.com/FissionAndFusion/lws/internal/gateway/mqtt"
	"github.com/FissionAndFusion/lws/test/helper"
)

var queueChan chan map[[33]byte][]mqtt.UTXOUpdate

func QueueUtxoUpdates(updates map[[33]byte][]mqtt.UTXOUpdate) {
	queueChan <- updates
}

func ConsumeUtxoUpdates(ctx context.Context) {
	queueChan = make(chan map[[33]byte][]mqtt.UTXOUpdate)
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-queueChan:
			queueUtxoUpdatesToAddr(data)
		}
	}
}

func queueUtxoUpdatesToAddr(updates map[[33]byte][]mqtt.UTXOUpdate) {
	defer helper.MeasureTime(helper.MeasureTitle("block queue utxo update"))
	var wg sync.WaitGroup
	wg.Add(len(updates))
	for destination, item := range updates {
		var addr [33]byte
		copy(addr[:], destination[:])
		go mqtt.NewUTXOUpdate(item, addr[:], &wg)
	}
	log.Printf("[DEBUG] wait NEW UTXO Update len [%d]", len(updates))
	wg.Wait()
	log.Printf("[DEBUG] done wait NEW UTXO Update")
}
