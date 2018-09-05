package model

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model

	AddressId   uint32 `gorm:"primary_key;AUTO_INCREMENT"`
	Address     []byte
	ApiKey      []byte
	TopicPrefix string
	ForkNum     uint8
	ForkList    string
	ReplyUTXON  uint16
	TimeStamp   uint32
}
