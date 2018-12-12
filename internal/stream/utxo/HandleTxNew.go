package utxo

import (
	"encoding/hex"
	"log"

	"github.com/FissionAndFusion/lws/internal/constant"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/db/util"
	"github.com/FissionAndFusion/lws/internal/gateway/mqtt"
	streamModel "github.com/FissionAndFusion/lws/internal/stream/model"
	"github.com/FissionAndFusion/lws/internal/utils"
	"github.com/FissionAndFusion/lws/test/helper"
	sqlbuilder "github.com/huandu/go-sqlbuilder"
	"github.com/jinzhu/gorm"
)

func generateNewTypeUpdateItem(tx *streamModel.StreamTx, blockHeight uint32, amount int64, out uint8, dest []byte) *mqtt.UTXOUpdateWithDestination {
	return &mqtt.UTXOUpdateWithDestination{
		UTXOUpdate: &mqtt.UTXOUpdate{
			OpType: constant.UTXO_UPDATE_TYPE_NEW,
			UTXO: &mqtt.UTXO{
				TXID:        tx.Hash,
				Out:         out,
				BlockHeight: blockHeight,
				Type:        uint16(tx.NType),
				Amount:      amount,
				LockUntil:   tx.NLockUntil,
				DataSize:    uint16(len(tx.VchData)),
				Data:        tx.VchData,
				Sender:      tx.Sender,
			},
		},
		Destination: dest,
	}
}

func GetUtxoUpdateList(tx *streamModel.StreamTx, blockHeight uint32, inPool bool) ([]*mqtt.UTXOUpdateWithDestination, error) {
	defer helper.MeasureTime(helper.MeasureTitle("handle utxo.handleTx"))

	inputLength := len(tx.VInput)
	destination := util.MapPBDestinationToBytes(tx.CDestination)

	var results []*mqtt.UTXOUpdateWithDestination

	//TODO dpos fake tx hash 空块的txmint可以直接扔掉

	log.Printf("[DEBUG] start handling utxo in tx: [%s] (%v inputs)", hex.EncodeToString(tx.Hash), inputLength)

	// inPool height
	if inPool {
		// transfer amount
		if tx.NAmount > 0 {
			updateItem := &mqtt.UTXOUpdateWithDestination{
				UTXOUpdate: &mqtt.UTXOUpdate{
					OpType:      constant.UTXO_UPDATE_TYPE_CHANGE,
					UTXOIndex:   append(tx.Hash, uint8(0)),
					BlockHeight: blockHeight,
				},
				Destination: destination,
			}

			results = append(results, updateItem)
		}
		// change output
		if tx.NChange > 0 {
			updateItem := &mqtt.UTXOUpdateWithDestination{
				UTXOUpdate: &mqtt.UTXOUpdate{
					OpType:      constant.UTXO_UPDATE_TYPE_CHANGE,
					UTXOIndex:   append(tx.Hash, uint8(1)),
					BlockHeight: blockHeight,
				},
				Destination: tx.Sender,
			}
			results = append(results, updateItem)
		}
		return results, nil
	}

	// remove inputs
	if inputCnt := len(tx.VInput); inputCnt > 0 {
		for _, input := range tx.VInput {
			item := &mqtt.UTXOUpdateWithDestination{
				UTXOUpdate: &mqtt.UTXOUpdate{
					OpType:    constant.UTXO_UPDATE_TYPE_REMOVE,
					UTXOIndex: append(input.Hash, uint8(input.N)),
				},
				Destination: tx.Sender,
			}
			results = append(results, item)
		}
	}

	// add transfer target output
	if tx.NAmount > 0 {
		results = append(results, generateNewTypeUpdateItem(tx, blockHeight, tx.NAmount, 0, destination))
	}
	// add change output
	if tx.NChange > 0 {
		results = append(results, generateNewTypeUpdateItem(tx, blockHeight, tx.NChange, 1, tx.Sender))
	}

	return results, nil
}

func HandleUtxos(db *gorm.DB, txs []*streamModel.StreamTx, blockModel *model.Block, oldHashes [][]byte) (map[[33]byte][]mqtt.UTXOUpdate, error) {
	defer helper.MeasureTime(helper.MeasureTitle("HandleUtxos"))
	var updateList []*mqtt.UTXOUpdateWithDestination

	blockHeight := constant.BLOCK_HEIGHT_IN_POOL
	if blockModel != nil {
		blockHeight = blockModel.Height
	}

	for _, tx := range txs {
		curUpdateList, err := GetUtxoUpdateList(tx, blockHeight, utils.IncludeHash(tx.Hash, oldHashes))
		if err != nil {
			log.Printf("[ERROR] handle tx failed [%s], tx hash[%s]", err, hex.EncodeToString(tx.Hash))
			return nil, err
		}
		updateList = append(updateList, curUpdateList...)
	}

	// TODO handle db
	err := writeUtxoDb(db, updateList, blockModel)
	if err != nil {
		log.Printf("[ERROR] handle utxos to db err: %s", err)
		return nil, err
	}

	updatesMap := getDestinationUtxoUpdateMap(updateList)
	return updatesMap, nil
}

