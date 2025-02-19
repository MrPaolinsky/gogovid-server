package cruds

import (
	"go-streamer/internal/models"
	"go-streamer/internal/repositorioes"
)

type VideosCrud struct {
	dbRepo *repositorioes.DBRepo
}

func NewVideosCrud(dbRepo *repositorioes.DBRepo) *VideosCrud {
	return &VideosCrud{
		dbRepo: dbRepo,
	}
}

func (c *VideosCrud) CreateVideo(video *models.Video) (*models.Video, error) {
	err := c.dbRepo.Db.Create(video).Error
	return video, err
}

func (c *VideosCrud) GetVideo(id uint) (*models.Video, error) {
	var video models.Video
	err := c.dbRepo.Db.First(&video, id).Error
	return &video, err
}
