package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/silenta-salmans/sbchecker/internal/discord"
	"github.com/silenta-salmans/sbchecker/internal/logger"
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
