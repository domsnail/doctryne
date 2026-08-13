package orm

import (
	"github.com/domsnail/doctryne/cfg"
	"gorm.io/gorm"
)

type DatabaseRepoImpl struct {
	db gorm.DB

	config cfg.DatabaseConfig
}

func NewDatabaseRepoImpl(db gorm.DB, config cfg.DatabaseConfig) *DatabaseRepoImpl {
	return &DatabaseRepoImpl{db: db, config: config}
}
