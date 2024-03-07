package models

import (
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model

	GuildID string `gorm:"index"`

	UserID     string `gorm:"index"`
	ChannelID  string
	Title      string
	LastStatus Status `gorm:"default:unknown"`
	LastCheck  int64  `gorm:"default:0"` // every 15 minutes
	SSOCookie  string
}

type Status string

const (
	StatusGood          Status = "good"
	StatusPermaban      Status = "permaban"
	StatusShadowban     Status = "shadowban"
	StatusUnknown       Status = "unknown"
	StatusInvalidCookie Status = "invalid_cookie"
)

type Ban struct {
	gorm.Model
	Account   Account
	AccountID uint
	Status    Status
}
