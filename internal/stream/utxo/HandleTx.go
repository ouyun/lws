package utxo

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/FissionAndFusion/lws/internal/constant"
	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/db/service/utxo"
	"github.com/FissionAndFusion/lws/internal/db/util"
	"github.com/FissionAndFusion/lws/internal/gateway/mqtt"
	sqlbuilder "github.com/huandu/go-sqlbuilder"
	"github.com/jinzhu/gorm"
)

func mapRemovedUtxosToUpdates(list []*model.Utxo) *[]mqtt.UTXOUpdate {
	result := make([]mqtt.UTXOUpdate, len(list))
	for i, item := range list {
		result[i] = mqtt.UTXOUpdate{
			OpType:    constant.UTXO_UPDATE_TYPE_REMOVE,
			UTXOIndex: append(item.TxHash, item.Out),
		}
	}

	return &result
}

func mapUtxoWithTxToUpdate(utxo *model.Utxo, tx *lws.Transaction) *mqtt.UTXO {
	return &mqtt.UTXO{
		TXID:        utxo.TxHash,
		Out:         utxo.Out,
		BlockHeight: utxo.BlockHeight,
		Type:        uint16(tx.NType),
		Amount:      utxo.Amount,
		LockUntil:   tx.NLockUntil,
		DataSize:    uint16(len(tx.VchData)),
		Data:        tx.VchData,
	}
}

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

		if result.Error != nil {
			return result.Error
		}

		mqtt.SendUTXOUpdate(&[]mqtt.UTXOUpdate{
			mqtt.UTXOUpdate{
				OpType: constant.UTXO_UPDATE_TYPE_NEW,
				UTXO: &mqtt.UTXO{
					TXID:        hash,
					Out:         0,
					BlockHeight: blockHeight,
					Type:        uint16(tx.NType),
					Amount:      amount,
					LockUntil:   tx.NLockUntil,
					DataSize:    uint16(len(tx.VchData)),
					Data:        tx.VchData,
				},
			},
		}, destination)

		return nil
	}
	log.Printf("start handling utxo in tx: [%s] (%v inputs)", hex.EncodeToString(tx.Hash), inputLength)

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

	mqtt.SendUTXOUpdate(mapRemovedUtxosToUpdates(inputList), inputList[0].Destination)

	var utxos []*model.Utxo
	if utxos, err = utxo.GetByTxHash(tx.Hash, db); err != nil {
		return err
	}

	// check if the utxos are already exist (all exist or none)
	// don't care where it from (txPool or block)
	if len(utxos) != 0 {
		for _, item := range utxos {
			if item.BlockHeight == constant.BLOCK_HEIGHT_IN_POOL && blockModel != nil {
				item.BlockHeight = blockHeight
				result := db.Save(item)
				if result.Error != nil {
					return result.Error
				}
				mqtt.SendUTXOUpdate(&[]mqtt.UTXOUpdate{
					mqtt.UTXOUpdate{
						OpType:      constant.UTXO_UPDATE_TYPE_CHANGE,
						UTXOIndex:   append(item.TxHash, item.Out),
						BlockHeight: blockHeight,
					},
				}, item.Destination)
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
	// and only if self change is larger than 0, will add additional utxo.
	if txFee != 0 && inputSum-tx.NAmount-txFee > 0 {
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
	if err != nil {
		return err
	}

	for _, item := range outputs {
		mqtt.SendUTXOUpdate(&[]mqtt.UTXOUpdate{
			mqtt.UTXOUpdate{
				OpType: constant.UTXO_UPDATE_TYPE_NEW,
				UTXO:   mapUtxoWithTxToUpdate(item, tx),
			},
		}, item.Destination)
	}

	return nil
}
