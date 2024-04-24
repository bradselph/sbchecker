package accountage

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"sbchecker/internal/database"
	"sbchecker/internal/logger"
	"sbchecker/models"
	"sbchecker/internal/services"
)

var choices []*discordgo.ApplicationCommandOptionChoice

func RegisterCommand(s *discordgo.Session, guildID string) {
	choices = getAllChoices(guildID)

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

func CheckAccountAgeCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	accountId := i.ApplicationCommandData().Options[0].IntValue()

	var account models.Account
	database.DB.Where("id = ?", accountId).First(&account)

	if account.UserID != userID {
		logger.Log.WithFields(map[string]interface{}{
			"account_id": accountId,
			"user_id":    userID,
		}).Warn("User tried to check age for account they don't own")
		return
	}

	years, months, days, err := services.CheckAccountAge(account.SSOCookie)
	if err != nil {
		logger.Log.WithError(err).Errorf("Error checking account age for account %s", account.Title)
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
