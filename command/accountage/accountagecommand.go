package accountage

import (
	"fmt"

	"codstatusbot2.0/database"
	"codstatusbot2.0/logger"
	"codstatusbot2.0/models"
	"codstatusbot2.0/services"
	"github.com/bwmarrin/discordgo"
)

func RegisterCommand(s *discordgo.Session, guildID string) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "accountage",
			Description: "Check the age of an account",
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
		logger.Log.WithError(err).Error("Error getting application commands")
		return
	}

	var existingCommand *discordgo.ApplicationCommand
	for _, command := range existingCommands {
		if command.Name == "accountage" {
			existingCommand = command
			break
		}
	}

	newCommand := commands[0]

	if existingCommand != nil {
		logger.Log.Info("Updating accountage command")
		_, err = s.ApplicationCommandEdit(s.State.User.ID, guildID, existingCommand.ID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error updating accountage command")
			return
		}
	} else {
		logger.Log.Info("Creating accountage command")
		_, err = s.ApplicationCommandCreate(s.State.User.ID, guildID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error creating accountage command")
			return
		}
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
			logger.Log.WithError(err).Errorf("Error deleting command %s", command.Name)
			return
		}
	}
}

func CommandAccountAge(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	accountId := i.ApplicationCommandData().Options[0].IntValue()

	var account models.Account
	database.DB.Where("id = ?", accountId).First(&account)

	if account.UserID != userID {
		logger.Log.WithFields(map[string]interface{}{
			"account_id": accountId,
			"user_id":    userID,
		}).Warn("User tried to check age for account they don't own")
		return
	}

	go func() {
		statusCode, err := services.VerifySSOCookie(account.SSOCookie)
		if err != nil || statusCode != 200 {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Invalid or Error verifying SSO cookie",
			})
			return
		}

		years, months, days, err := services.CheckAccountAge(account.SSOCookie)
		if err != nil {
			logger.Log.WithError(err).Errorf("Error checking account age for account %s", account.Title)
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("%s - %s", account.Title, account.LastStatus),
			Description: fmt.Sprintf("The account is %d years, %d months, and %d days old.", years, months, days),
			Color:       0x00ff00,
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	}()
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
