package command

import (
	"codstatusbot2.0/command/accountage"
	"codstatusbot2.0/command/accountlogs"
	"codstatusbot2.0/command/addaccount"
	"codstatusbot2.0/command/help"
	"codstatusbot2.0/command/removeaccount"
	"codstatusbot2.0/command/setpreference"
	"codstatusbot2.0/command/updateaccount"
	"codstatusbot2.0/logger"

	"github.com/bwmarrin/discordgo"
)

var Handlers = map[string]func(*discordgo.Session, *discordgo.InteractionCreate){}

func RegisterCommands(s *discordgo.Session, guildID string) {
	logger.Log.Info("Registering commands by command handler")

	removeaccount.RegisterCommand(s, guildID)
	Handlers["removeaccount"] = removeaccount.CommandRemoveAccount
	logger.Log.Info("Registering removeaccount command")

	accountlogs.RegisterCommand(s, guildID)
	Handlers["accountlogs"] = accountlogs.CommandAccountLogs
	logger.Log.Info("Registering accountlogs command")

	updateaccount.RegisterCommand(s, guildID)
	Handlers["updateaccount"] = updateaccount.CommandUpdateAccount
	logger.Log.Info("Registering updateaccount command")

	setpreference.RegisterCommand(s, guildID)
	Handlers["setpreference"] = setpreference.CommandSetPreference
	logger.Log.Info("Registering setpreference command")

	accountage.RegisterCommand(s, guildID)
	Handlers["accountage"] = accountage.CommandAccountAge
	logger.Log.Info("Registering accountage command")

	addaccount.RegisterCommand(s, guildID)
	Handlers["addaccount"] = addaccount.CommandAddAccount
	logger.Log.Info("Registering addaccount command")

	help.RegisterCommand(s, guildID)
	Handlers["help"] = help.CommandHelp
	logger.Log.Info("Registering help command")
}

func UnregisterCommands(s *discordgo.Session, guildID string) {
	logger.Log.Info("Unregistering commands by command handler")

	addaccount.UnregisterCommand(s, guildID)
	logger.Log.Info("Unregistering addaccount command")

	removeaccount.UnregisterCommand(s, guildID)
	logger.Log.Info("Unregistering removeaccount command")

	accountlogs.UnregisterCommand(s, guildID)
	logger.Log.Info("Unregistering accountlogs command")

	setpreference.UnregisterCommand(s, guildID)
	logger.Log.Info("Unregistering setpreference command")

	updateaccount.UnregisterCommand(s, guildID)
	logger.Log.Info("Unregistering updateaccount command")

	accountage.UnregisterCommand(s, guildID)
	logger.Log.Info("Unregistering accountage command")

	help.UnregisterCommand(s, guildID)
	logger.Log.Info("Unregistering help command")

}
