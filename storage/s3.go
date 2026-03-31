package storage

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	awsSession *session.Session
	sessionOnce sync.Once
)

func getSession() *session.Session {
	sessionOnce.Do(func() {
		awsSession = session.Must(session.NewSession(&aws.Config{
			Region:      aws.String(os.Getenv("AWS_REGION")),
			Credentials: credentials.NewEnvCredentials(),
		}))
	})
	return awsSession
}

func UploadPDF(filePath string) (string, error) {
	sess := getSession()

	// Create an S3 uploader
	uploader := s3manager.NewUploader(sess)

	// Open the file to upload
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Generate a unique key for the file
	key := fmt.Sprintf("pdfs/%d.pdf", time.Now().UnixNano())

	// Upload the file to S3
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("AWS_BUCKET")),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	// Create an S3 client to generate the pre-signed URL
	s3Client := s3.New(sess)

	// Generate the pre-signed URL
	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("AWS_BUCKET")),
		Key:    aws.String(key),
	})
	presignedURL, err := req.Presign(time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to generate pre-signed URL: %w", err)
	}

	return presignedURL, nil
}
