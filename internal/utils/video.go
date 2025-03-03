package utils

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
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

// Will return the video duration in minutes rounded up
func GetVideoDurationInMinutes(filePath string) (uint, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("error running ffprobe: %w", err)
	}

	durationStr := strings.TrimSpace(string(output))
	durationSec, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing duration: %w", err)
	}

	durationMin := uint(math.Ceil(durationSec / 60.0))

	return durationMin, nil
}
