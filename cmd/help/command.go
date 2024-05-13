package help

import (
	"codstatusbot/internal/logger"

	"github.com/bwmarrin/discordgo"
)

// RegisterCommand registers the "help" command for a given guild.
func RegisterCommand(s *discordgo.Session, guildID string) {
	// Define the "help" command.
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Get help or report an issue",
		},
	}

	// Get existing application commands for the guild.
	existingCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		logger.Log.WithError(err).Error("Error getting application commands")
		return
	}

	// Check if the "help" command already exists.
	var existingCommand *discordgo.ApplicationCommand
	for _, command := range existingCommands {
		if command.Name == "help" {
			existingCommand = command
			break
		}
	}

	newCommand := commands[0]

	// If the command exists, update it. Otherwise, create a new one.
	if existingCommand != nil {
		logger.Log.Info("Updating help command")
		_, err = s.ApplicationCommandEdit(s.State.User.ID, guildID, existingCommand.ID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error updating help command")
			return
		}
	} else {
		logger.Log.Info("Creating help command")
		_, err = s.ApplicationCommandCreate(s.State.User.ID, guildID, newCommand)
		if err != nil {
			logger.Log.WithError(err).Error("Error creating help command")
			return
		}
	}
}

// UnregisterCommand deletes all application commands for a given guild.
func UnregisterCommand(s *discordgo.Session, guildID string) {
	// Get existing application commands for the guild.
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

// CommandHelp handles the "help" command.
func CommandHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	logger.Log.Info("Received help command")
	// Create an embed for the response.
	embed := &discordgo.MessageEmbed{
		Title:       "Help",
		Description: "If you need help or have any issues, please message Susplayer32 on Discord.",
		Color:       0x00ff00,
	}

	// Respond to the interaction with the embed.
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  64, // Ephemeral response (only visible to the user who used the command)
		},
	})
}
