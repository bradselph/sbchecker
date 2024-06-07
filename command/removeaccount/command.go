package removeaccount

import (
	"codstatusbot2.0/database"
	"codstatusbot2.0/logger"
	"codstatusbot2.0/models"
	"github.com/bwmarrin/discordgo"
)

func RegisterCommand(s *discordgo.Session, guildID string) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "removeaccount",
			Description: "Remove an account from shadowban checking",
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
		if command.Name == "removeaccount" {
			existingCommand = command
			break
		}
	}
	newCommand := commands[0]
	if existingCommand != nil {
		logger.Log.Info("Updating the removeaccount command")
		_, err = s.ApplicationCommandEdit(s.State.User.ID, guildID, existingCommand.ID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error updating the removeaccount command")
			return
		}
	} else {
		logger.Log.Info("Adding the removeaccount command")
		_, err = s.ApplicationCommandCreate(s.State.User.ID, guildID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error adding the removeaccount command")
			return
		}
	}
}
func UnregisterCommand(s *discordgo.Session, guildID string) {
	commands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting application commands for unregistering removeaccount command")
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
func CommandRemoveAccount(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	guildID := i.GuildID
	accountId := i.ApplicationCommandData().Options[0].IntValue()
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else {
			err := tx.Exec("SET FOREIGN_KEY_CHECKS=1;").Error
			if err != nil {
				logger.Log.WithError(err).Error("Error re-enabling foreign key constraints")
				tx.Rollback()
				return
			}
			tx.Commit()
		}
	}()
	err := tx.Exec("SET FOREIGN_KEY_CHECKS=0;").Error
	if err != nil {
		logger.Log.WithError(err).Error("Error disabling foreign key constraints")
		tx.Rollback()
		return
	}
	var account models.Account
	result := tx.Where("user_id = ? AND id = ? AND guild_id = ?", userID, accountId, guildID).First(&account)
	if result.Error != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Account does not exist",
				Flags:   64,
			},
		})
		tx.Rollback()
		return
	}
	if err := tx.Unscoped().Where("account_id = ?", account.ID).Delete(&models.Ban{}).Error; err != nil {
		logger.Log.WithError(err).Error("Error deleting associated bans for account", account.ID)
		tx.Rollback()
		return
	}
	if err := tx.Unscoped().Where("id = ?", account.ID).Delete(&models.Account{}).Error; err != nil {
		logger.Log.WithError(err).Error("Error deleting account from database", account.ID)
		tx.Rollback()
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Account removed",
			Flags:   64,
		},
	})
	UpdateAccountChoices(s, guildID)
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
func UpdateAccountChoices(s *discordgo.Session, guildID string) {
	commands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting application command choices")
		return
	}
	logger.Log.Info("Updating account choices for commands")
	newChoices := getAllChoices(guildID)
	for _, command := range commands {
		if command.Name == "removeaccount" || command.Name == "accountlogs" || command.Name == "updateaccount" || command.Name == "accountage" {
			newCommand := &discordgo.ApplicationCommand{
				Name:        command.Name,
				Description: command.Description,
				Options:     command.Options,
			}
			for _, option := range newCommand.Options {
				if option.Name == "account" {
					option.Choices = newChoices
				}
			}
			_, err := s.ApplicationCommandEdit(s.State.User.ID, guildID, command.ID, newCommand)
			if err != nil {
				logger.Log.WithError(err).Errorf("Error updating command %s", command.Name)
				return
			}
			logger.Log.Infof("Command %s updated successfully", command.Name)
		}
	}
	logger.Log.Info("Account choices updated successfully")
}
