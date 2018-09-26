package main

import (
	"github.com/FissionAndFusion/lws/internal/db"
	"github.com/FissionAndFusion/lws/internal/db/migration"
	"gopkg.in/gormigrate.v1"
	"log"
	"os"
	"strconv"
)

func main() {
	var err error

	connection := db.GetConnection()

	defer connection.Close()

	migrations := migration.All(connection)
	m := gormigrate.New(connection, gormigrate.DefaultOptions, migrations)

	step := 1
	if len(os.Args) == 2 {
		step, err = strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatalln("wrong migration direction type provided, expected integer")
		}
	}

	switch {
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
