package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
	"os"
)

func Connect() (*gorm.DB, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	db, err := gorm.Open("mysql", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	db.SingularTable(true)

	return db, nil
}
