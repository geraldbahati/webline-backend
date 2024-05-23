package utils

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func UploadFileToS3(r *http.Request, s3Client *s3.Client, bucketName, uploadDir string) (string, error) {
	// parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // max 10 MB
		return "", fmt.Errorf("failed to parse multipart form: %w", err)
	}

	// get the file from the form
	file, handler, err := r.FormFile("image")
	if err != nil {
		return "", fmt.Errorf("failed to get the file from the form: %w", err)
	}
	defer file.Close()

	// create a unique file name
	fileName := fmt.Sprintf("%d-%s", time.Now().Unix(), handler.Filename)
	filePath := fmt.Sprintf("%s/%s", uploadDir, fileName)

	// upload to S3
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &filePath,
		Body:   file,
		//ACL:    types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return filePath, nil
}
