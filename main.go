package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	bot "sbchecker/cmd/dcbot"
	"sbchecker/internal/database"
	"sbchecker/internal/logger"
	"sbchecker/internal/services"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var instance *discordgo.Session

// The main function initializes the logger, database, and starts the bot.
// It also sets up a signal capture to handle interruptions and terminations.
func main() {
	log.Println("Bot starting...")
	logger.Initialize() // Initialize the logger
	logger.Log.Info("logger	initialized")
	err := database.Initialize() // Initialize the database
	if err != nil {
		logger.Log.WithError(err).Error("Error initializing database")
	}
	instance, err = bot.RunBot() // Start the bot
	if err != nil {
		logger.Log.WithError(err).Error("Error starting bot")
	}
	logger.Log.Info("Bot is running")
	instance.AddHandler(onGuildCreate)  // Add the onGuildCreate handler
	go services.CheckAccounts(instance) // Start the account checking service
	sc := make(chan os.Signal, 1)       // Create a signal channel
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc // Wait for a signal
}

// onGuildCreate is a handler function called when the bot joins a new server.
// It registers commands for the server and restarts the bot.
func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	guildID := event.Guild.ID
	fmt.Println("Bot joined server:", guildID)
	registerCommands(s, guildID) // Register commands for the server
	restartBot()                 // Restart the botregisterCommands(s, guildID)
}

// registerCommands registers commands for the specified server.
func registerCommands(s *discordgo.Session, guildID string) {
	fmt.Println("Registering commands for server:", guildID)
}

// restartBot closes the current Discord session, restarts the bot, and adds the onGuildCreate handler.
func restartBot() {
	if err := instance.Close(); err != nil {
		logger.Log.WithError(err).Error("Error closing Discord session")
	}

	var err error
	instance, err = bot.RunBot() // Restart the bot
	if err != nil {
		logger.Log.WithError(err).Error("Error restarting bot after	closing session")
		return
	}

	instance.AddHandler(onGuildCreate) // Add the onGuildCreate handler
	logger.Log.Info("Bot restarted successfully")
}
