package videoservice

import (
	"fmt"
	"go-streamer/internal/models"
	"go-streamer/internal/repositorioes"
	"go-streamer/internal/repositorioes/cruds"
	"go-streamer/internal/utils"
	"log"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type VideoService struct {
	c    *gin.Context
	repo *repositorioes.S3Repo
}

func NewVideoService(repo *repositorioes.S3Repo) *VideoService {
	return &VideoService{repo: repo}
}

func (vs *VideoService) StoreVideo(path string, file *multipart.FileHeader) (*models.Video, error) {
	drmInfo, err := vs.generateKeysInfo()
	if err != nil {
		log.Println("Error generating keys info:", err)
		return nil, err
	}

	err = utils.ConvertAndFormatToFragmentedMP4(path, drmInfo, func(path string) {
		repo := vs.c.MustGet(utils.S3_REPO_CTX_KEY).(*repositorioes.S3Repo)
		err := repo.UploadFragmentedVideoFromPath(path, file.Filename)

		if err != nil {
			log.Println(err)
			return
		}
	})

	if err != nil {
		log.Println(err)
		return nil, err
	}

	dbRepo := vs.c.MustGet(utils.DB_REPO_CTX_KEY).(*repositorioes.DBRepo)
	videosCrud := cruds.NewVideosCrud(dbRepo)
	video := &models.Video{
		Name:            file.Filename,
		UserId:          0, // TODO: get user id
		DurationMinutes: 0, // TODO: SET DURATION MINUTES
	}

	if _, err := videosCrud.CreateVideo(video); err != nil {
		return nil, err
	}

	return video, nil
}

func (vs *VideoService) SetupTempFile(file *multipart.FileHeader) (string, error) {
	tempFilePath := fmt.Sprintf("/tmp/%s", file.Filename)

	if err := vs.c.SaveUploadedFile(file, tempFilePath); err != nil {
		log.Println("Error setting up temporary file: ", err)
		return "", err
	}

	return tempFilePath, nil
}

func (vs *VideoService) generateKeysInfo() ([]*models.DRMInfo, error) {
	var keys []*models.DRMInfo

	for _, label := range models.Labels {
		kID, k, err := utils.GenerateDRMKey()

		if err != nil {
			log.Println("Error generating DRM Key")
			return nil, err
		}

		key := &models.DRMInfo{
			Key:   k,
			KeyID: kID,
			Label: label,
		}
		keys = append(keys, key)
	}

	return keys, nil
}
