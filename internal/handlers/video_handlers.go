package handlers

import (
	"fmt"
	"go-streamer/internal/repositorioes"
	"go-streamer/internal/utils"
	"log"
	"net/http"

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

func UploadVideo(c *gin.Context) {
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file found in request"})
		return
	}

	tempFilePath := fmt.Sprintf("/tmp/%s", file.Filename)
	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	err = utils.ConvertAndFormatToFragmentedMP4(tempFilePath, func(path string) {
		repo := c.MustGet(utils.S3_REPO_CTX_KEY).(*repositorioes.S3Repo)
		err := repo.UploadFragmentedVideoFromPath(path, file.Filename)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload files"})
			log.Println(err)
			return
		}
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		log.Println(err)
	}

}
