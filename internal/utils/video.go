package utils

import (
	"strings"
)

func FileIsMPD(fileName string) bool {
	return strings.HasSuffix(fileName, ".mpd")
}
