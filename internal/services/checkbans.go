package services

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/silenta-salmans/sbchecker/internal/database"
	"github.com/silenta-salmans/sbchecker/internal/logger"
	"github.com/silenta-salmans/sbchecker/models"
)

func CheckAccounts(s *discordgo.Session) {
	for {
		logger.Log.Info("Starting periodic account check")

		var accounts []models.Account
		database.DB.Find(&accounts)

		for _, account := range accounts {
			var lastCheck time.Time
			if account.LastCheck != 0 {
				lastCheck = time.Unix(account.LastCheck, 0)
			}

			if time.Since(lastCheck).Minutes() > 15 {
				go CheckSingleAccount(account, s)
			} else {
				logger.Log.WithField("account", account.Title).Info("Account checked recently, skipping")
			}
		}

		time.Sleep(1 * time.Minute)
	}
}

func CheckSingleAccount(account models.Account, discord *discordgo.Session) {
	result, err := CheckAccount(account.SSOCookie)
	if err != nil {
		logger.Log.WithError(err).Error("Error checking account")
		return
	}

	lastStatus := account.LastStatus
	account.LastCheck = time.Now().Unix()

	err = database.DB.Save(&account).Error
	if err != nil {
		logger.Log.WithError(err).Error("Error saving account")
		return
	}

	if result == lastStatus {
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("%s - %s", account.Title, EmbedTitleFromStatus(result)),
			Description: fmt.Sprintf("The status of account %s has not changed", account.Title),
			Color:       GetColorForBanStatus(result),
			Timestamp:   time.Now().Format(time.RFC3339),
		}

		_, err = discord.ChannelMessageSendComplex(account.ChannelID, &discordgo.MessageSend{
			Embed: embed,
		})
		if err != nil {
			logger.Log.WithError(err).Error("Error sending message")
		}

		return
	}

	account.LastStatus = result

	err = database.DB.Save(&account).Error
	if err != nil {
		logger.Log.WithError(err).Error("Error saving account")
		return
	}

	logger.Log.Infof("Account %s status changed to %s", account.Title, result)

	ban := models.Ban{
		Account:   account,
		Status:    result,
		AccountID: account.ID,
	}

	err = database.DB.Create(&ban).Error
	if err != nil {
		logger.Log.WithError(err).Error("Error creating ban")
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s - %s", account.Title, EmbedTitleFromStatus(result)),
		Description: fmt.Sprintf("The status of account %s has changed to %s <@%s>", account.Title, result, account.UserID),
		Color:       GetColorForBanStatus(result),
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	_, err = discord.ChannelMessageSendComplex(account.ChannelID, &discordgo.MessageSend{
		Embed:   embed,
		Content: fmt.Sprintf("<@%s>", account.UserID),
	})
	if err != nil {
		logger.Log.WithError(err).Error("Error sending message")
	}
}

func GetColorForBanStatus(status models.Status) int {
	switch status {
	case models.StatusPermaban:
		return 0xff0000
	case models.StatusShadowban:
		return 0xffff00
	default:
		return 0x00ff00
	}
}

func EmbedTitleFromStatus(status models.Status) string {
	switch status {
	case models.StatusPermaban:
		return "PERMANENT BAN DETECTED"
	case models.StatusShadowban:
		return "SHADOWBAN DETECTED"
	default:
		return "ACCOUNT NOT BANNED"
	}
}
