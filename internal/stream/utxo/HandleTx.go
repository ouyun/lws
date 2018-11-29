package utxo

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/FissionAndFusion/lws/internal/constant"
	// "github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/db/service/utxo"
	"github.com/FissionAndFusion/lws/internal/db/util"
	"github.com/FissionAndFusion/lws/internal/gateway/mqtt"
	streamModel "github.com/FissionAndFusion/lws/internal/stream/model"
	sqlbuilder "github.com/huandu/go-sqlbuilder"
	"github.com/jinzhu/gorm"
)

func mapRemovedUtxosToUpdates(list []*model.Utxo) []mqtt.UTXOUpdate {
	result := make([]mqtt.UTXOUpdate, len(list))
	for i, item := range list {
		result[i] = mqtt.UTXOUpdate{
			OpType:    constant.UTXO_UPDATE_TYPE_REMOVE,
			UTXOIndex: append(item.TxHash, item.Out),
		}
	}

	return result
}

func mapUtxoWithTxToUpdate(utxo *model.Utxo, tx *streamModel.StreamTx) *mqtt.UTXO {
	return &mqtt.UTXO{
		TXID:        utxo.TxHash,
		Out:         utxo.Out,
		BlockHeight: utxo.BlockHeight,
		Type:        uint16(tx.NType),
		Amount:      utxo.Amount,
		LockUntil:   tx.NLockUntil,
		DataSize:    uint16(len(tx.VchData)),
		Data:        tx.VchData,
		Sender:      tx.Sender,
	}
}

// entry for handling utxos in a single tx
func HandleTx(db *gorm.DB, tx *streamModel.StreamTx, blockModel *model.Block) (map[[33]byte][]mqtt.UTXOUpdate, error) {
	inputLength := len(tx.VInput)
	hash := tx.Hash
	txFee := tx.NTxFee
	destination := util.MapPBDestinationToBytes(tx.CDestination)
	amount := tx.NAmount
	blockHeight := constant.BLOCK_HEIGHT_IN_POOL

	var dest [33]byte
	updates := make(map[[33]byte][]mqtt.UTXOUpdate)

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
			return nil, result.Error
		}

		// mqtt.SendUTXOUpdate(&[]mqtt.UTXOUpdate{
		// 	mqtt.UTXOUpdate{
		// 		OpType: constant.UTXO_UPDATE_TYPE_NEW,
		// 		UTXO: &mqtt.UTXO{
		// 			TXID:        hash,
		// 			Out:         0,
		// 			BlockHeight: blockHeight,
		// 			Type:        uint16(tx.NType),
		// 			Amount:      amount,
		// 			LockUntil:   tx.NLockUntil,
		// 			DataSize:    uint16(len(tx.VchData)),
		// 			Data:        tx.VchData,
		// 		},
		// 	},
		// }, destination)
		copy(dest[:], destination)
		updates[dest] = []mqtt.UTXOUpdate{
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
					Sender:      tx.Sender,
				},
			},
		}

		return updates, nil
	}
	log.Printf("start handling utxo in tx: [%s] (%v inputs)", hex.EncodeToString(tx.Hash), inputLength)

	inputs := util.MapPBTxToUtxo(tx.Transaction, blockHeight)
	// check if inputs exists and delete them
	// if not exists just throw error
	// correctness depends on external flow
	inputListInDB, err := utxo.GetListByInputs(inputs, db)
	if err != nil {
		log.Printf("[ERROR] get list by inputs failed %s", err)
		return nil, err
	}

	inputLengthInDB := len(inputListInDB)
	// inputLengthInDB == 0 indicates the tx has been handled
	if inputLengthInDB > 0 && inputLengthInDB != inputLength {
		return nil, fmt.Errorf("utxo inputs (%d) in this tx does not match to utxos (%d) in database", inputLength, inputLengthInDB)
	}

	// when has inputs in DB, means no utxo (same in opposite)
	if inputLengthInDB == 0 {
		var utxos []*model.Utxo
		if utxos, err = utxo.GetByTxHash(tx.Hash, db); err != nil {
			return nil, err
		}

		// check if the utxos are already exist (all exist or none)
		// don't care where it from (txPool or block)
		for _, item := range utxos {
			if item.BlockHeight == constant.BLOCK_HEIGHT_IN_POOL && blockModel != nil {
				item.BlockHeight = blockHeight
				result := db.Save(item)
				if result.Error != nil {
					return nil, result.Error
				}
				copy(dest[:], item.Destination)
				if updates[dest] == nil {
					updates[dest] = []mqtt.UTXOUpdate{}
				}
				updates[dest] = append(updates[dest], mqtt.UTXOUpdate{
					OpType:      constant.UTXO_UPDATE_TYPE_CHANGE,
					UTXOIndex:   append(item.TxHash, item.Out),
					BlockHeight: blockHeight,
				})
			}
		}
		return updates, nil
	}

	// remove utxos based on inputs
	if err = utxo.RemoveInputs(inputListInDB, db); err != nil {
		log.Printf("[ERROR] remove utxos base on inputs failed")
		return nil, err
	}

	// mqtt.SendUTXOUpdate(mapRemovedUtxosToUpdates(inputListInDB), inputListInDB[0].Destination)
	copy(dest[:], inputListInDB[0].Destination)
	updates[dest] = mapRemovedUtxosToUpdates(inputListInDB)

	outputs = []*model.Utxo{
		&model.Utxo{
			TxHash:      tx.Hash,
			Destination: destination,
			Amount:      tx.NAmount,
			BlockHeight: blockHeight,
			Out:         0,
		},
	}

	var inputSum int64

	for _, item := range inputListInDB {
		inputSum += item.Amount
	}
	// when txFee is not 0, it's a normal tx. otherwise it is mint tx.
	// and only if self change is larger than 0, will add additional utxo.
	change := inputSum - tx.NAmount - txFee
	if txFee != 0 && change > 0 {
		outputs = append(outputs, &model.Utxo{
			TxHash:      tx.Hash,
			Destination: inputListInDB[0].Destination, // get self change destination from last input
			Amount:      change,                       // calculate change to self
			BlockHeight: blockHeight,
			Out:         1,
		})
	}

	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("utxo")
	ib.Cols("created_at", "updated_at", "tx_hash", "destination", "amount", "block_height", "`out`")

	log.Printf("[DEBUG] utxo in tx [%s] height[%d]", hex.EncodeToString(tx.Hash), blockHeight)
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
		log.Printf("[DEBUG] utxo new [%s] out[%d] height[%d] dest[%s]", hex.EncodeToString(item.TxHash), item.Out, item.BlockHeight, hex.EncodeToString(item.Destination))
	}

	sql, args := ib.Build()
	_, err = db.CommonDB().Exec(sql, args...)
	if err != nil {
		return nil, err
	}

	for _, item := range outputs {
		// mqtt.SendUTXOUpdate(&[]mqtt.UTXOUpdate{
		// 	mqtt.UTXOUpdate{
		// 		OpType: constant.UTXO_UPDATE_TYPE_NEW,
		// 		UTXO:   mapUtxoWithTxToUpdate(item, tx),
		// 	},
		// }, item.Destination)

		copy(dest[:], item.Destination)
		if updates[dest] == nil {
			updates[dest] = []mqtt.UTXOUpdate{}
		}
		updates[dest] = append(updates[dest], mqtt.UTXOUpdate{
			OpType: constant.UTXO_UPDATE_TYPE_NEW,
			UTXO:   mapUtxoWithTxToUpdate(item, tx),
		})
	}

	return updates, nil
}
