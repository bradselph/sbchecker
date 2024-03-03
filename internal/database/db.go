package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sbchecker/internal/logger"
	"sbchecker/models"
)

var DB *gorm.DB

func Initialize() error {
	dsn := "u49059_BZgf1wENL1:IITg+w=Ark3+TBpNrRLiz5dN@tcp(localhost:3306)/s49059_sbchecker?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.WithError(err).Error("Failed to connect to database")
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
