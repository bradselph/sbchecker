package addaccount

import (
	"codstatusbot/command/removeaccount"
	"codstatusbot/database"
	"codstatusbot/logger"
	"codstatusbot/models"
	"codstatusbot/services"

	"github.com/bwmarrin/discordgo"
)

// RegisterCommand registers the "addaccount" command in the Discord session for a specific guild.
func RegisterCommand(s *discordgo.Session, guildID string, commands map[string]*discordgo.ApplicationCommand) {
	command := &discordgo.ApplicationCommand{
		Name:        "addaccount",
		Description: "Add or remove an account for shadowban checking",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "title",
				Description: "The title of the account",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "sso_cookie",
				Description: "The SSO cookie for the account",
				Required:    true,
			},
		},
	}

	commands["addaccount"] = command
	existingCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting application commands")
		return
	}

	var existingCommand *discordgo.ApplicationCommand
	for _, command := range existingCommands {
		if command.Name == "addaccount" {
			existingCommand = command
			break
		}
	}

	newCommand := commands[0]

	if existingCommand != nil {
		logger.Log.Info("Updating addaccount command")
		_, err = s.ApplicationCommandEdit(s.State.User.ID, guildID, existingCommand.ID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error updating addaccount command")
			return
		}
	} else {
		logger.Log.Info("Creating addaccount command")
		_, err = s.ApplicationCommandCreate(s.State.User.ID, guildID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error creating addaccount command")
			return
		}
	}
}

// UnregisterCommand removes the "addaccount" command from the Discord session for a specific guild.
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

// CommandAddAccount handles the "addaccount" command when invoked.
func CommandAddAccount(s *discordgo.Session, i *discordgo.InteractionCreate) {
	logger.Log.Info("Invoked addaccount command")
	title := i.ApplicationCommandData().Options[0].StringValue()
	ssoCookie := i.ApplicationCommandData().Options[1].StringValue()
	guildID := i.GuildID
	channelID := i.ChannelID
	userID := i.Member.User.ID

	logger.Log.WithFields(map[string]interface{}{
		"title":      title,
		"sso_cookie": ssoCookie,
		"guild_id":   guildID,
		"channel_id": channelID,
		"user_id":    userID,
	}).Info("Add account command")

	var account models.Account
	result := database.DB.Where("user_id = ? AND title = ?", userID, title).First(&account)
	if result.Error == nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Account already exists",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	statusCode, err := services.VerifySSOCookie(ssoCookie)
	if err != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Error verifying SSO cookie",
		})
		return
	}

	logger.Log.WithField("status_code", statusCode).Info("SSO cookie verification status")
	if statusCode != 200 {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Invalid SSO cookie",
		})
		return
	}

	account = models.Account{
		UserID:    userID,
		Title:     title,
		SSOCookie: ssoCookie,
		GuildID:   guildID,
		ChannelID: channelID,
	}

	result = database.DB.Create(&account)
	if result.Error != nil {
		logger.Log.WithError(result.Error).Error("Error creating account")
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Error creating account",
		})
		return
	}

	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Flags:   discordgo.MessageFlagsEphemeral,
		Content: "Account added",
	})

	// Update the account choices for the "removeaccount" command.
	removeaccount.UpdateAccountChoices(s, guildID)

	go services.CheckSingleAccount(account, s)
}
