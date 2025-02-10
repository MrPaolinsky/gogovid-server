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
	UsedApiKeys      uint
	MaxApiKeys       uint
	UsedStorageMB    uint
	MaxStorageMB     uint
	UsedMinutes      uint64
	MaxMinutes       uint64
}

func DefaultSubscription() *Subscription {
	return &Subscription{
		SubscriptionType: TRIAL,
		MaxApiKeys:       1,
		MaxStorageMB:     100,
		MaxMinutes:       100,
	}
}
