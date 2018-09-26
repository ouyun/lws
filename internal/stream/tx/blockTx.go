package tx

import (
	// "encoding/hex"
	"bytes"
	"fmt"
	"log"

	"github.com/FissionAndFusion/lws/internal/coreclient/DBPMsg/go/lws"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/FissionAndFusion/lws/internal/stream/utxo"
	sqlbuilder "github.com/huandu/go-sqlbuilder"
	"github.com/jinzhu/gorm"
)

type BlockTxHandler struct {
	txs        []*lws.Transaction
	dbtx       *gorm.DB
	blockModel *model.Block
	// newTxs   []*lws.Transaction
	// oldTxs   []*lws.Transaction
}

func StartBlockTxHandler(db *gorm.DB, txs []*lws.Transaction, blockModel *model.Block) error {
	log.Printf("StartBlockTxHandler len txs [%d]", len(txs))
	h := &BlockTxHandler{
		txs:        txs,
		blockModel: blockModel,
		dbtx:       db,
	}

	err := h.handleTxs()
	// h.rollbackIfErr(err)
	return err
}

func (h *BlockTxHandler) handleTxs() error {
	// query existance
	oldHashes, err := h.queryExistance()
	if err != nil {
		log.Printf("queryExistance for [%s]", err)
		return err
	}

	newTxs, oldTxs := h.getOldNewTxList(oldHashes)
	err = h.deleteTxs(oldHashes)
	if err != nil {
		log.Printf("delete hashex failed for [%s]", err)
		return err
	}

	err = h.insertTxs(oldTxs, h.blockModel)
	if err != nil {
		log.Printf("insert old hashex failed for [%s]", err)
		return err
	}

	err = h.insertTxs(newTxs, h.blockModel)
	if err != nil {
		log.Printf("insert new hashex failed for [%s]", err)
		return err
	}

	return nil
}

func (h *BlockTxHandler) rollbackIfErr(err error) {
	if err != nil {
		log.Printf("roll back for [%s]", err)
		h.dbtx.Rollback()
	}
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

func (h *BlockTxHandler) insertTxs(txs []*lws.Transaction, block *model.Block) error {
	if len(txs) == 0 {
		return nil
	}
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("tx")
	// missing id
	ib.Cols("created_at", "updated_at", "hash", "version", "tx_type",
		"block_id", "block_hash", "block_height",
		"inputs", "send_to",
		"lock_until", "amount", "fee", "data", "sig")

	for _, tx := range txs {
		insertBuilderTxValue(ib, tx, block)
		if err := utxo.HandleTx(h.dbtx, tx, block); err != nil {
			return err
		}
	}

	sql, args := ib.Build()
	log.Println(sql, args)
	results, err := h.dbtx.CommonDB().Exec(sql, args...)
	if err != nil {
		log.Printf("bulk tx insertion failed: [%s]", err)
		return err
	}
	if cnt, err := results.RowsAffected(); int(cnt) != len(txs) {
		if err != nil {
			log.Printf("can not get inserted cnt error [%s]", err)
			return err
		}
		err = fmt.Errorf("try to insert [%d] tx, but [%d] success", len(txs), cnt)
		log.Printf("insert cnt error [%s]", err)
		return err
	}
	return nil
}

func insertBuilderTxValue(ib *sqlbuilder.InsertBuilder, tx *lws.Transaction, block *model.Block) {
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
		tx.VchSig)
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
