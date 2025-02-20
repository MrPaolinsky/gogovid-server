package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateDRMKey() (keyID string, key []byte, err error) {
	// Generate 16-byte key ID
	keyIDBytes := make([]byte, 16)
	if _, err := rand.Read(keyIDBytes); err != nil {
		return "", nil, err
	}

	// Generate 16-byte content key
	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", nil, err
	}

	return hex.EncodeToString(keyIDBytes), keyBytes, nil
}

func FormatKeyToHex(key []byte) string {
	return hex.EncodeToString(key)
}

func GenerateDRMKeys() {}
