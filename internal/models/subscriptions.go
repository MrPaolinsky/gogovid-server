package models

import "gorm.io/gorm"

type SubscriptionType string

const (
	TRIAL SubscriptionType = "free"
	PAID  SubscriptionType = "paid"
)

type Subscription struct {
	gorm.Model
	UserId           uint
	SubscriptionType SubscriptionType
	UsedTokens       uint
	MaxTokens        uint
	UsedStorageMB    uint
	MaxStorageMB     uint
	UsedMinutes      uint64
	MaxMinutes       uint64
}
