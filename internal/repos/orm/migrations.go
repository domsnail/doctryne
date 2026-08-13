package orm

import (
	"github.com/domsnail/doctryne/internal/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.InspectionModel{},
		&models.DeveloperModel{},
	)
}
