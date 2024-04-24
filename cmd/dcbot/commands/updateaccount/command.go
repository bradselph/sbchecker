package updateaccount

import (
	"github.com/bwmarrin/discordgo"
	"sbchecker/internal"
	"sbchecker/internal/database"
	"sbchecker/internal/logger"
	"sbchecker/models"
)

func RegisterCommand(s *discordgo.Session, guildID string) {
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
					Choices:     internal.GetAllChoices(guildID), // Use the GetAllChoices function
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
	// Similar to other commands, remove this command for the given guild
}

func Command(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	guildID := i.GuildID
	accountId := i.ApplicationCommandData().Options[0].IntValue()
	newSSOCookie := i.ApplicationCommandData().Options[1].StringValue()

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

	account.SSOCookie = newSSOCookie
	database.DB.Save(&account)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Account SSO cookie updated",
			Flags:   64, // Set ephemeral flag
		},
	})
}
