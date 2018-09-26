package model

import (
	// "github.com/jinzhu/gorm"
	"time"
)

type User struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`

	AddressId   uint32 `gorm:"AUTO_INCREMENT;primary_key;"`
	Address     []byte `gorm:"size:33;"`
	ApiKey      []byte `gorm:"size:32;"`
	TopicPrefix string
	ForkNum     uint8
	ForkList    []byte `gorm:"size:2048;type:blob;"`
	ReplyUTXON  uint16
	TimeStamp   uint32
	Nonce       uint16
}
