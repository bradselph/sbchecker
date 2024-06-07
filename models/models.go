package models

import (
	"gorm.io/gorm"
)

// Account represents a user's account within a guild.
type Account struct {
	gorm.Model
	GuildID                string `gorm:"index"` // The ID of the guild the account belongs to.
	UserID                 string `gorm:"index"` // The ID of the user.
	ChannelID              string // The ID of the channel associated with the account.
	Title                  string // The title of the account.
	LastStatus             Status `gorm:"default:unknown"` // The last known status of the account.
	LastCheck              int64  `gorm:"default:0"`       // The timestamp of the last check performed on the account.
	LastNotification       int64  // The timestamp of the last daily notification sent out on the account.
	LastCookieNotification int64  // The timestamp of the last notification sent out on the account for an expired ssocookie.
	SSOCookie              string // The SSO cookie associated with the account.
	Created                string // The timestamp of when the account was created on activision.
	IsExpiredCookie        bool   `gorm:"default:false"` // A flag indicating if the SSO cookie has expired.
}

// Status represents the status of an account.
type Status string

const (
	StatusGood          Status = "good"           // The account is in good standing.
	StatusPermaban      Status = "permaban"       // The account has been permanently banned.
	StatusShadowban     Status = "shadowban"      // The account has been shadowbanned.
	StatusUnknown       Status = "unknown"        // The status of the account is unknown.
	StatusInvalidCookie Status = "invalid_cookie" // The account has an invalid SSO cookie.
)

// Ban represents a ban on an account.
type Ban struct {
	gorm.Model
	Account   Account // The account that has been banned.
	AccountID uint    // The ID of the banned account.
	Status    Status  // The status of the ban.
}
