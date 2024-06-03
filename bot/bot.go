package bot

import (
	"codstatusbot2.0/command"
	"codstatusbot2.0/services"
	"errors"
	"os"

	"codstatusbot2.0/logger"
	"github.com/bwmarrin/discordgo"
)

var discord *discordgo.Session

func StartBot() error {
	envToken := os.Getenv("DISCORD_TOKEN")
	if envToken == "" {
		err := errors.New("DISCORD_TOKEN environment variable not set")
		logger.Log.WithError(err).WithField("env", "DISCORD_TOKEN").Error()
		return err
	}
	var err error
	discord, err = discordgo.New("Bot " + envToken)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Token").Error()
		return err
	}

	err = discord.Open()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Opening Session").Error()
		return err
	}

	err = discord.UpdateWatchStatus(0, "the Status of your Accounts so you dont have to.")
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Setting Presence Status").Error()
		return err
	}

	guilds, err := discord.UserGuilds(100, "", "", false)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Initiating Guilds").Error()
		return err
	}
	for _, guild := range guilds {
		logger.Log.WithField("guild", guild.Name).Info("Connected to guild")
		command.RegisterCommands(discord, guild.ID)
	}

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handler, ok := command.Handlers[i.ApplicationCommandData().Name]
		if ok {
			logger.Log.WithField("command", i.ApplicationCommandData().Name).Info("Handling command")
			handler(s, i)
		} else {
			logger.Log.WithField("command", i.ApplicationCommandData().Name).Error("Command handler not found")
		}
	})
	discord.AddHandler(OnGuildCreate)
	discord.AddHandler(OnGuildDelete)
	go services.CheckAccounts(discord)
	return nil
}

func OnGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	guildID := event.Guild.ID
	logger.Log.WithField("guild", guildID).Info("Bot joined server:")
	command.RegisterCommands(s, guildID)
}

func OnGuildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	guildID := event.Guild.ID
	logger.Log.WithField("guild", guildID).Info("Bot left guild")
	command.UnregisterCommands(s, guildID)
}

// Possibly unnecessary; not sure yet. Commenting it out for now.
/*	func StopBot() error {
	logger.Log.Info("Bot is shutting down")
	guilds, err := discord.UserGuilds(100, "", "", false)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Shutdown", "Disconnecting Guilds").Error()
		return err
	}
	for _, guild := range guilds {
		logger.Log.WithField("guild", guild.Name).Info("Disconnected from Guild")
	}
	err = discord.Close()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Shutdown", "Closing Session").Error()
		return err
	}
	return nil
}

func RestartBot() error {
	logger.Log.Info("Restarting bot")
	err := StopBot()
	if err != nil {
		logger.Log.WithError(err).Error("Error stopping bot")
		return err
	}

	err = StartBot()
	if err != nil {
		logger.Log.WithError(err).Error("Error starting bot")
		return err
	}
	logger.Log.Info("Bot restarted successfully")
	return nil
}
*/
