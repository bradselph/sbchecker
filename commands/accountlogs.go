package commands

import (
	"fmt"

	"codstatusbot/database"
	"codstatusbot/internal"
	"codstatusbot/internal/logger"
	"codstatusbot/models"

	"github.com/bwmarrin/discordgo"
)

// choices holds the available options for the account logs command.
var choices []*discordgo.ApplicationCommandOptionChoice

// RegisterCommand registers the "accountlogs" command for a specific guild.
func RegisterCommand(s *discordgo.Session, guildID string) {
	// Get all choices for the guild.
	choices = internal.GetAllChoices(guildID)

	// Define the "accountlogs" command.
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "accountlogs",
			Description: "View the logs for an account",
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

	// Fetch existing application commands.
	existingCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting application commands")
		return
	}

	// Check if the "accountlogs" command already exists.
	var existingCommand *discordgo.ApplicationCommand
	for _, command := range existingCommands {
		if command.Name == "accountlogs" {
			existingCommand = command
			break
		}
	}

	newCommand := commands[0]

	// If the command exists, update it. Otherwise, create a new one.
	if existingCommand != nil {
		logger.Log.Info("Updating accountlogs command")
		_, err = s.ApplicationCommandEdit(s.State.User.ID, guildID, existingCommand.ID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error updating accountlogs command")
			return
		}
	} else {
		logger.Log.Info("Creating accountlogs command")
		_, err = s.ApplicationCommandCreate(s.State.User.ID, guildID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error creating accountlogs command")
			return
		}
	}
}

// UnregisterCommand removes all application commands for a specific guild.
func UnregisterCommand(s *discordgo.Session, guildID string) {
	// Fetch existing application commands.
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

// CommandAccountLogs handles the "accountlogs" command.
func CommandAccountLogs(s *discordgo.Session, i *discordgo.InteractionCreate) {
	logger.Log.Info("Received accountlogs command")

	// Get the user ID and account ID from the interaction.
	userID := i.Member.User.ID
	accountId := i.ApplicationCommandData().Options[0].IntValue()
	logger.Log.Infof("User ID: %s, Account ID: %d", userID, accountId)

	// Fetch the account from the database.
	var account models.Account
	database.DB.Where("id = ?", accountId).First(&account)
	logger.Log.Infof("Account: %+v", account)

	// Check if the user owns the account.
	if account.UserID != userID {
		logger.Log.WithFields(map[string]interface{}{
			"account_id": accountId,
			"user_id":    userID,
		}).Warn("User tried to view logs for account they don't own")
		return
	}

	// Fetch the last 5 logs for the account.
	var logs []models.Ban
	database.DB.Where("account_id = ?", accountId).Order("created_at desc").Limit(5).Find(&logs)

	// Create an embed for the logs.
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s - %s", account.Title, account.LastStatus),
		Description: "The last 5 logs for this account",
		Color:       0x00ff00,
		Fields:      make([]*discordgo.MessageEmbedField, len(logs)),
	}

	// Add each log to the embed.
	for i, log := range logs {
		embed.Fields[i] = &discordgo.MessageEmbedField{
			Name:   string(log.Status),
			Value:  log.CreatedAt.String(),
			Inline: false,
		}
	}

	// Respond to the interaction with the embed.
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}
