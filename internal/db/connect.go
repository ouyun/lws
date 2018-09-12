package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"os"
)

var connection *gorm.DB

func connect() (db *gorm.DB, err error) {
	db, err = gorm.Open("mysql", os.Getenv("DATABASE_URL"))
	if err != nil {
		return
	}

	db.SingularTable(true)

	return
}

func GetConnection() (db *gorm.DB, err error) {
	if connection != nil {
		return connection, nil
	}

	db, err = gorm.Open("mysql", os.Getenv("DATABASE_URL"))
	if err != nil {
		return
	}

	db.SingularTable(true)
	// db.LogMode(true)

	connection = db

	return
}
