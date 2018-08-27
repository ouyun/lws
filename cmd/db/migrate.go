package main

import (
	"github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/internal/db/migration"
	"gopkg.in/gormigrate.v1"
	"log"
	"os"
	"strconv"
)

func main() {
	db, err := db.Connect()
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	defer db.Close()

	migrations := migration.All(db)
	m := gormigrate.New(db, gormigrate.DefaultOptions, migrations)

	step := 1
	if len(os.Args) == 2 {
		step, err = strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatalln("wrong migration direction type provided, expected integer")
		}
	}

	switch true {
	case step > 0:
		if err := m.Migrate(); err != nil {
			log.Fatalf("migrate failed: %v", err)
		}

	// TODO: step count should match stored migrations, not only objects described.
	case step < 0:
		if length := len(migrations); -step > length {
			log.Fatalf("rollback count is out of range: %v", length)
		}
		for i := 0; i > step; i-- {
			if err := m.RollbackLast(); err != nil {
				log.Fatalf("rollback failed: %v", err)
			}
		}

	default:
		log.Fatalln("0 means nothing to do")
	}

	log.Printf("migration did run successfully")
}
