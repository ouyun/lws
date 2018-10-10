package tx

import (
	"encoding/hex"
	"log"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
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
		log.Printf("error check pool tx existance failed [%s]", res.Error)
		dbtx.Rollback()
		return res.Error
	}

	if count != 0 {
		log.Printf("pool tx[%s] already exists, skip", hex.EncodeToString(tx.Hash))
		dbtx.Rollback()
		return nil
	}

	err := insertTx(dbtx, tx)
	if err != nil {
		log.Println("pool tx handler rollback")
		dbtx.Rollback()
		return err
	}

	err = utxo.HandleTx(dbtx, tx, nil)
	if err != nil {
		log.Printf("handle utxo error: %v", err)
		dbtx.Rollback()
		return err
	}

	dbtx.Commit()
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
		log.Printf("error query tx sender failed [%s]", res.Error)
		return nil, res.Error
	}
	return prevTx.SendTo, nil
}

func insertTx(dbtx *gorm.DB, tx *lws.Transaction) error {
	ormTx := convertTxFromDbpToOrm(tx)

	sender, err := getSingleTxSender(dbtx, ormTx)
	if err != nil {
		return err
	}
	ormTx.Sender = sender

	res := dbtx.Create(ormTx)
	if res.Error != nil {
		return res.Error
	}

	// if err := utxo.HandleTx(dbtx, tx); err != nil {
	// 	return err
	// }

	return nil
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
	}
}
