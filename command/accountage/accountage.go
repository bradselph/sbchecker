package accountage

import (
	"fmt"

	"codstatusbot/database"
	"codstatusbot/logger"
	"codstatusbot/models"
	"codstatusbot/services"

	"github.com/bwmarrin/discordgo"
)

var choices []*discordgo.ApplicationCommandOptionChoice

// RegisterCommand registers the "accountage" command for a given guild.
func RegisterCommand(s *discordgo.Session, guildID string) {
	choices = services.GetAllChoices(guildID)
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
					Choices:     choices,
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

// UnregisterCommand deletes all application commands for a given guild.
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

// CommandAccountAge handles the "accountage" command.
func CommandAccountAge(s *discordgo.Session, i *discordgo.InteractionCreate) {
	logger.Log.Info("Starting account age command")
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		logger.Log.WithError(err).Error("Error sending deferred response")
		return
	}
	userID := i.Member.User.ID
	AccountID := i.ApplicationCommandData().Options[0].IntValue()
	logger.Log.Infof("User ID: %s, Account ID: %d", userID, AccountID)

	var account models.Account
	database.DB.Where("id = ?", AccountID).First(&account)
	logger.Log.Infof("Account: %+v", account)

	if account.UserID != userID {
		logger.Log.WithFields(map[string]interface{}{
			"account_id": AccountID,
			"user_id":    userID,
		}).Warn("User tried to check age for account they don't own")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You do not own this account.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	statusCode, err := services.VerifySSOCookie(account.SSOCookie)
	if err != nil {

		logger.Log.WithError(err).Errorf("Error verifying SSO cookie for account %s ", account.Title)
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error verifying SSO cookie.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if statusCode != 200 {
		logger.Log.Errorf("Invalid SSO cookie for account %s ", account.Title)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid SSO cookie. Please update the cookie using the /updateaccount command.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	years, months, days, err := services.CheckAccountAge(account.SSOCookie)
	if err != nil {
     logger.Log.WithError(err).Errorf("Error checking account age for account %s ", account.Title)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error checking account age.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
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
}
