package discord

import (
	"errors"
	"os"
	"sbchecker/cmd/dcbot/commands/accountage"
	"sbchecker/cmd/dcbot/commands/accountlogs"
	"sbchecker/cmd/dcbot/commands/addaccount"
	"sbchecker/cmd/dcbot/commands/removeaccount"
	"sbchecker/cmd/dcbot/commands/updateaccount"
	"sbchecker/internal/logger"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// dc is a global variable to store the Discord session
var (
	dc *discordgo.Session
)

// commandHandlers is a map that stores the function handlers for each command
var commandHandlers = map[string]func(*discordgo.Session, *discordgo.InteractionCreate){}

// Initialize function initializes the Discord bot by loading the .env file,
// setting up the Discord session, and logging in to Discord
func Initialize() error {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		logger.Log.WithError(err).Error("Error loading .env file")
		return err
	}
	// Get Discord token from environment variables
	envToken := os.Getenv("DISCORD_TOKEN")
	if envToken == "" {
		err = errors.New("DISCORD_TOKEN environment variable not set")
		logger.Log.WithError(err).WithField("env", "DISCORD_TOKEN").Error()
		return err
	}
	// Create a new Discord session
	dc, err = discordgo.New("Bot " + envToken)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Password").Error()
		return err
	}
	return nil
}

// StartBot function starts the Discord bot by opening the Discord session,
// setting the presence status, and registering commands for each guild
func StartBot() (*discordgo.Session, error) {
	// Open the Discord session
	err := dc.Open()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Login").Error()
		return nil, err
	}
	// Set the bot's presence status
	err = dc.UpdateWatchStatus(0, "Watching the Status of your Accounts so you dont have to.")
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Setting Presence Status").Error()
		return nil, err
	}
	// Get the list of guilds the bot is connected to
	guilds, err := dc.UserGuilds(100, "", "", false)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Initiating Guilds").Error()
		return nil, err
	}
	// Register commands for each guild
	for _, guild := range guilds {
		logger.Log.WithField("guild", guild.Name).Info("Connected to guild")
		addaccount.RegisterCommand(dc, guild.ID)
		commandHandlers["addaccount"] = addaccount.CommandAddAccount
		removeaccount.RegisterCommand(dc, guild.ID)
		commandHandlers["removeaccount"] = removeaccount.CommandRemoveAccount
		accountlogs.RegisterCommand(dc, guild.ID)
		commandHandlers["accountlogs"] = accountlogs.CommandAccountLogs
		updateaccount.RegisterCommand(dc, guild.ID)
		commandHandlers["updateaccount"] = updateaccount.CommandUpdateAccount
		accountage.RegisterCommand(dc, guild.ID)
		commandHandlers["accountage"] = accountage.CommandAccountAge

	}
	// Add a handler for each command
	dc.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handler, ok := commandHandlers[i.ApplicationCommandData().Name]
		if ok {
			handler(s, i)
		}
	})
	return dc, nil
}

// StopBot function stops the Discord bot by closing the Discord session
// and unregistering commands for each guild
func StopBot() error {
	// Close the Discord session
	err := dc.Close()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Shutdown", "Shutdown Process").Error()
		return err
	}
	// Get the list of guilds the bot is connected to
	guilds, err := dc.UserGuilds(100, "", "", false)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Shutdown", "Disconnecting Guilds").Error()
		return err
	}
	// Unregister commands for each guild
	for _, guild := range guilds {
		logger.Log.WithField("guild", guild.Name).Info("Disconnected from Guild")
		addaccount.UnregisterCommand(dc, guild.ID)
		removeaccount.UnregisterCommand(dc, guild.ID)
		accountlogs.UnregisterCommand(dc, guild.ID)
		updateaccount.UnregisterCommand(dc, guild.ID)
		accountage.UnregisterCommand(dc, guild.ID)

	}
	return nil
}

// restartBot function restarts the Discord bot by stopping and starting it
func restartBot() error {
	// Stop the bot
	if err := StopBot(); err != nil {
		return err
	}
	// Start the bot
	if _, err := StartBot(); err != nil {
		return err
	}
	return nil
}
