package services

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"codstatusbot2.0/database"
	"codstatusbot2.0/logger"
	"codstatusbot2.0/models"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var notificationInterval string

func init() {
	err := godotenv.Load()
	if err != nil {
		logger.Log.WithError(err).Error("Failed to load .env file")
	}
	notificationInterval = os.Getenv("NOTIFICATION_INTERVAL")
}

func sendDailyUpdate(account models.Account, discord *discordgo.Session) {
	logger.Log.Infof("Sending daily update for account %s", account.Title)
	var description string
	if account.IsExpiredCookie {
		description = fmt.Sprintf("The SSO cookie for account %s has expired. Please update the cookie using the /updateaccount command or delete the account using the /removeaccount command.", account.Title)
	} else {
		description = fmt.Sprintf("The last status of account %s was %s.", account.Title, account.LastStatus)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Hour Update - %s", notificationInterval, account.Title),
		Description: description,
		Color:       GetColorForStatus(account.LastStatus, account.IsExpiredCookie),
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	var channelID string
	if account.InteractionType == "dm" {
		channel, err := discord.UserChannelCreate(account.UserID)
		if err != nil {
			logger.Log.WithError(err).Error("Failed to create DM channel")
			return
		}
		channelID = channel.ID
	} else {
		channelID = account.ChannelID
	}

	_, err := discord.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embed: embed,
	})
	if err != nil {
		logger.Log.WithError(err).Error("Failed to send scheduled update message for account ", account.Title)

	}

	account.LastCheck = time.Now().Unix()
	account.LastNotification = time.Now().Unix()
	if err := database.DB.Save(&account).Error; err != nil {
		logger.Log.WithError(err).Error("Failed to save account changes for account ", account.Title)

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

			if account.IsExpiredCookie {
				logger.Log.WithField("account", account.Title).Info("Skipping account with expired cookie")
				notificationInterval, _ := strconv.ParseFloat(os.Getenv("NOTIFICATION_INTERVAL"), 64)
				if time.Since(lastNotification).Hours() > notificationInterval {
					go sendDailyUpdate(account, s)
				} else {

					logger.Log.WithField("account", account.Title).Info("Owner of ", account.Title, "recently notified within ", notificationInterval, "Hours already, skipping")

				}
				continue
			}

			checkInterval, _ := strconv.ParseFloat(os.Getenv("CHECK_INTERVAL"), 64)
			if time.Since(lastCheck).Minutes() > checkInterval {
				go CheckSingleAccount(account, s)
			} else {

				logger.Log.WithField("account", account.Title).Info("Account ", account.Title, "checked recently less than ", checkInterval, "ago, skipping")

			}

			notificationInterval, _ := strconv.ParseFloat(os.Getenv("NOTIFICATION_INTERVAL"), 64)
			if time.Since(lastNotification).Hours() > notificationInterval {
				go sendDailyUpdate(account, s)
			} else {

				logger.Log.WithField("account", account.Title).Info("Owner of ", account.Title, "recently notified within ", notificationInterval, "Hours already, skipping")

			}
		}

		sleepDuration, _ := strconv.Atoi(os.Getenv("SLEEP_DURATION"))
		time.Sleep(time.Duration(sleepDuration) * time.Minute)
	}
}

func CheckSingleAccount(account models.Account, discord *discordgo.Session) {
	result, err := CheckAccount(account.SSOCookie)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to check account ", account.Title, "possible expired SSO Cookie")
		return
	}

	if result == models.StatusInvalidCookie {
		cooldownDuration, _ := strconv.ParseFloat(os.Getenv("COOLDOWN_DURATION"), 64)
		lastNotification := time.Unix(account.LastCookieNotification, 0)
		if time.Since(lastNotification) >= time.Duration(cooldownDuration)*time.Hour || account.LastCookieNotification == 0 {
			logger.Log.Infof("Account %s has an invalid SSO cookie", account.Title)
			embed := &discordgo.MessageEmbed{

				Title:       fmt.Sprintf("%s - Invalid SSO Cookie ", account.Title),
				Description: fmt.Sprintf("The SSO cookie for account %s has expired. Please update the cookie using the /updateaccount command or delete the account using the /deleteaccount command.", account.Title),
				Color:       0xff0000,
				Timestamp:   time.Now().Format(time.RFC3339),
			}

			var channelID string
			if account.InteractionType == "dm" {
				channel, err := discord.UserChannelCreate(account.UserID)
				if err != nil {
					logger.Log.WithError(err).Error("Failed to create DM channel")
					return
				}
				channelID = channel.ID
			} else {
				channelID = account.ChannelID
			}

			_, err = discord.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
				Embed: embed,
			})
			if err != nil {

				logger.Log.WithError(err).Error("Failed to send invalid cookie notification for account ", account.Title)

			}

			account.LastCookieNotification = time.Now().Unix()
			account.IsExpiredCookie = true
			if err := database.DB.Save(&account).Error; err != nil {

				logger.Log.WithError(err).Error("Failed to save account changes for account ", account.Title)
			}
		} else {
			logger.Log.Infof("Skipping expired cookie notification for account %s (cooldown) ", account.Title)

		}
		return
	}

	lastStatus := account.LastStatus
	account.LastCheck = time.Now().Unix()
	account.IsExpiredCookie = false
	if err := database.DB.Save(&account).Error; err != nil {

		logger.Log.WithError(err).Error("Failed to save account changes for account ", account.Title)

		return
	}
	if result != lastStatus {
		account.LastStatus = result
		if err := database.DB.Save(&account).Error; err != nil {

			logger.Log.WithError(err).Error("Failed to save account changes for account ", account.Title)

			return
		}
		logger.Log.Infof("Account %s status changed to %s", account.Title, result)
		ban := models.Ban{
			Account:   account,
			Status:    result,
			AccountID: account.ID,
		}
		if err := database.DB.Create(&ban).Error; err != nil {

			logger.Log.WithError(err).Error("Failed to create new ban record for account ", account.Title)
		}
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("%s - %s", account.Title, EmbedTitleFromStatus(result)),
			Description: fmt.Sprintf("The status of account %s has changed to %s <@%s>", account.Title, result, account.UserID),
			Color:       GetColorForStatus(result, account.IsExpiredCookie),
			Timestamp:   time.Now().Format(time.RFC3339),
		}

		var channelID string
		if account.InteractionType == "dm" {
			channel, err := discord.UserChannelCreate(account.UserID)
			if err != nil {
				logger.Log.WithError(err).Error("Failed to create DM channel")
				return
			}
			channelID = channel.ID
		} else {
			channelID = account.ChannelID
		}

		_, err = discord.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
			Embed:   embed,
			Content: fmt.Sprintf("<@%s>", account.UserID),
		})
		if err != nil {

			logger.Log.WithError(err).Error("Failed to send status update message for account ", account.Title)

		}
	}
}

func GetColorForStatus(status models.Status, isExpiredCookie bool) int {
	if isExpiredCookie {
		return 0xff0000
	}
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
