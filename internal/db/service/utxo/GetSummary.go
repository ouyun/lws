package utxo

import (
	// "github.com/FissionAndFusion/lws/internal/db/model"
	"github.com/jinzhu/gorm"
)

type sumResult struct {
	Sum   int64
	Count int
}

func GetSummary(indexList [][]byte, db *gorm.DB) (int64, int, error) {
	var r sumResult
	query := db.Raw("SELECT SUM(utxo.amount) as sum, COUNT(*) as count "+
		"FROM utxo "+
		"LEFT OUTER JOIN utxo_pool ON utxo.idx = utxo_pool.idx "+
		"WHERE utxo_pool.is_delete is NULL "+
		"AND utxo.idx in (?) ",
		indexList)

	result := query.Scan(&r)

	if result.Error != nil {
		return 0, 0, result.Error
	}

	return r.Sum, r.Count, nil
}

func GetPoolSummary(indexList [][]byte, db *gorm.DB) (int64, int, error) {
	var r sumResult
	query := db.Raw("SELECT SUM(new_utxo.amount) as sum, COUNT(*) as count "+
		"FROM utxo_pool new_utxo "+
		"LEFT OUTER JOIN utxo_pool used_utxo ON new_utxo.idx = used_utxo.idx AND used_utxo.is_delete = true "+
		"WHERE used_utxo.is_delete is NULL "+
		"AND new_utxo.idx in (?) ",
		indexList)

	result := query.Scan(&r)

	if result.Error != nil {
		return 0, 0, result.Error
	}

	return r.Sum, r.Count, nil
}
