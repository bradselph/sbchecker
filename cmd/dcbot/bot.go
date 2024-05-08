package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"sbchecker/internal/discord"
	"sbchecker/internal/logger"
)

// RunBot function initializes and starts the bot.
// It returns the bot instance and any error encountered.
func RunBot() (*discordgo.Session, error) {
	// Initialize the discord session
	err := discord.Initialize()
	if err != nil {
		// Log the error and return if discord initialization fails
		logger.Log.WithError(err).Error("Error Initializing Discord")
		return nil, fmt.Errorf("error initializing discord: %w", err)
	}

	// Start the bot
	instance, err := discord.StartBot()
	if err != nil {
		// Log the error and return if starting the bot fails
		logger.Log.WithError(err).Error("Error Starting Bot")
		return nil, fmt.Errorf("error starting bot: %w", err)
	}

	// Return the bot instance
	return instance, nil
}
