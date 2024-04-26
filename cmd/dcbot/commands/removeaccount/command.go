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
		logger.Log.WithError(err).Error("Error getting application commands")
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
		logger.Log.Info("Updating removeaccount command")
		_, err = s.ApplicationCommandEdit(s.State.User.ID, guildID, existingCommand.ID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error updating removeaccount command")
			return
		}
	} else {
		logger.Log.Info("Creating removeaccount command")
		_, err = s.ApplicationCommandCreate(s.State.User.ID, guildID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error creating removeaccount command")
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

func Command(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	guildID := i.GuildID
	accountId := i.ApplicationCommandData().Options[0].IntValue()

	var account models.Account
	result := database.DB.Where("user_id = ? AND id = ? AND guild_id = ?", userID, accountId, guildID).First(&account)
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

	// Disable foreign key constraints
	err := database.DB.Exec("SET FOREIGN_KEY_CHECKS=0;").Error
	if err != nil {
		logger.Log.WithError(err).Error("Error disabling foreign key constraints")
		return
	}
	defer database.DB.Exec("SET FOREIGN_KEY_CHECKS=1;") // Re-enable foreign key constraints

	// Delete associated bans
	if err := database.DB.Where("account_id = ?", account.ID).Delete(&models.Ban{}).Error; err != nil {
		logger.Log.WithError(err).Error("Error deleting associated bans")
		return
	}

	// Delete the account
	if err := database.DB.Unscoped().Delete(&account).Error; err != nil {
		logger.Log.WithError(err).Error("Error deleting account")
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
}

func getAllChoices(guildID string) []*discordgo.ApplicationCommandOptionChoice {
	return internal.GetAllChoices(guildID)
}

func UpdateAccountChoices(s *discordgo.Session, guildID string) {
	choices = getAllChoices(guildID)
	commands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting application commands")
		return
	}

	for _, command := range commands {
		if command.Name == "removeaccount" || command.Name == "accountlogs" {
			logger.Log.Infof("Updating command %s", command.Name)
			_, err := s.ApplicationCommandEdit(s.State.User.ID, guildID, command.ID, &discordgo.ApplicationCommand{
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
			})
			if err != nil {
				logger.Log.WithError(err).Errorf("Error updating command %s", command.Name)
				return
			}
		}
	}
}
