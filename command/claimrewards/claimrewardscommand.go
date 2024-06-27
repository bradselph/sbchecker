package claimrewards

import (
	"fmt"
	"strings"

	"codstatusbot2.0/database"
	"codstatusbot2.0/logger"
	"codstatusbot2.0/models"
	"codstatusbot2.0/services"
	"github.com/bwmarrin/discordgo"
)

func RegisterCommand(s *discordgo.Session, guildID string) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "claimavailablerewards",
			Description: "Claim available rewards for an account",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionType(discordgo.InteractionApplicationCommandAutocomplete),
					Name:        "account",
					Description: "The title of the account",
					Required:    true,
					Choices:     getAllChoices(guildID),
				},
			},
		},
	}

	existingCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting application commands from guild")
		return
	}

	var existingCommand *discordgo.ApplicationCommand
	for _, command := range existingCommands {
		if command.Name == "claimavailablerewards" {
			existingCommand = command
			break
		}
	}
	newCommand := commands[0]
	if existingCommand != nil {
		logger.Log.Info("Updating claimavailablerewards command")
		_, err = s.ApplicationCommandEdit(s.State.User.ID, guildID, existingCommand.ID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error updating the claimavailablerewards command")
			return
		}
	} else {
		logger.Log.Info("Adding the claimavailablerewards command")
		_, err = s.ApplicationCommandCreate(s.State.User.ID, guildID, newCommand)
	}
	if err != nil {
		logger.Log.WithError(err).Error("Error adding the claimavailablerewards command")
		return
	}
}

func UnregisterCommand(s *discordgo.Session, guildID string) {
	commands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting application commands")
		return
	}

	for _, command := range commands {
		logger.Log.Infof("Deleting command %s", command.Name)
		err := s.ApplicationCommandDelete(s.State.User.ID, guildID, command.ID)
		if err != nil {
			logger.Log.WithError(err).Errorf("Error unregistering the command %s", command.Name)
			continue
		}
	}
}

func CommandClaimRewards(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	userID := i.Member.User.ID
	accountTitle := i.ApplicationCommandData().Options[0].StringValue()

	var account models.Account
	result := database.DB.Where("title = ? AND user_id = ?", accountTitle, userID).First(&account)
	if result.Error != nil {
		logger.Log.WithError(result.Error).Error("Error retrieving account")
		sendFollowUpMessage(s, i, "Error retrieving account information")
		return
	}

	rewardResults := claimRewards(account.SSOCookie)

	message := fmt.Sprintf("Reward claim results for %s:\n%s", account.Title, strings.Join(rewardResults, "\n"))
	sendFollowUpMessage(s, i, message)
}

func claimRewards(ssoCookie string) []string {
	results := []string{}
	codes := getRewardCodes()

	for _, code := range codes {
		result, err := services.ClaimSingleReward(ssoCookie, code) // Use the services package function
		if err != nil {
			results = append(results, fmt.Sprintf("Failed to claim reward for code %s: %v", code, err))
		} else {
			results = append(results, result)
		}
	}

	return results
}

func sendFollowUpMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: content,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		logger.Log.WithError(err).Error("Error sending follow-up message")
	}
}

func getRewardCodes() []string {
	return []string{
		"XBLX3HN7X77NAH7", // Blue MONSTER ENERGY OPERATOR
		"KCZEKWKW6PCW3M3", // Double BattlePass XP
		"APSBAXLFVZCXVKW", // Double BattlePass XP
		"WK3SEKWCXSAE6EH", // Double BattlePass XP
		"YPFYEMWXWCVKVLX", // Limited Weapon Charm
	}
}

func getAllChoices(guildID string) []*discordgo.ApplicationCommandOptionChoice {
	logger.Log.Info("Getting all choices for account select dropdown")
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
