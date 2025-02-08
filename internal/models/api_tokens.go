package models

import "gorm.io/gorm"

type TokenType string

const (
	FREE    TokenType = "free"
	PREMIUM TokenType = "premium"
)

type ApiToken struct {
	gorm.Model
	UserId  uint
	Value   string
	Type    TokenType
	history []TokenHistory
}
