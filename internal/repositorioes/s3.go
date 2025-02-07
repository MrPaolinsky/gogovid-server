package repositorioes

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

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
	objectKey := "t.mp4/" + path
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

func (r *S3Repo) UploadFragmentedVideoFromPath(path string, folderName string) error {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Println("Error reading files in folder:", path)
		return err
	}

	if folderName[len(folderName)-1:] != "/" {
		folderName += "/"
	}

	wg := new(sync.WaitGroup)
	processFailed := false
	wg.Add(len(files))

	for _, file := range files {
		go r.uploadSinlgeFile(path, file, folderName, &processFailed, wg)
	}

	wg.Wait()

	if processFailed {
		r.DeleteFolder(folderName)
		return fmt.Errorf("Error uploading files, reverting sccessful uploads")
	}

	return nil
}

func (r *S3Repo) DeleteFolder(folderName string) error {
	if folderName[len(folderName)-1:] != "/" {
		folderName += "/"
	}

	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Prefix: aws.String(folderName),
	}

	paginator := s3.NewListObjectsV2Paginator(r.client, listInput)
	var objectIdentifiers []types.ObjectIdentifier

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}

		for _, object := range page.Contents {
			objectIdentifiers = append(objectIdentifiers, types.ObjectIdentifier{
				Key: object.Key,
			})
		}
	}

	if len(objectIdentifiers) == 0 {
		log.Printf("Folder '%s' is empty or does not exist in bucket '%s'.\n", folderName, os.Getenv("S3_BUCKET"))
		return nil
	}

	deleteInput := &s3.DeleteObjectsInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Delete: &types.Delete{
			Objects: objectIdentifiers,
			Quiet:   aws.Bool(true), // Suppress errors for individual objects
		},
	}

	_, err := r.client.DeleteObjects(context.TODO(), deleteInput)
	if err != nil {
		return fmt.Errorf("failed to delete objects: %v", err)
	}

	fmt.Printf("Deleted %d objects from folder '%s' in bucket '%s'.\n", len(objectIdentifiers), folderName, os.Getenv("S3_BUCKET"))
	return nil
}

func (r *S3Repo) uploadSinlgeFile(
	parentPath string,
	file os.DirEntry,
	destinationFolder string,
	failureFlag *bool,
	wg *sync.WaitGroup,
) error {
	defer wg.Done()

	if !file.IsDir() {
		localFilePath := filepath.Join(parentPath, file.Name())
		objectKey := destinationFolder + file.Name()
		fileContent, err := os.Open(localFilePath)

		if err != nil {
			log.Println("Error opening file:", localFilePath)
			*failureFlag = true
			return err
		}

		defer fileContent.Close()

		_, err = r.client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("S3_BUCKET")),
			Key:    aws.String(objectKey),
			Body:   fileContent,
		})

		if err != nil {
			log.Println("Error uploading file to S3:", objectKey)
			*failureFlag = true
			return err
		}
	}
	return nil
}
