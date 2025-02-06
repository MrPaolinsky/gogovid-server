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
func ConvertAndFormatToFragmentedMP4(videoPath string, fn FormattingCallback) error {
	name := fmt.Sprintf("gogovid-%d", time.Now().UnixMilli())
	actionPath := "/tmp/" + name

	defer func() {
		err1 := os.Remove(videoPath)
		err2 := os.RemoveAll(actionPath)

		if err1 != nil || err2 != nil {
			log.Println("Error deleting upload files: ", err1, "\n", err2)
		}
	}()

	newFolderCmd := exec.Command("mkdir", name)
	newFolderCmd.Dir = "/tmp/"
	if err := newFolderCmd.Run(); err != nil {
		log.Println("Error creating action folder")
		return err
	}

	err := generateVideoResolutionsForPackager(videoPath, actionPath, name)
	if err != nil {
		log.Println("Error generation multiple resolutions for video")
		return err
	}

	err = generateFragmentedMP4(actionPath, name)

	if err != nil {
		log.Println("Error generatiing fragmented mp4")
		return err
	}

	fn(actionPath)

	return nil
}

func generateVideoResolutionsForPackager(filePath string, actionPath string, name string) error {
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

func generateFragmentedMP4(actionPath string, name string) error {
	fragmentQualityCommands := []string{}

	for i := 0; i < len(qualities); i++ {
		fragmentQualityCommands = append(
			fragmentQualityCommands,
			fmt.Sprintf(
				"input=%s-%d.mp4,stream=video,segment_template=%s-%d_$Number$.m4s,init_segment=%s-%d_init.m4s",
				name, qualities[i].ResolutionY,
				name, qualities[i].ResolutionY,
				name, qualities[i].ResolutionY,
			),
		)
	}

	fragmentQualityCommands = append(
		fragmentQualityCommands,
		"--generate_static_live_mpd",
		"--mpd_output", fmt.Sprintf("%s.mpd", name),
		"--fragment_duration", "8",
		"--segment_duration", "8",
	)

	cmd := exec.Command(
		"packager",
		fragmentQualityCommands...,
	)

	cmd.Dir = actionPath

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		log.Println("out:", outb.String(), "\nerr:", errb.String())
		return err
	}

	return nil
}
