package services

import (
	"crypto/rand"
	"encoding/hex"
	"go-streamer/internal/models"
	"go-streamer/internal/repositorioes/cruds"
	"time"
)

type DRMService struct {
	drmCrud *cruds.DRMCrud
}

func NewDRMService(drmCrud *cruds.DRMCrud) *DRMService {
	return &DRMService{drmCrud: drmCrud}
}

// Generate a new encryption key for a video
func (s *DRMService) GenerateKey(videoId uint) (*models.DRMKey, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	keyIDBytes := make([]byte, 16)
	if _, err := rand.Read(keyIDBytes); err != nil {
		return nil, err
	}

	keyID := hex.EncodeToString(keyIDBytes)

	drmKey := &models.DRMKey{
		VideoId: videoId,
		KeyId:   keyID,
		Value:   key,
	}

	if err := s.drmCrud.StoreKey(drmKey); err != nil {
		return nil, err
	}

	return drmKey, nil
}

// Generates license for playback
func (s *DRMService) GenerateLicense(req models.LicenseRequest) (*models.License, error) {
	drmKey, err := s.drmCrud.GetKey(req.KeyID)
	if err != nil {
		return nil, err
	}

	license := &models.License{
		KeyID:     drmKey.KeyId,
		Key:       drmKey.Value,
		StartDate: time.Now(),
		EndDate:   time.Now().Add(24 * time.Hour),
	}

	return license, nil
}
