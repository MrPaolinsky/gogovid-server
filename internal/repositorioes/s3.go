package repositorioes

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Repo struct {
	client *s3.Client
}

func NewS3Repo() *S3Repo {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))

	if err != nil {
		log.Fatal("Error initializing S3 connection ", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("S3_ENDPOINT"))
		o.UsePathStyle = true
	})

	return &S3Repo{
		client: client,
	}
}

func (r *S3Repo) TestListObject() {
	output, err := r.client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println("first page results")

	for _, object := range output.Contents {
		log.Printf("key=%s size=%d\n", aws.ToString(object.Key), *object.Size)
	}
}

func (r *S3Repo) GetObjectByFileName(path string) (*s3.GetObjectOutput, error) {
	objectKey := "t/" + path
	result, err := r.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			log.Printf("Can't get object %s from bucket %s. No such key exists.\n", objectKey, os.Getenv("S3_BUCKET"))
			err = noKey
		} else {
			log.Printf("Couldn't get object %v:%v. Here's why: %v\n", os.Getenv("S3_BUCKET"), objectKey, err)
		}
		return nil, err
	}

	return result, nil
}

func (r *S3Repo) UploadFromPath(path string) error {

	return nil
}
