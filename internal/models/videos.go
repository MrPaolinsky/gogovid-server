package models

import "gorm.io/gorm"

type VideoQuality struct {
	ResolutionX uint16
	ResolutionY uint16
	Bitrate     uint16
}

type DRMInfo struct {
	KeyID   string
	Key     string
	KID     string
	PSSHBox string
}

type Video struct {
	gorm.Model
	UserId          uint
	Name            string
	DurationMinutes uint
	DRMInfo         *DRMInfo `gorm:"embedded"`
}
