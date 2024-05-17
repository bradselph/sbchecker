package services

import (
	"codstatusbot/database"
	"codstatusbot/models"

	"github.com/bwmarrin/discordgo"
)

// GetAllChoices returns all choices for the account select dropdown.
func GetAllChoices(guildID string) []*discordgo.ApplicationCommandOptionChoice {
	var accounts []models.Account
	database.DB.Where("guild_id = ?", guildID).Find(&accounts)

	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(accounts))
	for i, account := range accounts {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  account.Title,
			Value: account.ID,
		}
	}

	return choices
}
