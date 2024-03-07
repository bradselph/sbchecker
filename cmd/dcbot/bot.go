package bot

import (
	"github.com/bwmarrin/discordgo"
	"sbchecker/internal/discord"
	"sbchecker/internal/logger"
)

func RunBot() (*discordgo.Session, error) {
	err := discord.Initialize()
	if err != nil {
		logger.Log.WithError(err).Error("Error initializing discord")
		return nil, err
	}

	instance, err := discord.StartBot()
	if err != nil {
		logger.Log.WithError(err).Error("Error running discord")
		return nil, err
	}

	return instance, nil
}
