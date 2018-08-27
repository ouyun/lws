package migration

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func All(db *gorm.DB) []*gormigrate.Migration {
	return []*gormigrate.Migration{
		M20180824113600(),
	}
}
