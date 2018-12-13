package mqtt

import (
	"log"
	"sync"
	"time"

	"github.com/FissionAndFusion/lws/test/helper"
)

var once sync.Once

type SyncAddrChanMap struct {
	sync.RWMutex
	Map map[uint32](chan *UTXOUpdateQueueItem)
}

var syncMap *SyncAddrChanMap

func GetSyncAddrChanMap() *SyncAddrChanMap {
	once.Do(func() {
		syncMap = &SyncAddrChanMap{
			Map: make(map[uint32](chan *UTXOUpdateQueueItem)),
		}
	})
	return syncMap
}

func GetSyncAddrChan(addrId uint32) chan *UTXOUpdateQueueItem {
	defer helper.MeasureTime(helper.MeasureTitle("GetSyncAddrChan %d", addrId))
	m := GetSyncAddrChanMap()
	m.RLock()
	queueChan, ok := m.Map[addrId]
	m.RUnlock()
	if ok {
		return queueChan
	}
	return nil
}

func CloseSyncAddrChan(addrId uint32) {
	defer helper.MeasureTime(helper.MeasureTitle("CloseSyncAddrChan %d", addrId))
	m := GetSyncAddrChanMap()
	m.RLock()
	_, ok := m.Map[addrId]
	m.RUnlock()

	// chan is not existed
	if !ok {
		return
	}

	log.Printf("[DEBUG] Close try to get lock addr [%d]", addrId)
	m.Lock()
	log.Printf("[DEBUG] Close got lock addr [%d]", addrId)
	queueChan, ok := m.Map[addrId]
	if !ok {
		m.Unlock()
		log.Printf("[DEBUG] Close no ok unlock addr [%d]", addrId)
		return
	}

	delete(m.Map, addrId)
	m.Unlock()
	log.Printf("[DEBUG] Close delete ok unlock addr [%d]", addrId)

	close(queueChan)
}

func CloseAllSyncAddrChan() {
	log.Printf("[INFO] Close all sync addr chan")
	m := GetSyncAddrChanMap()
	var addrIds []uint32
	m.RLock()
	for key, _ := range m.Map {
		addrIds = append(addrIds, key)
	}
	m.RUnlock()

	for _, key := range addrIds {
		CloseSyncAddrChan(key)
	}
}

var cnt uint32

func NewSyncAddrChan(addrId uint32) {
	defer helper.MeasureTime(helper.MeasureTitle("NewSyncAddrChan %d", addrId))
	m := GetSyncAddrChanMap()
	m.RLock()
	_, ok := m.Map[addrId]
	m.RUnlock()

	// chan is existed already
	if ok {
		return
	}

	log.Printf("[DEBUG] NEW try to get lock addr [%d]", addrId)
	m.Lock()
	log.Printf("[DEBUG] NEW got lock addr [%d]", addrId)
	cnt += 1
	log.Printf("[DEBUG] NewSyncAddrChan addr [%d] cnt [%d]", addrId, cnt)
	queueChan := make(chan *UTXOUpdateQueueItem, 100)
	m.Map[addrId] = queueChan
	m.Unlock()
	log.Printf("[DEBUG] NEW unlock addr [%d]", addrId)

	go HandleSyncAddrUpdate(queueChan, addrId)
}

func HandleSyncAddrUpdate(queueChan chan *UTXOUpdateQueueItem, addrId uint32) {
	for {
		select {
		case item, ok := <-queueChan:
			if !ok {
				CloseSyncAddrChan(addrId)
				return
			}
			SendUTXOUpdate(item)
		case <-time.After(2 * time.Minute):
			CloseSyncAddrChan(addrId)
			return
		}
	}
}
