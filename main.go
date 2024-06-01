package main

import (
	"codstatusbot2.0/bot"
	"codstatusbot2.0/database"
	"codstatusbot2.0/logger"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	logger.Log.Info("Bot starting...")
	err := loadEnvironmentVariables()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Environment Variables").Error()
		os.Exit(1)
	}

	err = database.Databaselogin()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Database login").Error()
		os.Exit(1)
	}
	err = bot.StartBot()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Discord login").Error()
		os.Exit(1)
	}
	logger.Log.Info("Bot is running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}

func loadEnvironmentVariables() error {
	logger.Log.Info("Loading environment variables...")
	err := godotenv.Load()
	if err != nil {
		logger.Log.WithError(err).Error("Error loading .env file")
		return err
	}
	return nil
}
