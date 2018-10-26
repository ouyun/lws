package tx

import (
	"bytes"
	"fmt"
	"log"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/gateway/mqtt"
	streamModel "github.com/FissionAndFusion/lws/internal/stream/model"
	"github.com/FissionAndFusion/lws/internal/stream/utxo"
	sqlbuilder "github.com/huandu/go-sqlbuilder"
	"github.com/jinzhu/gorm"
)

type BlockTxHandler struct {
	txs           []*lws.Transaction
	dbtx          *gorm.DB
	blockModel    *model.Block
	mapTxToSender map[[32]byte][]byte
	// newTxs   []*lws.Transaction
	// oldTxs   []*lws.Transaction
}

func StartBlockTxHandler(db *gorm.DB, txs []*lws.Transaction, blockModel *model.Block) (map[[33]byte][]mqtt.UTXOUpdate, error) {
	log.Printf("StartBlockTxHandler len txs [%d]", len(txs))

	var updates map[[33]byte][]mqtt.UTXOUpdate

	h := &BlockTxHandler{
		txs:        txs,
		blockModel: blockModel,
		dbtx:       db,
	}

	// prepare txs
	err := h.prepareSenders()
	if err != nil {
		log.Printf("prepare senders failed [%s]", err)
		return nil, err
	}

	// query existance
	oldHashes, err := h.queryExistance()
	if err != nil {
		log.Printf("queryExistance for [%s]", err)
		return nil, err
	}

	newTxs, oldTxs := h.getOldNewTxList(oldHashes)
	err = h.deleteTxs(oldHashes)
	if err != nil {
		log.Printf("delete hashex failed for [%s]", err)
		return nil, err
	}

	pendingTxs := []*lws.Transaction{}
	pendingTxs = append(pendingTxs, oldTxs...)
	pendingTxs = append(pendingTxs, newTxs...)

	updates, err = h.insertTxs(pendingTxs, h.blockModel)
	if err != nil {
		log.Printf("insert old/new hashex failed for [%s]", err)
		return nil, err
	}

	return updates, nil
}

func (h *BlockTxHandler) rollbackIfErr(err error) {
	if err != nil {
		log.Printf("roll back for [%s]", err)
		h.dbtx.Rollback()
	}
}

