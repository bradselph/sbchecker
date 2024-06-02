package command

import (
	"codstatusbot2.0/command/help"
	"codstatusbot2.0/logger"

	"github.com/bwmarrin/discordgo"
)

var CommandHandlers = map[string]func(*discordgo.Session, *discordgo.InteractionCreate){}

// var Commands = make(map[string]*discordgo.ApplicationCommand)

func RegisterCommands(s *discordgo.Session, guildID string) {
	logger.Log.Info("Registering commands by command handler")

	//	addaccount.RegisterCommand(s, guildID, command)
	//	CommandHandlers["addaccount"] = addaccount.CommandAddAccount
	//	logger.Log.Info("Registering addaccount command")

	//	removeaccount.RegisterCommand(s, guildID, command)
	//	CommandHandlers["removeaccount"] = removeaccount.CommandRemoveAccount
	//	logger.Log.Info("Registering removeaccount command")

	//	accountlogs.RegisterCommand(s, guildID, command)
	//	CommandHandlers["accountlogs"] = accountlogs.CommandAccountLogs
	//	logger.Log.Info("Registering accountlogs command")

	//	updateaccount.RegisterCommand(s, guildID, command)
	//	CommandHandlers["updateaccount"] = updateaccount.CommandUpdateAccount
	//	logger.Log.Info("Registering updateaccount command")

	//	accountage.RegisterCommand(s, guildID, command)
	//	CommandHandlers["accountage"] = accountage.CommandAccountAge
	//	logger.Log.Info("Registering accountage command")

	help.RegisterCommand(s, guildID, command)
	CommandHandlers["help"] = help.CommandHelp
	logger.Log.Info("Registering help command")
}

func UnregisterCommands(s *discordgo.Session, guildID string) {
	logger.Log.Info("Unregistering commands by command handler")

	//	addaccount.UnregisterCommand(s, guildID)
	//	logger.Log.Info("Unregistering addaccount command")

	//	removeaccount.UnregisterCommand(s, guildID)
	//	logger.Log.Info("Unregistering removeaccount command")

	//	accountlogs.UnregisterCommand(s, guildID)
	//	logger.Log.Info("Unregistering accountlogs command")

	//	updateaccount.UnregisterCommand(s, guildID)
	//	logger.Log.Info("Unregistering updateaccount command")

	//	accountage.UnregisterCommand(s, guildID)
	//	logger.Log.Info("Unregistering accountage command")

	help.UnregisterCommand(s, guildID)
	logger.Log.Info("Unregistering help command")

}