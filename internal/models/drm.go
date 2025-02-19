package models

import (
	"time"

	"gorm.io/gorm"
)

type DRMKey struct {
	gorm.Model
	VideoId uint   `gorm:"unique"`
	KeyId   string `gorm:"unique"`
	Value   []byte // The actual encryption key
}

type LicenseRequest struct {
	KeyID    string `json:"key_id"`    // Which key is being requested
	VideoID  uint   `json:"video_id"`  // Which video
	DeviceID string `json:"device_id"` // Which device
	UserID   string `json:"user_id"`   // Who is requesting
}

type License struct {
	ID        string    `json:"id"`
	KeyID     string    `json:"key_id"`
	Key       []byte    `json:"key"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}
