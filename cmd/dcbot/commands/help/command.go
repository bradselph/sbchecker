package help

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"sbchecker/internal/logger"
)

var username string

func init() {
	username = os.Getenv("HELP_USERNAME")
	if username == "your-username" {
		username = "No Name Has Been Set" // default value
	}
}

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
	// Create an embed for the response.
	embed := &discordgo.MessageEmbed{
		Title:       "Help",
		Description: fmt.Sprintf("If you need help or have any issues, please message %s on Discord.", username),
		Color:       0x00ff00,
	}

	// Respond to the interaction with the embed.
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}
