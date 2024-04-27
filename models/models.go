package models

import (
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	GuildID          string `gorm:"index"` // GuildID represents the ID of the guild the account belongs to.
	UserID           string `gorm:"index"` // UserID represents the ID of the user.
	ChannelID        string // ChannelID represents the ID of the channel associated with the account.
	Title            string // Title represents the title of the account.
	LastStatus       Status `gorm:"default:unknown"` // LastStatus represents the last known status of the account.
	LastCheck        int64  `gorm:"default:0"`       // LastCheck represents the timestamp of the last check performed on the account.
	LastNotification int64  // LastNotification represents the timestamp of the last notification sent out on the account.
	SSOCookie        string // SSOCookie represents the SSO cookie associated with the account.
	Created          string // Created represents the timestamp of when the account was created.

}

type Status string

const (
	StatusGood          Status = "good"           // StatusGood indicates that the account is in good standing.
	StatusPermaban      Status = "permaban"       // StatusPermaban indicates that the account has been permanently banned.
	StatusShadowban     Status = "shadowban"      // StatusShadowban indicates that the account has been shadowbanned.
	StatusUnknown       Status = "unknown"        // StatusUnknown indicates that the status of the account is unknown.
	StatusInvalidCookie Status = "invalid_cookie" // StatusInvalidCookie indicates that the account has an invalid SSO cookie.

)

type Ban struct {
	gorm.Model
	Account   Account // Account represents the account that has been banned.
	AccountID uint    // AccountID represents the ID of the banned account.
	Status    Status  // Status represents the status of the ban.

}
