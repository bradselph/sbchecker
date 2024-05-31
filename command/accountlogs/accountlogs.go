package accountlogs

import (
	"fmt"

	"codstatusbot/database"
	"codstatusbot/logger"
	"codstatusbot/models"
	"codstatusbot/services"

	"github.com/bwmarrin/discordgo"
)

var choices []*discordgo.ApplicationCommandOptionChoice

// RegisterCommand registers the "accountlogs" command for a specific guild.
func RegisterCommand(s *discordgo.Session, guildID string) {
	choices = services.GetAllChoices(guildID)
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

	existingCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting application commands")
		return
	}

	var existingCommand *discordgo.ApplicationCommand
	for _, command := range existingCommands {
		if command.Name == "accountlogs" {
			existingCommand = command
			break
		}
	}

	newCommand := commands[0]

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

// CommandAccountLogs handles the "accountlogs" command.
func CommandAccountLogs(s *discordgo.Session, i *discordgo.InteractionCreate) {
	logger.Log.Info("Received accountlogs command")

	userID := i.Member.User.ID
	accountID := i.ApplicationCommandData().Options[0].IntValue()
	logger.Log.Infof("User ID: %s, Account ID: %d", userID, accountID)

	var account models.Account
	database.DB.Where("id = ?", accountID).First(&account)
	logger.Log.Infof("Account: %+v", account)

	if account.UserID != userID {
		logger.Log.WithFields(map[string]interface{}{
			"account_id": accountID,
			"user_id":    userID,
		}).Warn("User tried to view logs for account they don't own")
		return
	}

	var logs []models.Ban
	database.DB.Where("account_id = ?", accountID).Order("created_at desc").Limit(5).Find(&logs)

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s - %s", account.Title, account.LastStatus),
		Description: "The last 5 logs for this account",
		Color:       0x00ff00,
		Fields:      make([]*discordgo.MessageEmbedField, len(logs)),
	}

	for i, log := range logs {
		embed.Fields[i] = &discordgo.MessageEmbedField{
			Name:   string(log.Status),
			Value:  log.CreatedAt.String(),
			Inline: false,
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func getAllChoices(guildID string) []*discordgo.ApplicationCommandOptionChoice {
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
