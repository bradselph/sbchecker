package main
import (
	"fmt"
	"log"
	"os"
	"os/signal"
	bot "sbchecker/cmd/dcbot"
	"sbchecker/internal/database"
	"sbchecker/internal/logger"
	"sbchecker/internal/services"
	"syscall"
	
	"github.com/bwmarrin/discordgo"
)
func main() {
	log.Println("Initializing logger")
	logger.Initialize()
	logger.Log.Info("Initializing database connection")
	err := database.Initialize()
	if err != nil {
		logger.Log.WithError(err).Error("Error initializing database")
	}
	instance, err := bot.RunBot()
	if err != nil {
		logger.Log.WithError(err).Error("Error running bot")
	}
	logger.Log.Info("Bot is running")
	instance.AddHandler(onGuildCreate)
	go services.CheckAccounts(instance)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	guildID := event.Guild.ID
	fmt.Println("Bot joined server:", guildID)
	registerCommands(s, guildID)
	// restartBot()  // Uncomment this line if you want to restart the bot with new guild ad
}
func registerCommands(s *discordgo.Session, guildID string) {
	fmt.Println("Registering commands for server:", guildID)
}
