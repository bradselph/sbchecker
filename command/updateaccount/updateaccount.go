package updateaccount

import (
	"github.com/bwmarrin/discordgo"

	"codstatusbot/database"
	"codstatusbot/logger"
	"codstatusbot/models"
	"codstatusbot/services"
)

// RegisterCommand registers the "updateaccount" command in the Discord session.
// It creates or updates the command based on its existence.
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
					Choices:     services.GetAllChoices(guildID),
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
		_, err = s.ApplicationCommandEdit(s.State.User.ID, guildID, existingCommand.ID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error updating updateaccount command")
			return
		}
	} else {
		logger.Log.Info("Creating updateaccount command")
		_, err = s.ApplicationCommandCreate(s.State.User.ID, guildID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error creating updateaccount command")
			return
		}
	}
}

// UnregisterCommand deletes all commands from the Discord session.
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
			logger.Log.WithError(err).Errorf("Error deleting command %s ", command.Name)
			return
		}
	}
}

// CommandUpdateAccount handles the "updateaccount" command interaction.
// It updates the SSO cookie for a specific account.
func CommandUpdateAccount(s *discordgo.Session, i *discordgo.InteractionCreate) {
	logger.Log.Info("Received updateaccount command")

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	userID := i.Member.User.ID
	guildID := i.GuildID
	accountID := i.ApplicationCommandData().Options[0].IntValue()
	newSSOCookie := i.ApplicationCommandData().Options[1].StringValue()
	logger.Log.Infof("User ID: %s, Guild ID: %s Account ID: %d, New SSO Cookie: %s", userID, guildID, accountID, newSSOCookie)

	statusCode, err := services.VerifySSOCookie(newSSOCookie)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error verifying SSO cookie",
				Flags:   64,
			},
		})
		return
	}

	if statusCode != 200 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid SSO cookie",
				Flags:   64,
			},
		})
		return
	}

	var account models.Account
	result := tx.Where("user_id = ? AND id = ? AND guild_id = ?", userID, accountID, guildID).First(&account)
	if result.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Account does not exist",
				Flags:   64,
			},
		})
		return
	}

	account.SSOCookie = newSSOCookie
	account.LastStatus = models.StatusUnknown // Reset the status to prevent further notifications
	account.IsExpiredCookie = false           // Reset the expired cookie flag to prevent further notifications
	account.LastCookieNotification = 0        // Reset the last cookie notification timestamp to prevent further notifications
	tx.Save(&account)
	tx.Commit()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Account SSO cookie updated",
			Flags:   64,
		},
	})
}
