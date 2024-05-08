package database

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sbchecker/internal/logger"
	"sbchecker/models"
)

// DB is the global database connection handle.
var DB *gorm.DB

// Initialize sets up the database connection.
func Initialize() error {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", ".Env File Problem").Error()
		return err
	}

	// Get database connection details from environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbVar := os.Getenv("DB_VAR")

	// Check if all necessary environment variables are set
	if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" || dbVar == "" {
		err = errors.New("one or more environment variables for database not set or missing")
		logger.Log.WithError(err).WithField("Bot Startup", "database variables").Error()
		return err
	}

	// Create the DSN for the database connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbVar)

	// Open a new database connection
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Mysql Config").Error()
		return err
	}

	// Assign the database connection handle to the global variable
	DB = db

	// Auto migrate the models
	err = DB.AutoMigrate(&models.Account{}, &models.Ban{})
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Database Models Problem").Error()
		return err
	}

	// Return nil if no error occurred
	return nil
}
