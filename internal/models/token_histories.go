package models

import "gorm.io/gorm"

type TokenAction string

const (
	CREATE  TokenAction = "CREATE"
	DELETE  TokenAction = "DELETE"
	CONSUME TokenAction = "CONSUME"
)

type TokenHistory struct {
	gorm.Model
	TokenId uint
	Action  TokenAction
	Target  string
}
