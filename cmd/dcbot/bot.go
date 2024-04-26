package bot

import (
	"github.com/bwmarrin/discordgo"
	"sbchecker/internal/discord"
	"sbchecker/internal/logger"
	"fmt"
)

func RunBot() (*discordgo.Session, error) {
	err := discord.Initialize()
	if err != nil {
		logger.Log.WithError(err).Error("Error Initializing Discord")
		return nil, fmt.Errorf("error initializing discord: %w", err)
	}
	instance, err := discord.StartBot()
	if err != nil {
		logger.Log.WithError(err).Error("Error Starting Bot")
		return nil, fmt.Errorf("error starting bot: %w", err)
	}
	return instance, nil
}
