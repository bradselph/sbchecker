package database

import (
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"os"
	"sbchecker/internal/logger"
	"sbchecker/models"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
)

var DB *gorm.DB

func Initialize() error {
	err := godotenv.Load()
	if err != nil {
		logger.Log.WithError(err).Error("Error loading .env file")
		return err
	}
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbVar := os.Getenv("DB_VAR")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbVar)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
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
