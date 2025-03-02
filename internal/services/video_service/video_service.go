package videoservice

import (
	"fmt"
	"go-streamer/internal/models"
	"go-streamer/internal/repositorioes"
	"go-streamer/internal/repositorioes/cruds"
	"go-streamer/internal/utils"
	"log"
	"mime/multipart"
	"os"

	"github.com/gin-gonic/gin"
)

type VideoService struct {
	s3Repo *repositorioes.S3Repo
	dbRepo *repositorioes.DBRepo
}

func NewVideoService(s3 *repositorioes.S3Repo, db *repositorioes.DBRepo) *VideoService {
	return &VideoService{s3Repo: s3, dbRepo: db}
}

// Will create new DB entity, and create a goroutine to format it, the goroutine
// will delete the temp files once done.
func (vs *VideoService) StoreVideo(path string, file *multipart.FileHeader, c *gin.Context) (*models.Video, error) {
	userId := c.MustGet(utils.USER_ID_CTX_KEY).(uint)
	videosCrud := cruds.NewVideosCrud(vs.dbRepo)
	video := &models.Video{
		Name:            file.Filename,
		UserId:          userId,
		Formatted:       false,
		DurationMinutes: 0, // TODO: SET DURATION MINUTES
	}

	nv, err := videosCrud.CreateVideo(video)
	if err != nil {
		log.Println("Error inserting new video in database")
		return nil, err
	}

	go func() {
		err := vs.formatAndUploadVideo(path, file, nv)

		if err != nil {
			log.Println(err)
			vs.dbRepo.Db.Delete(nv)
			return
		} else {
			nv.Formatted = true
			vs.dbRepo.Db.Save(nv)
		}
	}()

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

// Intended to be used in a goroutine, pass a VideoID, and it will process it
// and then update the Route property of the passed video.
func (vs *VideoService) formatAndUploadVideo(
	path string,
	file *multipart.FileHeader,
	nv *models.Video,
) error {
	defer func() {
		err1 := os.Remove(path)
		err2 := os.RemoveAll(path)

		if err1 != nil || err2 != nil {
			log.Println("Error deleting upload files: ", err1, "\n", err2)
		}
	}()

	err := utils.ConvertAndFormatToFragmentedMP4(
		path,
		func(path string) { vs.uploadToS3(path, file, nv) },
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

func (vs *VideoService) uploadToS3(path string, file *multipart.FileHeader, nv *models.Video) {
	route, err := vs.s3Repo.UploadFragmentedVideoFromPath(path, file.Filename)

	if err != nil {
		log.Println(err)
		vs.dbRepo.Db.Delete(nv)
		return
	}
	nv.Route = route
	vs.dbRepo.Db.Save(nv)
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
