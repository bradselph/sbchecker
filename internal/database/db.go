package database

import (
	"github.com/silenta-salmans/sbchecker/internal/logger"
	"github.com/silenta-salmans/sbchecker/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Initialize() error {
	db, err := gorm.Open(sqlite.Open("sbchecker.db"), &gorm.Config{})
	if err != nil {
		logger.Log.WithError(err).Error("Error opening database")
		return err
	}

	DB = db

	err = DB.AutoMigrate(&models.Account{}, &models.Ban{})
	if err != nil {
		logger.Log.WithError(err).Error("Error migrating database")
		return err
	}

	return nil
}
