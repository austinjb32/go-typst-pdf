package storage

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func UploadPDF(filePath string) (string, error) {
	// Create a new AWS session
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewEnvCredentials(),
	}))

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

	// Define the expiration time for the pre-signed URL
	expiration := time.Hour // 1 hour

	// Generate the pre-signed URL
	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("AWS_BUCKET")),
		Key:    aws.String(key),
	})
	presignedURL, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate pre-signed URL: %w", err)
	}

	// Return the pre-signed URL
	return presignedURL, nil
}
