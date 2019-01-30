package tx

import (
	"encoding/hex"
	"log"
	"sync"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/gateway/mqtt"
	streamModel "github.com/FissionAndFusion/lws/internal/stream/model"
	"github.com/FissionAndFusion/lws/internal/stream/utxo"
	"github.com/FissionAndFusion/lws/test/helper"
	"github.com/jinzhu/gorm"
)

type PoolTxHandler struct {
	tx   *lws.Transaction
	dbtx *gorm.DB
}

func StartPoolTxHandler(tx *lws.Transaction) error {
	defer helper.MeasureTime(helper.MeasureTitle("handle tx pool hash[%s]", hex.EncodeToString(tx.Hash)))
	log.Printf("[DEBUG] tx pool hash[%s]", hex.EncodeToString(tx.Hash))
	defer log.Printf("[DEBUG] tx pool done hash[%s]", hex.EncodeToString(tx.Hash))
	connection := db.GetConnection()

	dbtx := connection.Begin()

	// check exsitance
	var count int
	res := dbtx.Model(&model.Tx{}).Where("hash = ?", tx.Hash).Count(&count)
	if res.Error != nil {
		log.Printf("[ERROR] error check pool tx existance failed [%s]", res.Error)
		dbtx.Rollback()
		return res.Error
	}

	if count != 0 {
		log.Printf("[DEBUG] pool tx[%s] already exists, skip", hex.EncodeToString(tx.Hash))
		dbtx.Rollback()
		return nil
	}

	updates, err := insertTx(dbtx, tx)
	if err != nil {
		log.Println("[ERROR] pool tx handler rollback")
		dbtx.Rollback()
		return err
	}

	dbtx.Commit()

	var wg sync.WaitGroup
	wg.Add(len(updates))
	for destination, item := range updates {
		// var addr []byte
		// copy(addr, destination[:])
		// go mqtt.NewUTXOUpdate(item, addr, &wg)
		var addr [33]byte
		copy(addr[:], destination[:])
		go mqtt.NewUTXOUpdate(item, addr[:], &wg)
	}
	log.Printf("[DEBUG] wait NEW UTXO Update len [%d]", len(updates))
	wg.Wait()
	log.Printf("[DEBUG] done wait NEW UTXO Update")

	return nil
}

func getSingleTxSender(dbtx *gorm.DB, inputs []byte) ([]byte, error) {
	if len(inputs) < 33 {
		return nil, nil
	}
	prevHash := inputs[:32]
	prevTx := &model.Tx{}

	res := dbtx.Select("send_to").Where("hash = ?", prevHash).Take(&prevTx)
	if res.RecordNotFound() {
		// try to query from tx pool
		prevTxPool := &model.TxPool{}
		txPoolRes := dbtx.Select("send_to").Where("hash = ?", prevHash).Take(&prevTxPool)
		if txPoolRes.Error != nil {
			log.Printf("[ERROR] error query tx sender failed [%s]", res.Error)
			return nil, res.Error
		}
		return prevTxPool.SendTo, nil
		// query from tx_pool
	} else if res.Error != nil {
		log.Printf("[ERROR] error query tx sender failed [%s]", res.Error)
		return nil, res.Error
	}
	return prevTx.SendTo, nil
}

func insertTx(dbtx *gorm.DB, tx *lws.Transaction) (map[[33]byte][]mqtt.UTXOUpdate, error) {
	ormTxPool := convertTxPoolFromDbpToOrm(tx)

	sender, err := getSingleTxSender(dbtx, ormTxPool.Inputs)
	if err != nil {
		return nil, err
	}
	ormTxPool.Sender = sender

	res := dbtx.Create(ormTxPool)
	if res.Error != nil {
		return nil, res.Error
	}

	streamTx := &streamModel.StreamTx{
		Transaction: tx,
		Sender:      sender,
	}

	txs := []*streamModel.StreamTx{streamTx}
	updates, err := utxo.HandleUtxoPool(dbtx, txs)
	if err != nil {
		log.Printf("[ERROR] txpool handle utxo failed %s", err)
		return nil, err
	}

	return updates, nil
}

func convertTxPoolFromDbpToOrm(tx *lws.Transaction) *model.TxPool {
	inputs := calculateOrmTxInputs(tx.VInput)
	sendTo := calculateOrmTxSendTo(tx.CDestination)
	return &model.TxPool{
		Hash:      tx.Hash,
		Version:   uint16(tx.NVersion),
		TxType:    uint16(tx.NType),
		LockUntil: tx.NLockUntil,
		Amount:    tx.NAmount,
		Fee:       tx.NTxFee,
		Data:      tx.VchData,
		Sig:       tx.VchSig,
		Inputs:    inputs,
		SendTo:    sendTo,
		Change:    tx.NChange,
	}
}
