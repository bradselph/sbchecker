package main

import (
	"log"
	"os"
	"os/signal"
	bot "sbchecker/cmd/dcbot"
	"sbchecker/internal/database"
	"sbchecker/internal/logger"
	"sbchecker/internal/services"
	"syscall"
)

func main() {
	log.Println("Initializing logger")
	logger.Initialize()

	logger.Log.Info("Initializing database connection")
	err := database.Initialize()
	if err != nil {
		logger.Log.WithError(err).Error("Error initializing database")
	}

	instance, err := bot.RunBot()
	if err != nil {
		logger.Log.WithError(err).Error("Error running bot")
	}

	logger.Log.Info("Bot is running")

	go services.CheckAccounts(instance)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
