package updateaccount

import (
	"github.com/bwmarrin/discordgo"
	"sbchecker/internal"
	"sbchecker/internal/database"
	"sbchecker/internal/logger"
	"sbchecker/internal/services"
	"sbchecker/models"
)

// RegisterCommand registers the "updateaccount" command in the Discord session.
// It creates or updates the command based on its existence.
func RegisterCommand(s *discordgo.Session, guildID string) {
	// Define the command with its options
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
					Choices:     internal.GetAllChoices(guildID), // Use the GetAllChoices function to get all account choices
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

	// Fetch existing commands
	existingCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting application commands")
		return
	}

	// Check if the "updateaccount" command already exists
	var existingCommand *discordgo.ApplicationCommand
	for _, command := range existingCommands {
		if command.Name == "updateaccount" {
			existingCommand = command
			break
		}
	}

	newCommand := commands[0]

	// If the command exists, update it. Otherwise, create a new one.
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
	// Fetch existing commands
	commands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting application commands")
		return
	}

	// Delete each command
	for _, command := range commands {
		logger.Log.Infof("Deleting command %s", command.Name)
		err := s.ApplicationCommandDelete(s.State.User.ID, guildID, command.ID)
		if err != nil {
			logger.Log.WithError(err).Errorf("Error deleting command %s", command.Name)
			return
		}
	}
}

// CommandUpdateAccount handles the "updateaccount" command interaction.
// It updates the SSO cookie for a specific account.
func CommandUpdateAccount(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Start a new database transaction
	tx := database.DB.Begin()
	defer func() {
		// Rollback the transaction in case of panic
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Extract data from the interaction
	userID := i.Member.User.ID
	guildID := i.GuildID
	accountId := i.ApplicationCommandData().Options[0].IntValue()
	newSSOCookie := i.ApplicationCommandData().Options[1].StringValue()

	// Verify the new SSO cookie
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

	// Check if the SSO cookie is valid
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

	// Fetch the account from the database
	var account models.Account
	result := tx.Where("user_id = ? AND id = ? AND guild_id = ?", userID, accountId, guildID).First(&account)
	if result.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Account does not exist",
				Flags:   64, // Set ephemeral flag
			},
		})
		return
	}

	// Update the account's SSO cookie and status
	account.SSOCookie = newSSOCookie
	account.LastStatus = models.StatusUnknown // Reset the status to prevent further notifications
	tx.Save(&account)
	tx.Commit()

	// Respond to the interaction
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Account SSO cookie updated",
			Flags:   64, // Set ephemeral flag
		},
	})
}
