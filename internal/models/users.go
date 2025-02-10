package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name           string
	Email          string `gorm:"unique"`
	ProviderUserId string `gorm:"unique"`
	Videos         []Video
	ApiTokens      []ApiToken
	Subscription   Subscription
}

// CRUD Models
type CreateUser struct {
	Name           string
	Email          string
	ProviderUserId string
}
