package accountlogs

import (
	"fmt"

	"codstatusbot2.0/database"
	"codstatusbot2.0/logger"
	"codstatusbot2.0/models"
	"github.com/bwmarrin/discordgo"
)

func RegisterCommand(s *discordgo.Session, guildID string) {
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

func CommandAccountLogs(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	accountId := i.ApplicationCommandData().Options[0].IntValue()

	var account models.Account
	database.DB.Where("id = ?", accountId).First(&account)

	if account.UserID != userID {
		logger.Log.WithFields(map[string]interface{}{
			"account_id": accountId,
			"user_id":    userID,
		}).Warn("User tried to view logs for account they don't own")
		return
	}

	var logs []models.Ban
	database.DB.Where("account_id = ?", accountId).Order("created_at desc").Limit(5).Find(&logs)

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
