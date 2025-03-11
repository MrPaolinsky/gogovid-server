package signing

import (
	"fmt"
	signingservice "go-streamer/internal/services/signing_service"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SigningHandler struct {
	ss *signingservice.SigningService
}

func NewSigningHandler(ss *signingservice.SigningService) *SigningHandler {
	return &SigningHandler{ss: ss}
}

func (sh *SigningHandler) GetSignedUrl(c *gin.Context) {
	vID := c.Query("video_id")
	file := c.Param("file")

	if vID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing vID (video id) in query parameters"})
		return
	}
	parsedVID, err := strconv.ParseUint(vID, 10, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video ID must be a number"})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseUrl := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	url, err := sh.ss.SignUrlForVideoFile(uint(parsedVID), file, baseUrl)
	if err != nil || url == "" {
		log.Println("Error signing url:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error signing url"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"signed_url": url})
}
