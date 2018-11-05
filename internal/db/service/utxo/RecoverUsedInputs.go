package utxo

import (
	"log"

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
		}
	}

	// hash --> n
	inputOutMap := make(map[[32]byte]uint8)

	// get inputs txs (ignore height >= n)
	txHashList := make([][]byte, len(inputs))
	for i, input := range inputs {
		var hashArr [32]byte
		txHashList[i] = input.Hash
		copy(hashArr[:], input.Hash)
		inputOutMap[hashArr] = input.N
		// log.Printf("map hash [%v] out [%d]", input.Hash, input.N)
	}

	log.Printf("recover: inputs query done [%d]", len(inputs))

	var inputTxs []model.Tx
	results = connection.Where("hash in (?) and block_height < ?", txHashList, height).Find(&inputTxs)
	if results.Error != nil {
		return results.Error
	}

	utxos := make([]*model.Utxo, len(inputTxs))
	// genereate utxo from inputs-txs
	for i, inputTx := range inputTxs {
		var hashArr [32]byte
		copy(hashArr[:], inputTx.Hash)
		out, ok := inputOutMap[hashArr]
		if !ok {
			log.Printf("should not happen: recovery tx-hash[%s] not found", inputTx.Hash)
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
		// log.Printf("insert utxo#%d: [%v]", i, utxo)
		utxos[i] = utxo
	}

	if len(utxos) <= 0 {
		log.Printf("no utxo should be recovered")
		return nil
	}

	log.Printf("prepare bulk insert [%d]", len(utxos))
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
