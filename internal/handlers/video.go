package handlers

import (
	"go-streamer/internal/repositorioes"
	"go-streamer/internal/utils"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ServeVideo(c *gin.Context) {
	fileId := c.Param("fileId")
	file := fileId

	repo := c.MustGet(utils.S3_REPO_CTX_KEY).(*repositorioes.S3Repo)

	fileResp, err := repo.GetObjectByFileName(file)
	if err != nil {
		log.Printf("Error getting file from S3: %v", err)
		c.String(http.StatusInternalServerError, "Error getting file")
		return
	}
	defer fileResp.Body.Close()

	var contentHeader string

	if fileIsMPD(file) {
		contentHeader = "application/dash+xml"
	} else {
		contentHeader = "video/mp4"
	}

	c.DataFromReader(
		http.StatusOK,
		*fileResp.ContentLength,
		contentHeader,
		fileResp.Body,
		map[string]string{},
	)
}

func fileIsMPD(fileName string) bool {
	return strings.HasSuffix(fileName, ".mpd")
}

func ServeVideoSegment(c *gin.Context) {
	videoId := c.Param("videoId")
	segmentId := c.Param("segmentId")
	segmentFile := "video/" + videoId + segmentId

	repo := c.MustGet(utils.S3_REPO_CTX_KEY).(*repositorioes.S3Repo)

	fileResp, err := repo.GetObjectByFileName(segmentFile)
	if err != nil {
		log.Printf("Error getting video segment from S3: %v", err)
		c.String(http.StatusInternalServerError, "Error getting video segment from S3")
		return
	}
	defer fileResp.Body.Close()

	c.DataFromReader(
		http.StatusOK,
		*fileResp.ContentLength,
		"application/dash+xml",
		fileResp.Body,
		map[string]string{},
	)
}
