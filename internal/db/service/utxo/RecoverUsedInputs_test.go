package utxo

import (
	"bytes"
	"encoding/hex"
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/model"
	"testing"

	"github.com/FissionAndFusion/lws/test/helper"
)

func TestRecoverUsedInputs(t *testing.T) {
	helper.ResetDb()
	helper.LoadTestSeed("seedBasic.sql")
	helper.LoadTestSeed("seedBasicPlus.sql")

	connection := db.GetConnection()

	err := RecoverUsedInputs(3, connection)
	if err != nil {
		t.Fatalf("recover used inputs error [%s]", err)
	}

	dest, _ := hex.DecodeString("020000000000000000000000000000000000000000000000000000000000000002")
	hash1, _ := hex.DecodeString("0003000000000000000000000000000000000000000000000000000000000001")

	utxo1 := &model.Utxo{}
	res := connection.Where("tx_hash = ? and `out` = ?", hash1, 0).Take(utxo1)
	if res.Error != nil {
		t.Fatalf("take utxo1 failed [%s]", res.Error)
	}

	if bytes.Compare(dest, utxo1.Destination) != 0 {
		t.Errorf("utxo1.destination expect [%v], but [%v]", hash1, utxo1.Destination)
	}
	if utxo1.Amount != 15000100 {
		t.Errorf("utxo1.destination expect 15000100, but [%d]", utxo1.Amount)
	}

	hash2, _ := hex.DecodeString("0003000000000000000000000000000000000000000000000000000000000002")
	utxo2 := &model.Utxo{}
	res = connection.Where("tx_hash = ? and `out` = ?", hash2, 1).Take(utxo2)
	if res.Error != nil {
		t.Fatalf("take utxo1 failed [%s]", res.Error)
	}

	if bytes.Compare(dest, utxo2.Destination) != 0 {
		t.Errorf("utxo1.destination expect [%v], but [%v]", hash1, utxo1.Destination)
	}
	if utxo2.Amount != 6999900 {
		t.Errorf("utxo1.destination expect 6999900, but [%d]", utxo2.Amount)
	}

}
