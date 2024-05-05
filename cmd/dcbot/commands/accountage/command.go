package accountage

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"sbchecker/internal/database"
	"sbchecker/internal/logger"
	"sbchecker/internal/services"
	"sbchecker/models"
)

// choices holds the list of account choices.
var choices []*discordgo.ApplicationCommandOptionChoice

// RegisterCommand registers the "accountage" command for a given guild.
func RegisterCommand(s *discordgo.Session, guildID string) {
	// Get all account choices for the guild.
	choices = getAllChoices(guildID)

	// Define the "accountage" command.
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

	// Get existing application commands for the guild.
	existingCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting application commands")
		return
	}

	// Check if the "accountage" command already exists.
	var existingCommand *discordgo.ApplicationCommand
	for _, command := range existingCommands {
		if command.Name == "accountage" {
			existingCommand = command
			break
		}
	}

	newCommand := commands[0]

	// If the command exists, update it. Otherwise, create a new one.
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
	// Get existing application commands for the guild.
	commands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting application commands")
		return
	}

	// Delete each command.
	for _, command := range commands {
		logger.Log.Infof("Deleting command %s", command.Name)
		err := s.ApplicationCommandDelete(s.State.User.ID, guildID, command.ID)
		if err != nil {
			logger.Log.WithError(err).Errorf("Error deleting command %s", command.Name)
			return
		}
	}
}

// CommandAccountAge handles the "accountage" command.
func CommandAccountAge(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Get the user ID and account ID from the interaction.
	userID := i.Member.User.ID
	accountId := i.ApplicationCommandData().Options[0].IntValue()

	// Get the account from the database.
	var account models.Account
	database.DB.Where("id = ?", accountId).First(&account)

	// If the account does not belong to the user, log a warning and return.
	if account.UserID != userID {
		logger.Log.WithFields(map[string]interface{}{
			"account_id": accountId,
			"user_id":    userID,
		}).Warn("User tried to check age for account they don't own")
		return
	}

	// Verify the SSO cookie.
	statusCode, err := services.VerifySSOCookie(account.SSOCookie)
	if err != nil {
		logger.Log.WithError(err).Errorf("Error verifying SSO cookie for account %s", account.Title)
		// Handle the error case (e.g., send a notification to the user)
		return
	}

	// If the status code is not 200, the SSO cookie is invalid.
	if statusCode != 200 {
		logger.Log.Errorf("Invalid SSO cookie for account %s", account.Title)
		// Handle the invalid cookie case (e.g., send a notification to the user, mark the account as having an expired cookie)
		return
	}

	// The SSO cookie is valid, proceed to check the account age.
	// Check the age of the account.
	years, months, days, err := services.CheckAccountAge(account.SSOCookie)
	if err != nil {
		logger.Log.WithError(err).Errorf("Error checking account age for account %s", account.Title)
		return
	}

	// Create an embed for the response.
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s - %s", account.Title, account.LastStatus),
		Description: fmt.Sprintf("The account is %d years, %d months, and %d days old.", years, months, days),
		Color:       0x00ff00,
	}

	// Respond to the interaction with the embed.
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// getAllChoices gets all account choices for a given guild.
func getAllChoices(guildID string) []*discordgo.ApplicationCommandOptionChoice {
	// Get all accounts for the guild from the database.
	var accounts []models.Account
	database.DB.Where("guild_id = ?", guildID).Find(&accounts)
	// Create a list of choices from the accounts.
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(accounts))
	for i, account := range accounts {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  account.Title,
			Value: account.ID,
		}
	}

	return choices
}
