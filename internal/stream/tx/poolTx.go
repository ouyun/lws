package tx

import (
	"encoding/hex"
	"log"

	"github.com/jinzhu/gorm"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	dbmodule "github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/internal/db/model"
)

type PoolTxHandler struct {
	tx   *lws.Transaction
	dbtx *gorm.DB
}

func StartPoolTxHandler(tx *lws.Transaction) error {
	gormdb := dbmodule.GetGormDb()

	dbtx := gormdb.Begin()

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

	dbtx.Commit()
	return nil
}

func insertTx(dbtx *gorm.DB, tx *lws.Transaction) error {
	ormTx := convertTxFromDbpToOrm(tx)

	res := dbtx.Create(ormTx)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func convertTxFromDbpToOrm(tx *lws.Transaction) *model.Tx {
	return &model.Tx{
		Hash:      tx.Hash,
		Version:   uint16(tx.NVersion),
		Type:      uint16(tx.NType),
		LockUntil: int(tx.NLockUntil),
		Amount:    tx.NAmount,
		Fee:       tx.NTxFee,
		Data:      tx.VchData,
		Sig:       tx.VchSig,
	}
}
