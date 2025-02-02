package repositorioes

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
