package updateaccount

import (
	"codstatusbot2.0/database"
	"codstatusbot2.0/logger"
	"codstatusbot2.0/models"
	"codstatusbot2.0/services"
	"github.com/bwmarrin/discordgo"
)

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

func RegisterCommand(s *discordgo.Session, guildID string) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "updateaccount",
			Description: "Update the SSO cookie for an account",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionType(discordgo.InteractionApplicationCommandAutocomplete),
					Name:        "account",
					Description: "The title of the account",
					Required:    true,
					Choices:     getAllChoices(guildID),
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "sso_cookie",
					Description: "The new SSO cookie for the account",
					Required:    true,
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
		if command.Name == "updateaccount" {
			existingCommand = command
			break
		}
	}

	newCommand := commands[0]

	if existingCommand != nil {
		logger.Log.Info("Updating updateaccount command")
		_, err := s.ApplicationCommandEdit(s.State.User.ID, guildID, existingCommand.ID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error updating updateaccount command")
			return
		}
	} else {
		logger.Log.Info("Creating updateaccount command")
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error creating updateaccount command")
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

func CommandUpdateAccount(s *discordgo.Session, i *discordgo.InteractionCreate) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	userID := i.Member.User.ID
	guildID := i.GuildID
	accountId := i.ApplicationCommandData().Options[0].IntValue()
	newSSOCookie := i.ApplicationCommandData().Options[1].StringValue()

	var account models.Account
	result := tx.Where("user_id = ? AND id = ? AND guild_id = ?", userID, accountId, guildID).First(&account)
	if result.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Account does not exist",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	go func() {
		statusCode, err := services.VerifySSOCookie(newSSOCookie)
		if err != nil || statusCode != 200 {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Invalid or Error verifying new SSO cookie",
			})
			return
		}

		account.SSOCookie = newSSOCookie
		account.LastStatus = models.StatusUnknown
		account.IsExpiredCookie = false
		account.LastCookieNotification = 0
		tx.Save(&account)
		tx.Commit()
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Account SSO cookie updated",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}()
}
