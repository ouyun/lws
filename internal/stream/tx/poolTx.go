package tx

import (
	"encoding/hex"
	"log"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/gateway/mqtt"
	streamModel "github.com/FissionAndFusion/lws/internal/stream/model"
	"github.com/FissionAndFusion/lws/internal/stream/utxo"
	"github.com/jinzhu/gorm"
)

type PoolTxHandler struct {
	tx   *lws.Transaction
	dbtx *gorm.DB
}

func StartPoolTxHandler(tx *lws.Transaction) error {
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

	for destination, item := range updates {
		mqtt.NewUTXOUpdate(item, destination[:])
	}

	return nil
}

func getSingleTxSender(dbtx *gorm.DB, tx *model.Tx) ([]byte, error) {
	if len(tx.Inputs) < 33 {
		return nil, nil
	}
	prevHash := tx.Inputs[:32]
	prevTx := &model.Tx{}

	res := dbtx.Select("send_to").Where("hash = ?", prevHash).Take(&prevTx)
	if res.Error != nil {
		log.Printf("[ERROR] error query tx sender failed [%s]", res.Error)
		return nil, res.Error
	}
	return prevTx.SendTo, nil
}

func insertTx(dbtx *gorm.DB, tx *lws.Transaction) (map[[33]byte][]mqtt.UTXOUpdate, error) {
	ormTx := convertTxFromDbpToOrm(tx)

	sender, err := getSingleTxSender(dbtx, ormTx)
	if err != nil {
		return nil, err
	}
	ormTx.Sender = sender

	res := dbtx.Create(ormTx)
	if res.Error != nil {
		return nil, res.Error
	}

	streamTx := &streamModel.StreamTx{
		Transaction: tx,
		Sender:      sender,
	}
	updates, err := utxo.HandleTx(dbtx, streamTx, nil)
	if err != nil {
		return nil, err
	}

	return updates, nil
}

func convertTxFromDbpToOrm(tx *lws.Transaction) *model.Tx {
	inputs := calculateOrmTxInputs(tx.VInput)
	sendTo := calculateOrmTxSendTo(tx.CDestination)
	return &model.Tx{
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