func getDestinationUtxoUpdateMap(updateList []*mqtt.UTXOUpdateWithDestination) map[[33]byte][]mqtt.UTXOUpdate {
	updates := make(map[[33]byte][]mqtt.UTXOUpdate)
	for _, item := range updateList {
		var dest [33]byte
		copy(dest[:], item.Destination)
		if updates[dest] == nil {
			updates[dest] = []mqtt.UTXOUpdate{}
		}
		updates[dest] = append(updates[dest], *(item.UTXOUpdate))
	}
	return updates
}

func writeUtxoDb(db *gorm.DB, updateList []*mqtt.UTXOUpdateWithDestination, blockModel *model.Block) error {
	var err error
	var removeIndexList [][]byte
	var addList []*mqtt.UTXOUpdateWithDestination
	var changeIndexList [][]byte

	for idx, item := range updateList {
		switch item.OpType {
		case constant.UTXO_UPDATE_TYPE_NEW:
			addList = append(addList, updateList[idx])
		case constant.UTXO_UPDATE_TYPE_CHANGE:
			changeIndexList = append(changeIndexList, item.UTXOIndex)
		case constant.UTXO_UPDATE_TYPE_REMOVE:
			removeIndexList = append(removeIndexList, item.UTXOIndex)
		default:
			log.Printf("[WARN] Unkonwn utxo update type [%d]", item.OpType)
		}
	}

	var actualAddList []*mqtt.UTXOUpdateWithDestination
	var actualChangeList [][]byte

	// assign actualAddList = addList - removeList
	for idx, item := range addList {
		var utxoIndex [33]byte
		copy(utxoIndex[:32], item.UTXO.TXID)
		utxoIndex[32] = item.UTXO.Out
		if !utils.IncludeHash(utxoIndex[:], removeIndexList) {
			actualAddList = append(actualAddList, addList[idx])
		}
	}

	// assign actualAddList = addList - removeList
	for idx, _ := range changeIndexList {
		if !utils.IncludeHash(changeIndexList[idx], removeIndexList) {
			actualChangeList = append(actualChangeList, changeIndexList[idx])
		}
	}

	// add new list
	err = addUtxoDb(db, actualAddList)
	if err != nil {
		log.Printf("[ERROR] add new utxo to db failed: %s", err)
		return err
	}

	err = changeUtxoDb(db, actualChangeList, blockModel)
	if err != nil {
		log.Printf("[ERROR] change utxo height to db failed: %s", err)
		return err
	}

	err = removeUtxoDb(db, removeIndexList)
	if err != nil {
		log.Printf("[ERROR] remov utxo db failed: %s", err)
		return err
	}

	return nil
}

func addUtxoDb(db *gorm.DB, addList []*mqtt.UTXOUpdateWithDestination) error {
	defer helper.MeasureTime(helper.MeasureTitle("handle utxo addUtxoDb"))
	if len(addList) == 0 {
		return nil
	}
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("utxo")
	ib.Cols("created_at", "updated_at", "tx_hash", "destination", "amount", "block_height", "`out`")

	for _, item := range addList {
		ib.Values(
			sqlbuilder.Raw("now()"),
			sqlbuilder.Raw("now()"),
			item.UTXO.TXID,
			item.Destination,
			item.UTXO.Amount,
			item.UTXO.BlockHeight,
			item.UTXO.Out,
		)
		log.Printf("[DEBUG] utxo new [%s] out[%d] height[%d] dest[%s]", hex.EncodeToString(item.UTXO.TXID), item.UTXO.Out, item.UTXO.BlockHeight, hex.EncodeToString(item.Destination))
	}
	sql, args := ib.Build()
	_, err := db.CommonDB().Exec(sql, args...)
	if err != nil {
		return err
	}
	return nil
}

func changeUtxoDb(db *gorm.DB, changeList [][]byte, blockModel *model.Block) error {
	defer helper.MeasureTime(helper.MeasureTitle("handle utxo changeUtxoDb"))
	if len(changeList) == 0 {
		return nil
	}
	query := db.Model(&model.Utxo{})

	for idx, _ := range changeList {
		hash := changeList[idx][:32]
		out := uint8(changeList[idx][32])
		query = query.Or(map[string]interface{}{
			"tx_hash": hash,
			"out":     out,
		})
	}

	result := query.Updates(map[string]interface{}{
		"block_height": blockModel.Height,
		"block_id":     blockModel.ID,
		"block_hash":   blockModel.Hash,
	})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func removeUtxoDb(db *gorm.DB, removeList [][]byte) error {
	defer helper.MeasureTime(helper.MeasureTitle("handle utxo removeUtxoDb"))
	if len(removeList) == 0 {
		return nil
	}

	query := db.Model(&model.Utxo{})

	for idx, _ := range removeList {
		hash := removeList[idx][:32]
		out := uint8(removeList[idx][32])
		query = query.Or(map[string]interface{}{
			"tx_hash": hash,
			"out":     out,
		})
		log.Printf("[DEBUG] remove utxo hash [%s] out[%d]", hex.EncodeToString(hash), out)
	}

	result := query.Unscoped().Delete(model.Utxo{})

	if result.Error != nil {
		return result.Error
	}

	return nil
}