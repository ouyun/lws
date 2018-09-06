package testdb

import (
	"database/sql"
	"log"
	"os"

	"github.com/lomocoin/lws/internal/db"
	"github.com/lomocoin/lws/internal/db/migration"
	"gopkg.in/gormigrate.v1"
)

func ResetDb() {
	dbIns, err := sql.Open("mysql", os.Getenv("TEST_ROOT_DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	defer dbIns.Close()

	DATABASE_NAME := os.Getenv("TEST_DATABASE_NAME")
	_, err = dbIns.Exec("drop DATABASE " + DATABASE_NAME)
	if err != nil {
		panic(err)
	}
	_, err = dbIns.Exec("Create DATABASE " + DATABASE_NAME)
	if err != nil {
		panic(err)
	}

	connection, err := db.Connect()
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	defer connection.Close()

	connection.Exec("drop database test;create database test;")
	log.Println("drop and create done", os.Getenv("DATABASE_URL"))

	migrations := migration.All(connection)
	m := gormigrate.New(connection, gormigrate.DefaultOptions, migrations)

	if err := m.Migrate(); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}
}
