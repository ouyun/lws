package helper

import (
	"log"

	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/FissionAndFusion/lws/internal/db"
)

func LoadTestSeed(filename string) {
	connection := db.GetConnection()

	_, curFile, _, _ := runtime.Caller(0)

	schemaSqlPath := filepath.Join(filepath.Dir(curFile), "..", "data", filename)

	schema, err := ioutil.ReadFile(schemaSqlPath)
	if err != nil {
		log.Fatal("load schema.sql failed", err)
	}

	_, err = connection.DB().Exec(string(schema))
	if err != nil {
		log.Fatal("run schema.sql failed", err)
	}
}
