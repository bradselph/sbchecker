package commands

import (
	"github.com/bwmarrin/discordgo"
)

func RegisterCommandHandlers() {
	commandHandlers["addaccount"] = addaccount.CommandAddAccount
	commandHandlers["removeaccount"] = removeaccount.CommandRemoveAccount
	commandHandlers["accountlogs"] = accountlogs.CommandAccountLogs
	commandHandlers["updateaccount"] = updateaccount.CommandUpdateAccount
	commandHandlers["accountage"] = accountage.CommandAccountAge
	commandHandlers["help"] = help.CommandHelp
}

// RegisterCommands registers all command handlers for a specific guild.
func RegisterCommands(s *discordgo.Session, guildID string) {
	addaccount.RegisterCommand(s, guildID)
	removeaccount.RegisterCommand(s, guildID)
	accountlogs.RegisterCommand(s, guildID)
	updateaccount.RegisterCommand(s, guildID)
	accountage.RegisterCommand(s, guildID)
	help.RegisterCommand(s, guildID)
}

// UnregisterCommands unregisters all command handlers for a specific guild.
func UnregisterCommands(s *discordgo.Session, guildID string) {
	addaccount.UnregisterCommand(s, guildID)
	removeaccount.UnregisterCommand(s, guildID)
	accountlogs.UnregisterCommand(s, guildID)
	updateaccount.UnregisterCommand(s, guildID)
	accountage.UnregisterCommand(s, guildID)
	help.UnregisterCommand(s, guildID)
}
