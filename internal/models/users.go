package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name         string
	Email        string
	Videos       []Video
	ApiTokens    []ApiToken
	Subscription Subscription
}
