package utils

import (
	"bytes"
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
	  input=t_1080p.mp4,stream=video,segment_template=t_1080p_\$Number\$.m4s,init_segment=t_1080p_init.m4s \
	  input=t_720p.mp4,stream=video,segment_template=t_720p_\$Number\$.m4s,init_segment=t_720p_init.m4s \
	  input=t_480p.mp4,stream=video,segment_template=t_480p_\$Number\$.m4s,init_segment=t_480p_init.m4s \
	  input=t_1080p.mp4,stream=audio,segment_template=audio_\$Number\$.m4s,init_segment=audio_init.m4s \
	  --generate_static_live_mpd --mpd_output ./manifest.mpd \
	  --fragment_duration 8 \
	  --segment_duration 8
*/

// Callback with directory where all the generated files are.
type FormattingCallback func(string)

var qualities [3]models.VideoQuality = [3]models.VideoQuality{
	{Bitrate: 7000, ResolutionX: 1920, ResolutionY: 1080},
	{Bitrate: 2500, ResolutionX: 1280, ResolutionY: 720},
	{Bitrate: 1200, ResolutionX: 854, ResolutionY: 480},
}

// Generate different qualities for video and then generates the mpd manifest and all its fragments,
// it deletes all the files once the excecution of the function is completed, pass a callback func
// to do something with the files
func ConvertAndFormatToFragmentedMP4(videoPath string, drmInfo *models.DRMInfo, fn FormattingCallback) error {
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
	for i := 0; i < len(qualities); i++ {
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

func generateEncryptedFragmentedMP4(filePath, actionPath, name string, drmInfo *models.DRMInfo) error {
	var packagerArgs []string

	// Add input streams with encryption
	for _, quality := range qualities {
		packagerArgs = append(packagerArgs,
			fmt.Sprintf(
				"input=%s-%d.mp4,stream=video,segment_template=%s-%d_$Number$.m4s,"+
					"init_segment=%s-%d_init.m4s,"+
					"encryption_key=%s,key_id=%s,"+
					"protection_scheme=cenc",
				name, quality.ResolutionY,
				name, quality.ResolutionY,
				name, quality.ResolutionY,
				drmInfo.Key, drmInfo.KeyID,
			),
		)
	}

	// Add audio stream with encryption
	packagerArgs = append(packagerArgs,
		fmt.Sprintf(
			"input=%s,stream=audio,"+
				"segment_template=audio_$Number$.m4s,"+
				"init_segment=audio_init.m4s,"+
				"encryption_key=%s,key_id=%s,"+
				"protection_scheme=cenc",
			filePath,
			drmInfo.Key, drmInfo.KeyID,
		),
	)

	// Add general packager arguments
	packagerArgs = append(packagerArgs,
		"--generate_static_live_mpd",
		"--mpd_output", fmt.Sprintf("%s.mpd", name),
		"--fragment_duration", "8",
		"--segment_duration", "8",
		"--enable_raw_key_encryption",
		"--protection_systems", "Widevine,PlayReady",
	)

	cmd := exec.Command("packager", packagerArgs...)
	cmd.Dir = actionPath

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		log.Printf("Packager output: %s\nError: %s\n", outb.String(), errb.String())
		return err
	}

	return nil
}
