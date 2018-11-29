package utxo

import (
	"encoding/hex"
	"log"

	// "github.com/FissionAndFusion/lws/internal/constant"
	model "github.com/FissionAndFusion/lws/internal/db/model"
	sqlbuilder "github.com/huandu/go-sqlbuilder"
	"github.com/jinzhu/gorm"
)

type Input struct {
	Hash []byte
	N    uint8
}

func RecoverUsedInputs(height uint32, connection *gorm.DB) error {
	var txs []model.Tx
	results := connection.Select("inputs").Where("block_height >= ?", height).Find(&txs)
	if results.Error != nil {
		return results.Error
	}

	// get inputs
	var inputs []*Input
	for _, tx := range txs {
		inputLen := len(tx.Inputs)
		for i := 0; i < inputLen; i += 33 {
			input := &Input{
				Hash: tx.Inputs[i : i+32],
				N:    tx.Inputs[i+32],
			}
			inputs = append(inputs, input)
			log.Printf("[DEBUG] input hash[%s] out[%d] tx[%s] tx-height[%d]", hex.EncodeToString(input.Hash), input.N, hex.EncodeToString(tx.Hash), tx.BlockHeight)
		}
	}

	// get inputs txs (ignore height >= n)
	txHashList := make([][]byte, len(inputs))
	for i, input := range inputs {
		txHashList[i] = input.Hash
	}

	log.Printf("[DEBUG] recover: inputs query done [%d]", len(inputs))

	var inputTxs []model.Tx
	results = connection.Where("hash in (?) and block_height < ?", txHashList, height).Find(&inputTxs)
	if results.Error != nil {
		return results.Error
	}

	// map hash -> modelTx
	txMap := make(map[[32]byte]*model.Tx)

	for idx, inputTx := range inputTxs {
		var hashArr [32]byte
		copy(hashArr[:], inputTx.Hash)
		txMap[hashArr] = &inputTxs[idx]
		// log.Printf("[DEBUG] txMap hash[%s] txHash[%s]", hex.EncodeToString(hashArr[:]), hex.EncodeToString(txMap[hashArr].Hash))
	}

	// utxos := make([]*model.Utxo, len(inputTxs))
	var utxos []*model.Utxo
	// genereate utxo from inputs-txs
	var utxoIdx int
	for _, input := range inputs {
		var hashArr [32]byte
		copy(hashArr[:], input.Hash)
		out := input.N

		inputTx, ok := txMap[hashArr]
		if !ok {
			log.Printf("[DEBUG] ignore inputs %v", input)
			continue
		}

		var dest []byte
		var amount int64
		if out == 0 {
			dest = inputTx.SendTo
			amount = inputTx.Amount
		} else {
			dest = inputTx.Sender
			amount = inputTx.Change
		}

		utxo := &model.Utxo{
			TxHash:      inputTx.Hash,
			Destination: dest,
			Amount:      amount,
			BlockHeight: inputTx.BlockHeight,
			Out:         out,
		}
		utxoIdx += 1
		log.Printf("[DEBUG] insert utxo#%d: hash[%s] out[%d] amount[%d] height[%d] dest[%s]", utxoIdx, hex.EncodeToString(utxo.TxHash), utxo.Out, utxo.Amount, utxo.BlockHeight, hex.EncodeToString(dest))
		// utxos[i] = utxo
		utxos = append(utxos, utxo)
	}

	if len(utxos) <= 0 {
		log.Printf("[DEBUG] no utxo should be recovered")
		return nil
	}

	log.Printf("[DEBUG] prepare bulk insert [%d]", len(utxos))
	// bulk create recover utxos
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("utxo")
	ib.Cols("created_at", "updated_at", "tx_hash", "destination", "amount", "block_height", "`out`")

	for _, item := range utxos {
		// log.Printf("insert utxo: [%v]", item)
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
	// log.Printf("sql: [%s] arg[%s]", sql, args)
	_, err := connection.CommonDB().Exec(sql, args...)
	if err != nil {
		return err
	}

	return nil
}
