package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func Connect() (*gorm.DB, error) {
	if err := godotenv.Overload(); err != nil {
		log.Println("no .env file found, will try to use native environment variables")
	}

	db, err := gorm.Open("mysql", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	db.SingularTable(true)

	return db, nil
}
