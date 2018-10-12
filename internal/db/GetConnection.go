package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"os"
)

var connection *gorm.DB

func GetConnection() *gorm.DB {
	if connection != nil {
		return connection
	}

	db, err := gorm.Open("mysql", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("database connect failed", err)
	}

	db.SingularTable(true)
	db.LogMode(true)

	connection = db

	return db
}
