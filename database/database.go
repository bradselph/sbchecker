package database

import (
	"codstatusbot/models"
	"errors"
	"fmt"
	"os"

	"codstatusbot/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB is the main database connection handle
var DB *gorm.DB

// Databaselogin connects to the database using the environment variables.
func Databaselogin() error {
	logger.Log.Info("Connecting to database...")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbVar := os.Getenv("DB_VAR")

	var err error

	if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" || dbVar == "" {
		err = errors.New("one or more environment variables for database not set or missing")
		logger.Log.WithError(err).WithField("Bot Startup", "database variables").Error()
		return err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbVar)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Mysql Config").Error()
		return err
	}

	DB = db

	err = DB.AutoMigrate(&models.Account{}, &models.Ban{})
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Database Models Problem").Error()
		return err
	}

	return nil
}
