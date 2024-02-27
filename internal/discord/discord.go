package discord

import (
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/silenta-salmans/sbchecker/cmd/dcbot/commands/accountlogs"
	"github.com/silenta-salmans/sbchecker/cmd/dcbot/commands/addaccount"
	"github.com/silenta-salmans/sbchecker/cmd/dcbot/commands/removeaccount"
	"github.com/silenta-salmans/sbchecker/internal/logger"
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
		logger.Log.WithField("env", "DISCORD_TOKEN").Error("Environment variable not set")

		return nil
	}

	dc, err = discordgo.New("Bot " + envToken)
	if err != nil {
		logger.Log.WithError(err).Error("Error creating new Discord session")
		return err
	}

	return nil
}

func StartBot() (*discordgo.Session, error) {
	err := dc.Open()
	if err != nil {
		logger.Log.WithError(err).Error("Error opening connection to Discord")
		return nil, err
	}

	guilds, err := dc.UserGuilds(100, "", "")
	if err != nil {
		logger.Log.WithError(err).Error("Error getting guilds")
		return nil, err
	}

	for _, guild := range guilds {
		logger.Log.WithField("guild", guild.Name).Info("Connected to guild")

		addaccount.RegisterCommand(dc, guild.ID)
		commandHandlers["addaccount"] = addaccount.AddAccountCommand

		removeaccount.RegisterCommand(dc, guild.ID)
		commandHandlers["removeaccount"] = removeaccount.RemoveAccountCommand

		accountlogs.RegisterCommand(dc, guild.ID)
		commandHandlers["accountlogs"] = accountlogs.CheckAccountLogsCommand
	}

	dc.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handler, ok := commandHandlers[i.ApplicationCommandData().Name]
		if ok {
			handler(s, i)
		}
	})

	return dc, nil
}

func StopBot() error {
	err := dc.Close()
	if err != nil {
		logger.Log.WithError(err).Error("Error closing connection to Discord")
		return err
	}

	guilds, err := dc.UserGuilds(100, "", "")
	if err != nil {
		logger.Log.WithError(err).Error("Error getting guilds")
		return err
	}

	for _, guild := range guilds {
		logger.Log.WithField("guild", guild.Name).Info("Disconnected from guild")

		addaccount.UnregisterCommand(dc, guild.ID)
		removeaccount.UnregisterCommand(dc, guild.ID)
		accountlogs.UnregisterCommand(dc, guild.ID)
	}

	return nil
}
