package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"sbchecker/internal/logger"
	"sbchecker/models"
)

var DB *gorm.DB

func Initialize() error {
	db, err := gorm.Open(sqlite.Open("internal/database/sbchecker.db"), &gorm.Config{})
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
