package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"codstatusbot/cmd/accountage"
	"codstatusbot/cmd/accountlogs"
	"codstatusbot/cmd/addaccount"
	"codstatusbot/cmd/help"
	"codstatusbot/cmd/removeaccount"
	"codstatusbot/cmd/updateaccount"
	"codstatusbot/internal/logger"
	"codstatusbot/internal/services"
	"codstatusbot/models"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB              *gorm.DB
	dc              *discordgo.Session
	commandHandlers = map[string]func(*discordgo.Session, *discordgo.InteractionCreate){}
	instance        *discordgo.Session
)

func main() {
	logger.Log.Info("Bot starting...")
	err := loadEnvironmentVariables()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Environment Variables").Error()
	}
	err = databaselogin()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Databaselogin").Error()
	}
	instance, err = discordlogin()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Discordlogin").Error()
	}
	logger.Log.Info("Bot is running")
	instance.AddHandler(onGuildCreate)
	go services.CheckAccounts(instance)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	err = StopBot()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Shutdown", "Shutdown Process").Error()
	}
}

func loadEnvironmentVariables() error {
	err := godotenv.Load()
	if err != nil {
		logger.Log.WithError(err).Error("Error loading .env file")
		return err
	}
	return nil
}


		logger.Log.Info("Bot is running")
		instance.AddHandler(onGuildCreate)
		go services.CheckAccounts(instance)
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc
		StopBot()
	}


func databaselogin() error {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbVar := os.Getenv("DB_VAR")

	if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" || dbVar == "" {
		err = errors.New("one or more environment variables for database not set or missing")
		logger.Log.WithError(err).WithField("Bot Startup", "database variables").Error()
		return err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", dbUser, dbPassword, dbHost, dbPort, dbName, dbVar)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Mysql Config").Error()
		return err
	}

	DB = db

	err = DB.AutoMigrate(&models.Account{}, &models.Ban{})
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Startup", "Database Models Problem").Error()
		return err
	}

	return nil
}

func discordlogin() error {
	envToken := os.Getenv("DISCORD_TOKEN")
	if envToken == "" {
		err = errors.New("DISCORD_TOKEN environment variable not set")
		logger.Log.WithError(err).WithField("env", "DISCORD_TOKEN").Error()
		return err
	}

	dc, err = discordgo.New("Bot " + envToken)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Password").Error()
		return err
	}

	return nil
}

func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	guildID := event.Guild.ID
	fmt.Println("Bot joined server:", guildID)
	registerCommands(s, guildID)
}

func StartBot() (*discordgo.Session, error) {
	err := dc.Open()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Login").Error()
		return nil, err
	}

	err = dc.UpdateWatchStatus(0, "the Status of your Accounts so you dont have to.")
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Setting Presence Status").Error()
		return nil, err
	}

	guilds, err := dc.UserGuilds(100, "", "", false)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot startup", "Initiating Guilds").Error()
		return nil, err
	}

	for _, guild := range guilds {
		logger.Log.WithField("guild", guild.Name).Info("Connected to guild")
		registerCommands(dc, guild.ID)
	}

	dc.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handler, ok := commandHandlers[i.ApplicationCommandData().Name]
		if ok {
			handler(s, i)
		}
	})

	return dc, nil
}

func StopBot() error {
	err := dc.Close()
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Shutdown", "Shutdown Process").Error()
		return err
	}

	guilds, err := dc.UserGuilds(100, "", "", false)
	if err != nil {
		logger.Log.WithError(err).WithField("Bot Shutdown", "Disconnecting Guilds").Error()
		return err
	}

	for _, guild := range guilds {
		logger.Log.WithField("guild", guild.Name).Info("Disconnected from Guild")
		unregisterCommands(dc, guild.ID)

	}
	return nil
}

func RestartBot() error {
	err := StopBot()
	if err != nil {
		logger.Log.WithError(err).Error("Error stopping bot")
		return err
	}

	instance, err = StartBot()
	if err != nil {
		logger.Log.WithError(err).Error("Error starting bot")
		return err
	}

	instance.AddHandler(onGuildCreate)
	logger.Log.Info("Bot restarted successfully")
	return nil
}

func GetAllChoices(guildID string) []*discordgo.ApplicationCommandOptionChoice {
	var accounts []models.Account
	DB.Where("guild_id = ?", guildID).Find(&accounts)

	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(accounts))
	for i, account := range accounts {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  account.Title,
			Value: account.ID,
		}
	}

	return choices
}

func registerCommands(s *discordgo.Session, guildID string) {
	addaccount.RegisterCommand(s, guildID)
	commandHandlers["addaccount"] = addaccount.CommandAddAccount
	removeaccount.RegisterCommand(s, guildID)
	commandHandlers["removeaccount"] = removeaccount.CommandRemoveAccount
	accountlogs.RegisterCommand(s, guildID)
	commandHandlers["accountlogs"] = accountlogs.CommandAccountLogs
	updateaccount.RegisterCommand(s, guildID)
	commandHandlers["updateaccount"] = updateaccount.CommandUpdateAccount
	accountage.RegisterCommand(s, guildID)
	commandHandlers["accountage"] = accountage.CommandAccountAge
	help.RegisterCommand(s, guildID)
	commandHandlers["help"] = help.CommandHelp
}

func unregisterCommands(s *discordgo.Session, guildID string) {
	addaccount.UnregisterCommand(s, guildID)
	removeaccount.UnregisterCommand(s, guildID)
	accountlogs.UnregisterCommand(s, guildID)
	updateaccount.UnregisterCommand(s, guildID)
	accountage.UnregisterCommand(s, guildID)
	help.UnregisterCommand(s, guildID)
}
