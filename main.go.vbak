package main

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"codstatusbot/commands"
	"codstatusbot/database"
	"codstatusbot/logger"
	"codstatusbot/services"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var session *discordgo.Session
var commandHandlers = map[string]func(*discordgo.Session, *discordgo.InteractionCreate){}

func main() {
	logger.Log.Info("Bot starting...")
	err := loadEnvironmentVariables()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Environment Variables").Error()
		os.Exit(1)
	}
	err = database.Databaselogin()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Databaselogin").Error()
		os.Exit(1)
	}
	err = startBot()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Discordlogin").Error()
		os.Exit(1)
	}
	logger.Log.Info("Bot is running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	err = stopBot()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Shutdown", "Shutdown Process").Error()
		os.Exit(1)
	}
}

func loadEnvironmentVariables() error {
	err := godotenv.Load()
	if err != nil {
		logger.Log.WithError(err).Error("Error loading .env file")
		return err
	}
	return nil
}

func startBot() error {
	envToken := os.Getenv("DISCORD_TOKEN")
	if envToken == "" {
		err := errors.New("DISCORD_TOKEN environment variable not set")
		logger.Log.WithError(err).WithField("env", "DISCORD_TOKEN").Error()
		return err
	}
	var err error
	session, err = discordgo.New("Bot " + envToken)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Token").Error()
		return err
	}

	err = session.Open()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Opening Session").Error()
		return err
	}

	err = session.UpdateWatchStatus(0, "the Status of your Accounts so you dont have to.")
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Setting Presence Status").Error()
		return err
	}

	guilds, err := session.UserGuilds(100, "", "", false)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Initiating Guilds").Error()
		return err
	}

	for _, guild := range guilds {
		logger.Log.WithField("guild", guild.Name).Info("Connected to guild")
		commands.RegisterCommands(session, guild.ID)
	}

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handler, ok := commandHandlers[i.ApplicationCommandData().Name]
		if ok {
			handler(s, i)
		}
	})

	session.AddHandler(OnGuildCreate)
	session.AddHandler(OnGuildDelete)
	go services.CheckAccounts(session)

	return nil
}

func stopBot() error {
	logger.Log.Info("Bot is shutting down")
	guilds, err := session.UserGuilds(100, "", "", false)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Shutdown", "Disconnecting Guilds").Error()
		return err
	}
	for _, guild := range guilds {
		logger.Log.WithField("guild", guild.Name).Info("Disconnected from Guild")

	}
	err = session.Close()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Shutdown", "Closing Session").Error()
		return err
	}
	return nil
}

func restartBot() error {
	err := stopBot()
	if err != nil {
		logger.Log.WithError(err).Error("Error stopping bot")
		return err
	}

	err = startBot()
	if err != nil {
		logger.Log.WithError(err).Error("Error starting bot")
		return err
	}
	logger.Log.Info("Bot restarted successfully")
	return nil
}

// OnGuildCreate is called when the bot joins a guild.
func OnGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	guildID := event.Guild.ID
	logger.Log.WithField("guild", guildID).Info("Bot joined server:")
	commands.RegisterCommands(s, guildID)
}

// OnGuildDelete is called when the bot leaves a guild.
func OnGuildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	guildID := event.Guild.ID
	logger.Log.WithField("guild", guildID).Info("Bot left guild")
	commands.UnregisterCommands(s, guildID)
}
