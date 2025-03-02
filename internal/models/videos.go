package models

import "gorm.io/gorm"

type VideoQuality struct {
	ResolutionX uint16
	ResolutionY uint16
	Bitrate     uint16
}

type Video struct {
	gorm.Model
	UserId          uint
	Name            string
	DurationMinutes uint
    Formatted       bool
    Route   string
}
