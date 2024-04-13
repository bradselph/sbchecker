package discord

import (
	"os"
	"sbchecker/cmd/dcbot/commands/updateaccount"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"sbchecker/cmd/dcbot/commands/accountlogs"
	"sbchecker/cmd/dcbot/commands/addaccount"
	"sbchecker/cmd/dcbot/commands/removeaccount"
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

	guilds, err := dc.UserGuilds(100, "", "", false)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting guilds")
		return nil, err
	}

	for _, guild := range guilds {
		logger.Log.WithField("guild", guild.Name).Info("Connected to guild")

		addaccount.RegisterCommand(dc, guild.ID)
		commandHandlers["addaccount"] = addaccount.Command

		removeaccount.RegisterCommand(dc, guild.ID)
		commandHandlers["removeaccount"] = removeaccount.Command

		accountlogs.RegisterCommand(dc, guild.ID)
		commandHandlers["accountlogs"] = accountlogs.CheckAccountLogsCommand

		updateaccount.RegisterCommand(dc, guild.ID)
		commandHandlers["updateaccount"] = updateaccount.Command
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

	guilds, err := dc.UserGuilds(100, "", "", false)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting guilds")
		return err
	}

	for _, guild := range guilds {
		logger.Log.WithField("guild", guild.Name).Info("Disconnected from guild")

		addaccount.UnregisterCommand(dc, guild.ID)
		removeaccount.UnregisterCommand(dc, guild.ID)
		accountlogs.UnregisterCommand(dc, guild.ID)
		updateaccount.UnregisterCommand(dc, guild.ID)
	}

	return nil
}

func RestartBot() error {
	if err := StopBot(); err != nil {
		return err
	}
	if _, err := StartBot(); err != nil {
		return err
	}
	return nil
}
