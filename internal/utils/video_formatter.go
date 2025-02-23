package utils

import (
	"bytes"
	"errors"
	"fmt"
	"go-streamer/internal/models"
	"log"
	"os"
	"os/exec"
	"time"
)

/*
ffmpeg -i t.mp4 -vf "scale=1920:1080" -c:v libx264 -b:v 7000k -c:a aac -b:a 128k t_1080p.mp4
ffmpeg -i t.mp4 -vf "scale=1280:720" -c:v libx264 -b:v 2500k -c:a aac -b:a 128k t_720p.mp4
ffmpeg -i t.mp4 -vf "scale=854:480" -c:v libx264 -b:v 1200k -c:a aac -b:a 128k t_480p.mp4

packager \
  in=audio.mp4,stream=audio,output=encrypted_audio.mp4,drm_label=AUDIO \
  in=480p.mp4,stream=video,output=encrypted_480p.mp4,drm_label=SD \
  in=720p.mp4,stream=video,output=encrypted_720p.mp4,drm_label=HD \
  in=1080p.mp4,stream=video,output=encrypted_1080p.mp4,drm_label=HD \
  --enable_raw_key_encryption \
  --keys label=AUDIO:key_id=1234567890abcdef1234567890abcdef:key=abcdef1234567890abcdef1234567890,label=SD:key_id=234567890abcdef1234567890abcdef1:key=bcdef1234567890abcdef1234567890a,label=HD:key_id=34567890abcdef1234567890abcdef12:key=cdef1234567890abcdef1234567890ab \
  --protection_systems Widevine,PlayReady \
  --mpd_output stream.mpd
*/

// Callback with directory where all the generated files are.
type FormattingCallback func(string)

var res1080 = models.VideoQuality{Bitrate: 7000, ResolutionX: 1920, ResolutionY: 1080}
var res720 = models.VideoQuality{Bitrate: 2500, ResolutionX: 1280, ResolutionY: 720}
var res480 = models.VideoQuality{Bitrate: 1200, ResolutionX: 854, ResolutionY: 480}

var qualities = [3]models.VideoQuality{res1080, res720, res480}

// Generate different qualities for video and then generates the mpd manifest and all its fragments,
// it deletes all the files once the excecution of the function is completed, pass a callback func
// to do something with the files
func ConvertAndFormatToFragmentedMP4(videoPath string, drmInfo []*models.DRMInfo, fn FormattingCallback) error {
	name := fmt.Sprintf("gogovid-%d", time.Now().UnixMilli())
	actionPath := "/tmp/" + name

	// Create working directory
	if err := os.MkdirAll(actionPath, 0755); err != nil {
		return err
	}

	err := generateVideoResolutions(videoPath, actionPath, name)
	if err != nil {
		log.Println("Error generation multiple resolutions for video")
		return err
	}

	err = generateEncryptedFragmentedMP4(videoPath, actionPath, name, drmInfo)

	if err != nil {
		log.Println("Error generatiing fragmented mp4")
		return err
	}

	fn(actionPath)

	return nil
}

func generateVideoResolutions(filePath string, actionPath string, name string) error {
	for i := range len(qualities) {
		cmd := exec.Command(
			"ffmpeg",
			"-i", fmt.Sprintf("%s", filePath),
			"-vf", fmt.Sprintf("scale=%d:%d", qualities[i].ResolutionX, qualities[i].ResolutionY),
			"-c:v", "libx264",
			"-b:v", fmt.Sprintf("%dk", qualities[i].Bitrate),
			"-c:a", "aac",
			"-b:a", "128k",
			"-threads", "1",
			fmt.Sprintf("%s-%d.mp4", name, qualities[i].ResolutionY),
		)
		cmd.Dir = actionPath
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func generateEncryptedFragmentedMP4(filePath, actionPath, name string, drmInfo []*models.DRMInfo) error {
	var packagerArgs []string

	for _, info := range drmInfo {
		if info.Label == models.AUDIO {
			packagerArgs = append(packagerArgs,
				fmt.Sprintf(
					"in=%s,stream=audio,segment_template=audio_$Number$.m4s,"+
						"init_segment=audio_init.m4s,"+
						"drm_label=%s",
					filePath, info.Label,
				),
			)
		} else {
			quality, err := drmInfoToQuality(info)
			if err != nil {
				return err
			}
			packagerArgs = append(packagerArgs,
				fmt.Sprintf(
					"in=%s-%d.mp4,stream=video,segment_template=%s-%d_$Number$.m4s,"+
						"init_segment=%s-%d_init.m4s,"+
						"drm_label=%s",
					name, quality.ResolutionY,
					name, quality.ResolutionY,
					name, quality.ResolutionY,
					info.Label,
				),
			)
		}
	}

	// Append packager global configurations and keys for each label
	packagerArgs = append(packagerArgs,
		"--generate_static_live_mpd",
		"--mpd_output", fmt.Sprintf("%s.mpd", name),
		"--segment_duration", "8",
		"--fragment_duration", "8",
		"--enable_raw_key_encryption",
		"--protection_scheme", "cenc",
		"--keys",
	)

	var keys string
	for i, info := range drmInfo {
		shouldAddComma := func() string {
			if i < len(drmInfo)-1 {
				return ","
			}
			return ""
		}

		keys += fmt.Sprintf(
			"label=%s:key_id=%s:key=%s%s",
			info.Label,
			info.KeyID,
			FormatKeyToHex(info.Key),
			shouldAddComma(),
		)
	}

	packagerArgs = append(packagerArgs,
		keys,
		"--protection_systems", "Widevine,PlayReady",
	)

	cmd := exec.Command("packager", packagerArgs...)
	cmd.Dir = actionPath

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		log.Println(packagerArgs)
		log.Printf("Packager output: %s\n %s\n", outb.String(), errb.String())
		return err
	}

	return nil
}

func drmInfoToQuality(drmInfo *models.DRMInfo) (*models.VideoQuality, error) {
	if drmInfo.Label == models.AUDIO {
		return nil, errors.New("Cant convert AUDIO label to resolution")
	} else {
		switch drmInfo.Label {
		case models.R1080:
			return &res1080, nil
		case models.R720:
			return &res720, nil
		case models.R480:
			return &res480, nil
		}
	}
	return nil, errors.New("No valid DRM info found")
}
