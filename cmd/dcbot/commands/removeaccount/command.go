package removeaccount

import (
	"github.com/bwmarrin/discordgo"
	"github.com/silenta-salmans/sbchecker/internal/database"
	"github.com/silenta-salmans/sbchecker/internal/logger"
	"github.com/silenta-salmans/sbchecker/models"
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

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, command := range commands {
		logger.Log.Infof("Creating command %s", command.Name)
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, command)
		if err != nil {
			logger.Log.WithError(err).Errorf("Error creating command %s", command.Name)
			return
		}

		registeredCommands[i] = cmd
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

func RemoveAccountCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
			},
		})
		return
	}

	database.DB.Delete(&account)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Account removed",
		},
	})

	UpdateAccountChoices(s, guildID)
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
