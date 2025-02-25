package models

import (
	"time"

	"gorm.io/gorm"
)

type DRMLabel string

const (
	AUDIO DRMLabel = "AUDIO"
	R480  DRMLabel = "R480"
	R720  DRMLabel = "R720"
	R1080 DRMLabel = "R1080"
)

var Labels []DRMLabel = []DRMLabel{
	AUDIO,
	R480,
	R720,
	R1080,
}

type DRMKey struct {
	gorm.Model
	VideoId uint
	DRMInfo DRMInfo `gorm:"embedded"`
}

type DRMInfo struct {
	KeyID string `gorm:"unique"`
	Key   []byte
	Label DRMLabel
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
