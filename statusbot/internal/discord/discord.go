package discord

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"os"
	"sbchecker/cmd/dcbot/commands/accountage"
	"sbchecker/cmd/dcbot/commands/accountlogs"
	"sbchecker/cmd/dcbot/commands/addaccount"
	"sbchecker/cmd/dcbot/commands/removeaccount"
	"sbchecker/cmd/dcbot/commands/updateaccount"
	"sbchecker/internal/logger"
)

var (
	dc *discordgo.Session
)

var commandHandlers = map[string]func(*discordgo.Session, *discordgo.InteractionCreate){}

func Initialize() error {
	err := godotenv.Load()
	if err != nil {
		logger.Log.WithError(err).Error("Error loading .env file")
		return err
	}
	envToken := os.Getenv("DISCORD_TOKEN")
	if envToken == "" {
		err = errors.New("DISCORD_TOKEN environment variable not set")
		logger.Log.WithError(err).WithField("env", "DISCORD_TOKEN").Error()
		return err
	}
	dc, err = discordgo.New("Bot " + envToken)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Password").Error()
		return err
	}
	return nil
}

func StartBot() (*discordgo.Session, error) {
	err := dc.Open()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Login").Error()
		return nil, err
	}
	guilds, err := dc.UserGuilds(100, "", "", false)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Initiating Guilds").Error()
		return nil, err
	}
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
	dc.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handler, ok := commandHandlers[i.ApplicationCommandData().Name]
		if ok {
			handler(s, i)
		}
	})
	return dc, nil
}
func stopBot() error {
	err := dc.Close()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Shutdown", "Shutdown Process").Error()
		return err
	}
	guilds, err := dc.UserGuilds(100, "", "", false)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Shutdown", "Disconnecting Guilds").Error()
		return err
	}
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
func restartBot() error {
	if err := stopBot(); err != nil {
		return err
	}
	if _, err := StartBot(); err != nil {
		return err
	}
	return nil
}
