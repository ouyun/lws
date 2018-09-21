package utxo

import (
	"fmt"
	"log"

	sqlbuilder "github.com/huandu/go-sqlbuilder"
	"github.com/jinzhu/gorm"
	"github.com/lomocoin/lws/internal/constant"
	"github.com/lomocoin/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/lomocoin/lws/internal/db/model"
	"github.com/lomocoin/lws/internal/db/service/utxo"
	"github.com/lomocoin/lws/internal/db/util"
)

// entry for handling utxos in a single tx
func HandleTx(db *gorm.DB, tx *lws.Transaction, blockModel *model.Block) error {
	inputLength := len(tx.VInput)
	hash := tx.Hash
	txFee := tx.NTxFee
	destination := util.MapPBDestinationToBytes(tx.CDestination)
	amount := tx.NAmount
	blockHeight := constant.BLOCK_HEIGHT_IN_POOL
	if blockModel != nil {
		blockHeight = blockModel.Height
	}

	var outputs []*model.Utxo

	if inputLength == 0 {
		log.Printf("add utxo in mint tx")
		var mintUtxo model.Utxo
		result := db.FirstOrCreate(&mintUtxo, model.Utxo{
			TxHash:      hash,
			Destination: destination,
			Amount:      amount,
			BlockHeight: blockHeight,
			Out:         0,
		})
		return result.Error
	}
	log.Printf("start handling utxo in tx: %v (%v inputs)", tx.Hash, inputLength)

	inputs := util.MapPBTxToUtxo(tx, blockHeight)
	// check if inputs exists and delete them
	// if not exists just throw error
	// correctness depends on external flow
	inputList, err := utxo.GetListByInputs(inputs, db)
	if err != nil {
		return err
	}

	if len(inputList) != inputLength {
		return fmt.Errorf("utxo inputs (%d) in this tx does not match to utxos (%d) in database", inputLength, len(inputList))
	}

	var inputSum int64

	for _, item := range inputList {
		inputSum += item.Amount
	}

	// remove utxos based on inputs
	if err = utxo.RemoveInputs(inputList, db); err != nil {
		return err
	}

	var utxos []*model.Utxo
	if utxos, err = utxo.GetByTxHash(tx.Hash, db); err != nil {
		return err
	}

	// check if the utxos are already exist (all exist or none)
	// don't care where it from (txPool or block)
	if len(utxos) != 0 {
		for _, item := range utxos {
			if item.BlockHeight == constant.BLOCK_HEIGHT_IN_POOL && blockModel != nil {
				item.BlockHeight = blockModel.Height
				result := db.Save(item)
				if result.Error != nil {
					return result.Error
				}
			}
		}
		return nil
	}

	outputs = []*model.Utxo{
		&model.Utxo{
			TxHash:      tx.Hash,
			Destination: destination,
			Amount:      tx.NAmount,
			BlockHeight: blockHeight,
			Out:         0,
		},
	}
	// when txFee is not 0, it's a normal tx. otherwise it is mint tx.
	if txFee != 0 {
		outputs = append(outputs, &model.Utxo{
			TxHash:      tx.Hash,
			Destination: inputList[0].Destination,      // get self change destination from last input
			Amount:      inputSum - tx.NAmount - txFee, // calculate change to self
			BlockHeight: blockHeight,
			Out:         1,
		})
	}

	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("utxo")
	ib.Cols("created_at", "updated_at", "tx_hash", "destination", "amount", "block_height", "`out`")

	for _, item := range outputs {
		ib.Values(
			sqlbuilder.Raw("now()"),
			sqlbuilder.Raw("now()"),
			item.TxHash,
			item.Destination,
			item.Amount,
			item.BlockHeight,
			item.Out,
		)
	}

	sql, args := ib.Build()
	_, err = db.CommonDB().Exec(sql, args...)
	return err
}
