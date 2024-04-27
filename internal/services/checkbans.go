package services

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"sbchecker/internal/database"
	"sbchecker/internal/logger"
	"sbchecker/models"
)

func SendDailyUpdate(account models.Account, discord *discordgo.Session) {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("24 Hour Update - %s", account.Title),
		Description: fmt.Sprintf("The last status of account named %s was %s. Your account is still being monitored.", account.Title, account.LastStatus),
		Color:       GetColorForBanStatus(account.LastStatus),
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	_, err := discord.ChannelMessageSendComplex(account.ChannelID, &discordgo.MessageSend{
		Embed: embed,
	})
	if err != nil {
		logger.Log.WithError(err).Error("Failed to send daily update message for account named", account.Title)
	}
	account.LastCheck = time.Now().Unix()        // set the LastCheck to current time.
	account.LastNotification = time.Now().Unix() // set the LastNotification to current time.
	if err := database.DB.Save(&account).Error; err != nil {
		logger.Log.WithError(err).Error("Failed to save account changes for account named", account.Title)
	}
}

func CheckAccounts(s *discordgo.Session) {
	for {
		logger.Log.Info("Starting periodic account check")

		var accounts []models.Account
		if err := database.DB.Find(&accounts).Error; err != nil {
			logger.Log.WithError(err).Error("Failed to fetch accounts from the database")
			continue
		}

		for _, account := range accounts {
			var lastCheck time.Time
			if account.LastCheck != 0 {
				lastCheck = time.Unix(account.LastCheck, 0)
			}
			var lastNotification time.Time
			if account.LastNotification != 0 {
				lastNotification = time.Unix(account.LastNotification, 0)
			}
			if time.Since(lastCheck).Minutes() > 15 {
				go CheckSingleAccount(account, s)
			} else {
				logger.Log.WithField("account", account.Title).Info("Account named", account.Title, "checked recently, skipping")
			}
			if time.Since(lastNotification).Hours() > 24 {
				go SendDailyUpdate(account, s)
			} else {
				logger.Log.WithField("account", account.Title).Info("Owner of Account Named", account.Title, "recently notified within 24Hours already, skipping")
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func CheckSingleAccount(account models.Account, discord *discordgo.Session) {
	result, err := CheckAccount(account.SSOCookie)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to check account named", account.Title)
		return
	}
	lastStatus := account.LastStatus
	account.LastCheck = time.Now().Unix()
	if err := database.DB.Save(&account).Error; err != nil {
		logger.Log.WithError(err).Error("Failed to save account changes for account named", account.Title)
		return
	}
	if result != lastStatus {
		account.LastStatus = result
		if err := database.DB.Save(&account).Error; err != nil {
			logger.Log.WithError(err).Error("Failed to save account changes for account named", account.Title)
			return
		}
		logger.Log.Infof("Account named %s status changed to %s", account.Title, result)
		ban := models.Ban{
			Account:   account,
			Status:    result,
			AccountID: account.ID,
		}
		if err := database.DB.Create(&ban).Error; err != nil {
			logger.Log.WithError(err).Error("Failed to create new ban record for account named", account.Title)
		}
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("%s - %s", account.Title, EmbedTitleFromStatus(result)),
			Description: fmt.Sprintf("The status of account named %s has changed to %s <@%s>", account.Title, result, account.UserID),
			Color:       GetColorForBanStatus(result),
			Timestamp:   time.Now().Format(time.RFC3339),
		}
		_, err = discord.ChannelMessageSendComplex(account.ChannelID, &discordgo.MessageSend{
			Embed:   embed,
			Content: fmt.Sprintf("<@%s>", account.UserID),
		})
		if err != nil {
			logger.Log.WithError(err).Error("Failed to send status update message for account named", account.Title)
		}
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
