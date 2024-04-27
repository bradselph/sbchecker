package removeaccount

import (
	"github.com/bwmarrin/discordgo"
	"sbchecker/internal"
	"sbchecker/internal/database"
	"sbchecker/internal/logger"
	"sbchecker/models"
)

var choices []*discordgo.ApplicationCommandOptionChoice

func RegisterCommand(s *discordgo.Session, guildID string) {
	choices = getAllChoices(guildID)
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
					Choices:     choices,
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
		}
	}()

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
		tx.Rollback()
		return
	}

	// Disable foreign key constraints
	err := tx.Exec("SET FOREIGN_KEY_CHECKS=0;").Error
	if err != nil {
		logger.Log.WithError(err).Error("Error disabling foreign key constraints")
		tx.Rollback()
		return
	}
	defer tx.Exec("SET FOREIGN_KEY_CHECKS=1;") // Re-enable foreign key constraints

	// Delete associated bans
	if err := tx.Unscoped().Where("account_id = ?", account.ID).Delete(&models.Ban{}).Error; err != nil {
		logger.Log.WithError(err).Error("Error deleting associated bans for account", account.ID)
		tx.Rollback()
		return
	}

	// Delete the account
	if err := tx.Unscoped().Where("id = ?", account.ID).Delete(&models.Account{}).Error; err != nil {
		logger.Log.WithError(err).Error("Error deleting account from database", account.ID)
		tx.Rollback()
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Account removed",
			Flags:   64, // Set ephemeral flag
		},
	})

	UpdateAccountChoices(s, guildID)

	tx.Commit()
}

func getAllChoices(guildID string) []*discordgo.ApplicationCommandOptionChoice {
	return internal.GetAllChoices(guildID)
}

func UpdateAccountChoices(s *discordgo.Session, guildID string) {
	choices = getAllChoices(guildID)
	commands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting application command choices")
		return
	}

	commandConfigs := map[string]*discordgo.ApplicationCommand{
		"removeaccount": {
			Name:        "removeaccount",
			Description: "Remove an account from automated shadowban checking",
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
		"accountlogs": {
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
		"updateaccount": {
			Name:        "updateaccount",
			Description: "Update the SSO cookie for an account",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionType(discordgo.InteractionApplicationCommandAutocomplete),
					Name:        "account",
					Description: "The title of the account",
					Required:    true,
					Choices:     choices,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "sso_cookie",
					Description: "The new SSO cookie for the account",
					Required:    true,
				},
			},
		},
		"accountage": {
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

	for _, command := range commands {
		if config, ok := commandConfigs[command.Name]; ok {
			logger.Log.Infof("Updating command %s", command.Name)
			_, err := s.ApplicationCommandEdit(s.State.User.ID, guildID, command.ID, config)
			if err != nil {
				logger.Log.WithError(err).Errorf("Error updating command %s", command.Name)
				return
			}
		}
	}
}
