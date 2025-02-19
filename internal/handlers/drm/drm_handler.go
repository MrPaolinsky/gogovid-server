package drm

import (
	"go-streamer/internal/models"
	"go-streamer/internal/repositorioes"
	"go-streamer/internal/repositorioes/cruds"
	"go-streamer/internal/services"
	"go-streamer/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DRMHandler struct {
	drmService *services.DRMService
}

func NewDRMHandler(drmService *services.DRMService) *DRMHandler {
	return &DRMHandler{
		drmService: drmService,
	}
}

func (h *DRMHandler) HandleLicenseRequest(c *gin.Context) {
	var req models.LicenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get video and DRM info
	dbRepo := c.MustGet(utils.DB_REPO_CTX_KEY).(*repositorioes.DBRepo)
	videosCrud := cruds.NewVideosCrud(dbRepo)
	video, err := videosCrud.GetVideo(req.VideoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	// TODO Validate if user has access to this content

	c.Data(http.StatusOK, "application/octet-stream", []byte(video.DRMInfo.Key))
}
