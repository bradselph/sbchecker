package services

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"sbchecker/internal/database"
	"sbchecker/internal/logger"
	"sbchecker/models"
)

// SendDailyUpdate sends a daily update to the Discord channel associated with the given account.
func SendDailyUpdate(account models.Account, discord *discordgo.Session) {
	// Create a new Discord message embed with the account's title, last status,
	// and a color based on the ban status.
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("24 Hour Update - %s", account.Title),
		Description: fmt.Sprintf("The last status of account named %s was %s. Your account is still being monitored.", account.Title, account.LastStatus),
		Color:       GetColorForBanStatus(account.LastStatus),
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	// Send the embed to the account's Discord channel.
	_, err := discord.ChannelMessageSendComplex(account.ChannelID, &discordgo.MessageSend{
		Embed: embed,
	})
	if err != nil {
		logger.Log.WithError(err).Error("Failed to send daily update message for account named", account.Title)
	}

	// Update the account's LastCheck and LastNotification fields in the database.
	account.LastCheck = time.Now().Unix()        // set the LastCheck to current time.
	account.LastNotification = time.Now().Unix() // set the LastNotification to current time.
	if err := database.DB.Save(&account).Error; err != nil {
		logger.Log.WithError(err).Error("Failed to save account changes for account named", account.Title)
	}
}

// CheckAccounts periodically checks all accounts in the database and sends
// notifications if necessary. It also updates the LastCheck and LastNotification
// fields for each account.
func CheckAccounts(s *discordgo.Session) {
	for {
		logger.Log.Info("Starting periodic account check")

		// Fetch all accounts from the database.
		var accounts []models.Account
		if err := database.DB.Find(&accounts).Error; err != nil {
			logger.Log.WithError(err).Error("Failed to fetch accounts from the database")
			continue
		}

		// Iterate through each account and perform checks.
		for _, account := range accounts {
			// Get the last check and notification times for the account.
			var lastCheck time.Time
			if account.LastCheck != 0 {
				lastCheck = time.Unix(account.LastCheck, 0)
			}
			var lastNotification time.Time
			if account.LastNotification != 0 {
				lastNotification = time.Unix(account.LastNotification, 0)
			}

			// Handle accounts with expired cookies.
			if account.IsExpiredCookie {
				logger.Log.WithField("account", account.Title).Info("Skipping account with expired cookie")
				if time.Since(lastNotification).Hours() > 24 {
					go SendDailyUpdate(account, s)
				} else {
					logger.Log.WithField("account", account.Title).Info("Owner of account named", account.Title, "recently notified within 24 hours, skipping")
				}
				continue
			}

			// Check the account if it hasn't been checked in the last 15 minutes.
			if time.Since(lastCheck).Minutes() > 15 {
				go CheckSingleAccount(account, s)
			} else {
				logger.Log.WithField("account", account.Title).Info("Account named", account.Title, "checked recently, skipping")
			}

			// Send a daily update if the account hasn't been notified in the last 24 hours.
			if time.Since(lastNotification).Hours() > 24 {
				go SendDailyUpdate(account, s)
			} else {
				logger.Log.WithField("account", account.Title).Info("Owner of Account Named", account.Title, "recently notified within 24Hours already, skipping")
			}
		}

		// Wait for 1 minute before checking again.
		time.Sleep(1 * time.Minute)
	}
}

// CheckSingleAccount checks the status of a single account and sends a notification
// if the status has changed. It also updates the account's LastCheck and IsExpiredCookie
// fields in the database.
func CheckSingleAccount(account models.Account, discord *discordgo.Session) {
	// Check the account's status.
	result, err := checkAccount(account.SSOCookie)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to check account named ", account.Title, "possible expired SSO Cookie")
		return
	}

	// If the account has an invalid cookie, send a notification and update the
	// account's LastCookieNotification and IsExpiredCookie fields in the database.
	if result == models.StatusInvalidCookie {
		cooldownDuration := 6 * time.Hour // Consider moving this to a constant or configuration variable
		lastNotification := time.Unix(account.LastCookieNotification, 0)
		if time.Since(lastNotification) >= cooldownDuration || account.LastCookieNotification == 0 {
			logger.Log.Infof("Account named %s has an invalid SSO cookie", account.Title)
			embed := &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("%s - Invalid SSO Cookie", account.Title),
				Description: fmt.Sprintf("The SSO cookie for account %s has expired. Please update the cookie using the /updateaccount command or delete the account using the /deleteaccount command.", account.Title),
				Color:       0xff0000,
				Timestamp:   time.Now().Format(time.RFC3339),
			}
			_, err = discord.ChannelMessageSendComplex(account.ChannelID, &discordgo.MessageSend{
				Embed: embed,
			})
			if err != nil {
				logger.Log.WithError(err).Error("Failed to send invalid cookie notification for account named", account.Title)
			}

			account.LastCookieNotification = time.Now().Unix() // Store the current time as the last notification time
			account.IsExpiredCookie = true                     // Mark the account as having an expired cookie
			if err := database.DB.Save(&account).Error; err != nil {
				logger.Log.WithError(err).Error("Failed to save account changes for account named", account.Title)
			}
		} else {
			logger.Log.Infof("Skipping expired cookie notification for account named %s (cooldown)", account.Title)
		}
		return
	}

	lastStatus := account.LastStatus

	account.LastCheck = time.Now().Unix()
	account.IsExpiredCookie = false // Reset the expired cookie status if the account is successfully checked
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

// GetColorForBanStatus returns a color code based on the ban status.
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

// EmbedTitleFromStatus returns a string title based on the ban status.
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
