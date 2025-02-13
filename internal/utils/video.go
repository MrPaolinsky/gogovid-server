package utils

import (
	"bytes"
	"os"
	"strings"
)

func FileIsMPD(fileName string) bool {
	return strings.HasSuffix(fileName, ".mpd")
}

func FileIsMP4(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read the first 8 bytes to check the magic number
	header := make([]byte, 8)
	_, err = file.Read(header)
	if err != nil {
		return false, err
	}

	// MP4 files start with ftyp at byte 4
	if bytes.Equal(header[4:8], []byte("ftyp")) {
		return true, nil
	}

	return false, nil
}
