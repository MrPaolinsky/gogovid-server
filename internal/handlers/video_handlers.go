package handlers

import (
	"go-streamer/internal/repositorioes"
	videoservice "go-streamer/internal/services/video_service"
	"go-streamer/internal/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	vs *videoservice.VideoService
}

func NewVideoHandler(vs *videoservice.VideoService) *VideoHandler {
	return &VideoHandler{vs: vs}
}

func (vh *VideoHandler) ServeVideo(c *gin.Context) {
	fileId := c.Param("fileId")
    route := c.Query("route")
	file := route + "/" + fileId

    if route == "" {
        c.String(http.StatusBadRequest, "Missing route query parameter")
        return
    }

	repo := c.MustGet(utils.S3_REPO_CTX_KEY).(*repositorioes.S3Repo)

	fileResp, err := repo.GetObjectByFileName(file)
	if err != nil {
		log.Printf("Error getting file from S3: %v", err)
		c.String(http.StatusInternalServerError, "Error getting file")
		return
	}
	defer fileResp.Body.Close()

	var contentHeader string

	if utils.FileIsMPD(file) {
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

func (vh *VideoHandler) UploadVideo(c *gin.Context) {
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file found in request"})
		return
	}

	tempFilePath, err := vh.vs.SetupTempFile(file, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	isMP4, err := utils.FileIsMP4(tempFilePath)
	if !isMP4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Upload must be mp4"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vide is not an mp4 or it is corrupted"})
	}

	video, err := vh.vs.StoreVideo(tempFilePath, file, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video"})
		return
	}

	c.JSON(http.StatusOK, video)
}
