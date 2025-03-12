package signingservice

import (
	"errors"
	"fmt"
	"go-streamer/internal/models"
	"go-streamer/internal/repositorioes"
	"go-streamer/internal/repositorioes/cruds"
	"os"
	"time"

	"github.com/kodefluence/aurelia"
)

type SigningService struct {
	dbRepo *repositorioes.DBRepo
	s3Repo *repositorioes.S3Repo
}

func NewSigningService(db *repositorioes.DBRepo, s3 *repositorioes.S3Repo) *SigningService {
	return &SigningService{dbRepo: db, s3Repo: s3}
}

func (ss *SigningService) SignUrlForVideoFile(vID uint, file, baseUrl string) (string, error) {
	vCrud := cruds.NewVideosCrud(ss.dbRepo)
	video, err := vCrud.GetVideo(vID)

	if err != nil || video == nil {
		return "", err
	}

	// Check if file exists for video
	videoParts, err := ss.s3Repo.ListObjectsInFolder(video.Route)
	if err != nil {
		return "", err
	}

	url := baseUrl + "/video/" + file + "?route="
	for i, part := range videoParts {
		if part == file {
			url += video.Route
			break
		}
		if i == len(videoParts)-1 {
			return "", errors.New("Requested resource do not exists for video")
		}
	}

	signature, expiry := ss.generateSignature(file, video)
	url += "&signature=" + signature + "&expires_at=" + expiry + "&video_id=" + fmt.Sprintf("%d", video.ID)

	return url, nil
}

func (ss *SigningService) generateSignature(file string, video *models.Video) (string, string) {
	s := os.Getenv("URL_SIGNING_KEY")
	expiresAt := time.Now().Add(5 * time.Minute).Unix()
	signature := aurelia.Hash(s, fmt.Sprintf("%s:%d:%d", file, video.ID, expiresAt))

	return signature, fmt.Sprintf("%d", expiresAt)
}
