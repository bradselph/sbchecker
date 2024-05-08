package internal

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"sbchecker/internal/database"
	"sbchecker/models"
)

func GenerateHeaders(ssoCookie string) map[string]string {
	return map[string]string{
		"accept":             "*/*",
		"accept-language":    "en-US,en;q=0.9,es;q=0.8",
		"sec-ch-ua":          `"Not /ABrand";v="99", "Google Chrome";v="115", "Chromium";v="115"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-origin",
		"x-requested-with":   "XMLHttpRequest",
		"cookie":             fmt.Sprintf("ACT_SSO_COOKIE=%s", ssoCookie),
		"Referrer-Policy":    "strict-origin-when-cross-origin",
	}
}

func GetAllChoices(guildID string) []*discordgo.ApplicationCommandOptionChoice {
	var accounts []models.Account
	// Query the database to find all accounts with the given guild ID.
	database.DB.Where("guild_id = ?", guildID).Find(&accounts)

	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(accounts))
	for i, account := range accounts {
		// Create a new ApplicationCommandOptionChoice for each account.
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  account.Title,
			Value: account.ID,
		}
	}

	return choices
}