func (h *BlockTxHandler) prepareSenders() error {
	if len(h.txs) == 0 {
		return nil
	}
	var prevTxs []interface{}
	mapPrevTxToTx := make(map[[32]byte][]byte)
	mapTxToSender := make(map[[32]byte][]byte)
	// cache current txs sendto, incase tx3's input is tx2.
	mapTxToSendto := make(map[[32]byte][]byte)

	for _, tx := range h.txs {
		var hash [32]byte
		copy(hash[:], tx.Hash)
		mapTxToSendto[hash] = calculateOrmTxSendTo(tx.CDestination)
	}

	for _, tx := range h.txs {
		if len(tx.VInput) > 0 {
			var prevTx [32]byte
			copy(prevTx[:], tx.VInput[0].Hash)
			// if prev tx in current txs
			if sendto, ok := mapTxToSendto[prevTx]; ok {
				var hash [32]byte
				copy(hash[:], tx.Hash)
				mapTxToSender[hash] = sendto
			} else {
				prevTxs = append(prevTxs, prevTx[:])
				mapPrevTxToTx[prevTx] = tx.Hash
			}
		}
	}

	if len(prevTxs) == 0 {
		return nil
	}

	// build sender query sql
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("hash", "send_to")
	sb.From("tx")
	sb.Where(sb.In("hash", prevTxs...))
	sql, args := sb.Build()

	// log.Printf("prepare sender sql[%s] args[%v]", sql, args)
	// get results
	rows, err := h.dbtx.CommonDB().Query(sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	var (
		prevHash []byte
		sender   []byte
	)
	for rows.Next() {
		err := rows.Scan(&prevHash, &sender)
		if err != nil {
			return err
		}
		// log.Printf("prevHash[%v], sender[%v]", prevHash, sender)

		// current tx --> inputs(prevHash) --> sendto
		var prevHashArr [32]byte
		copy(prevHashArr[:], prevHash)

		if hash, ok := mapPrevTxToTx[prevHashArr]; ok {
			var hashArr [32]byte
			copy(hashArr[:], hash)
			mapTxToSender[hashArr] = sender
		}
	}
	// log.Printf("mapTxToSender %v", mapTxToSender)
	h.mapTxToSender = mapTxToSender
	return nil
}

// for testing
func (h *BlockTxHandler) GetMapTxToSender() map[[32]byte][]byte {
	return h.mapTxToSender
}

func (h *BlockTxHandler) getOldNewTxList(oldHashes [][]byte) ([]*lws.Transaction, []*lws.Transaction) {
	newTxs := make([]*lws.Transaction, 0)
	oldTxs := make([]*lws.Transaction, 0)
	for _, tx := range h.txs {
		hash := tx.Hash
		if included := includeHash(hash, oldHashes); !included {
			newTxs = append(newTxs, tx)
		} else {
			oldTxs = append(oldTxs, tx)
		}
	}
	return newTxs, oldTxs
}

func (h *BlockTxHandler) queryExistance() ([][]byte, error) {
	txids := make([][]byte, len(h.txs))
	for idx, tx := range h.txs {
		txids[idx] = tx.Hash
	}
	return h.queryExistanceTxids(txids)
}

func (h *BlockTxHandler) queryExistanceTxids(txids [][]byte) ([][]byte, error) {
	var newHashes [][]byte
	results := h.dbtx.Model(&model.Tx{}).
		Where("hash in (?)", txids).
		Pluck("Hash", &newHashes)

	if results.Error != nil {
		log.Printf("query existance err[%s]", results.Error)
		// h.dbtx.Rollback()
		return nil, results.Error
	}

	log.Printf("newHashes = %+v\n", newHashes)

	return newHashes, nil
}

func (h *BlockTxHandler) deleteTxs(hashes [][]byte) error {
	if len(hashes) == 0 {
		return nil
	}
	// use hard delete
	result := h.dbtx.Unscoped().Where("hash in (?)", hashes).Delete(&model.Tx{})
	if result.Error != nil {
		log.Printf("delete txs failed [%s]", result.Error)
		return result.Error
	}
	return nil
}

func (h *BlockTxHandler) insertTxs(txs []*lws.Transaction, block *model.Block) (map[[33]byte][]mqtt.UTXOUpdate, error) {
	updates := make(map[[33]byte][]mqtt.UTXOUpdate)
	if len(txs) == 0 {
		return updates, nil
	}
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("tx")
	// missing id
	ib.Cols("created_at", "updated_at", "hash", "version", "tx_type",
		"block_id", "block_hash", "block_height",
		"inputs", "send_to",
		"lock_until", "amount", "fee", "data", "sig", "sender")

	for _, tx := range txs {
		streamTx := mapLwsTxToStreamTx(tx, h.mapTxToSender)
		insertBuilderTxValue(ib, streamTx, block)
		txUpdates, err := utxo.HandleTx(h.dbtx, streamTx, block)
		if err != nil {
			return nil, err
		}
		for destination, items := range txUpdates {
			if updates[destination] == nil {
				updates[destination] = []mqtt.UTXOUpdate{}
			}
			updates[destination] = append(updates[destination], items...)
		}
	}

	sql, args := ib.Build()
	log.Println(sql, args)
	results, err := h.dbtx.CommonDB().Exec(sql, args...)
	if err != nil {
		log.Printf("bulk tx insertion failed: [%s]", err)
		return nil, err
	}
	if cnt, err := results.RowsAffected(); int(cnt) != len(txs) {
		if err != nil {
			log.Printf("can not get inserted cnt error [%s]", err)
			return nil, err
		}
		err = fmt.Errorf("try to insert [%d] tx, but [%d] success", len(txs), cnt)
		log.Printf("insert cnt error [%s]", err)
		return nil, err
	}
	return updates, nil
}

func mapLwsTxToStreamTx(lwsTx *lws.Transaction, mapTxToSender map[[32]byte][]byte) *streamModel.StreamTx {
	sender := getSenderFromMap(lwsTx.Hash, mapTxToSender)
	return &streamModel.StreamTx{
		Transaction: lwsTx,
		Sender:      sender,
	}
}

func insertBuilderTxValue(ib *sqlbuilder.InsertBuilder, tx *streamModel.StreamTx, block *model.Block) {
	inputs := calculateOrmTxInputs(tx.VInput)
	sendTo := calculateOrmTxSendTo(tx.CDestination)

	ib.Values(
		sqlbuilder.Raw("now()"), //created_at
		sqlbuilder.Raw("now()"), //updated_at
		tx.Hash,
		uint16(tx.NVersion),
		uint16(tx.NType),
		block.ID,
		block.Hash,
		block.Height,
		inputs,
		sendTo,
		tx.NLockUntil,
		tx.NAmount,
		tx.NTxFee,
		tx.VchData,
		tx.VchSig,
		tx.Sender)
}

func getSenderFromMap(txHash []byte, mapTxToSender map[[32]byte][]byte) []byte {
	var hashArr [32]byte
	copy(hashArr[:], txHash)
	sender, ok := mapTxToSender[hashArr]
	if !ok {
		log.Printf("sender not found hash[%v]", hashArr)
		return nil
	}
	log.Printf("got sender [%v]", sender)
	return sender
}

func includeHash(hash []byte, hashList [][]byte) bool {
	for _, target := range hashList {
		if bytes.Compare(hash, target) == 0 {
			return true
		}
	}
	return false
}

func calculateOrmTxSendTo(dest *lws.Transaction_CDestination) []byte {
	sendToBuf := bytes.NewBuffer(make([]byte, 0))
	sendToBuf.Grow(33)
	sendToBuf.WriteByte(byte(uint8(dest.Prefix)))
	sendToBuf.Write(dest.Data)
	return sendToBuf.Bytes()
}

func calculateOrmTxInputs(vInput []*lws.Transaction_CTxIn) []byte {
	inputNum := len(vInput)
	inputBuf := bytes.NewBuffer(make([]byte, 0))
	inputBuf.Grow(33 * inputNum)
	for _, input := range vInput {
		inputBuf.Write(input.Hash)
		inputBuf.WriteByte(byte(uint8(input.N)))
	}
	return inputBuf.Bytes()
}
