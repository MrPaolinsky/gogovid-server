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
	s3Repo *repositorioes.S3Repo
	dbRepo *repositorioes.DBRepo
}

func NewVideoService(s3 *repositorioes.S3Repo, db *repositorioes.DBRepo) *VideoService {
	return &VideoService{s3Repo: s3, dbRepo: db}
}

func (vs *VideoService) StoreVideo(path string, file *multipart.FileHeader, c *gin.Context) (*models.Video, error) {
	drmInfo, err := vs.generateKeysInfo()
	if err != nil {
		log.Println("Error generating keys info:", err)
		return nil, err
	}

	err = utils.ConvertAndFormatToFragmentedMP4(path, drmInfo, func(path string) {
		err := vs.s3Repo.UploadFragmentedVideoFromPath(path, file.Filename)

		if err != nil {
			log.Println(err)
			return
		}
	})

	if err != nil {
		log.Println(err)
		return nil, err
	}

	userId := c.MustGet(utils.USER_ID_CTX_KEY).(uint)
	log.Println("HEHERHEHREHREHRHEHRER:", userId)
	videosCrud := cruds.NewVideosCrud(vs.dbRepo)
	video := &models.Video{
		Name:            file.Filename,
		UserId:          userId,
		DurationMinutes: 0, // TODO: SET DURATION MINUTES
	}

	newVid, err := videosCrud.CreateVideo(video)
	if err != nil {
		log.Println("Error inserting new video in database")
		return nil, err
	}

	err = vs.storeKeysForVideo(newVid.ID, drmInfo)
	if err != nil {
		log.Println("Error inserting new drm keys in database")
		return nil, err
	}

	return video, nil
}

func (vs *VideoService) SetupTempFile(file *multipart.FileHeader, c *gin.Context) (string, error) {
	tempFilePath := fmt.Sprintf("/tmp/%s", file.Filename)

	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		log.Println("Error setting up temporary file: ", err)
		return "", err
	}

	return tempFilePath, nil
}

func (vs *VideoService) storeKeysForVideo(videoId uint, keys []*models.DRMInfo) error {
	var videoKeys []*models.DRMKey
	for _, key := range keys {
		k := models.DRMKey{
			VideoId: videoId,
			DRMInfo: models.DRMInfo{
				KeyID: key.KeyID,
				Key:   key.Key,
				Label: key.Label,
			},
		}
		videoKeys = append(videoKeys, &k)
	}

	drmCrud := cruds.NewDRMCrud(vs.dbRepo)
	err := drmCrud.StoreManyKeys(videoKeys)
	if err != nil {
		return err
	}

	return nil
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
