package helper

import (
	"log"
)

func ResetDb() {
	LoadTestSeed("schema.sql")
	log.Println("database schema reseted")
}
