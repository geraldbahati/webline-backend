package utils

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const maxUploadSize = 10 << 20 // 10 MB

// UploadFileToS3 uploads a single file to S3
func UploadFileToS3(ctx context.Context, r *http.Request, s3Client *s3.Client, bucketName, uploadDir string) (string, error) {
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		return "", fmt.Errorf("failed to parse multipart form: %w", err)
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		return "", fmt.Errorf("failed to get the file from the form: %w", err)
	}
	defer file.Close()

	filePath, err := generateFilePath(uploadDir, handler.Filename)
	if err != nil {
		return "", err
	}

	if err := uploadToS3(ctx, s3Client, bucketName, filePath, file); err != nil {
		return "", err
	}

	return filePath, nil
}

// UploadMultipleFilesToS3 uploads multiple files to S3
func UploadMultipleFilesToS3(ctx context.Context, files []*multipart.FileHeader, s3Client *s3.Client, bucketName, uploadDir string) ([]string, error) {
	var wg sync.WaitGroup
	uploadedFiles := make([]string, len(files))
	errChan := make(chan error, len(files))
	filePaths := make(chan string, len(files))

	for _, fileHeader := range files {
		wg.Add(1)
		go func(fileHeader *multipart.FileHeader) {
			defer wg.Done()

			file, err := fileHeader.Open()
			if err != nil {
				errChan <- fmt.Errorf("failed to open file: %w", err)
				return
			}
			defer file.Close()

			filePath, err := generateFilePath(uploadDir, fileHeader.Filename)
			if err != nil {
				errChan <- err
				return
			}

			if err := uploadToS3(ctx, s3Client, bucketName, filePath, file); err != nil {
				errChan <- err
				return
			}

			filePaths <- filePath
		}(fileHeader)
	}

	go func() {
		wg.Wait()
		close(errChan)
		close(filePaths)
	}()

	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	for filePath := range filePaths {
		uploadedFiles = append(uploadedFiles, filePath)
	}

	return uploadedFiles, nil
}

// DeleteFileFromS3 deletes a single file from S3
func DeleteFileFromS3(ctx context.Context, s3Client *s3.Client, bucketName, filePath string) error {
	_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucketName,
		Key:    &filePath,
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}
	return nil
}

// DeleteMultipleFilesFromS3 deletes multiple files from S3
func DeleteMultipleFilesFromS3(ctx context.Context, s3Client *s3.Client, bucketName string, filePaths []string) error {
	objects := make([]types.ObjectIdentifier, len(filePaths))
	for i, filePath := range filePaths {
		objects[i] = types.ObjectIdentifier{
			Key: &filePath,
		}
	}

	quiet := false
	_, err := s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: &bucketName,
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   &quiet,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete files from S3: %w", err)
	}

	return nil
}

// generateFilePath creates a unique file path for the uploaded file
func generateFilePath(uploadDir, filename string) (string, error) {
	if uploadDir == "" || filename == "" {
		return "", fmt.Errorf("upload directory or filename cannot be empty")
	}
	fileName := fmt.Sprintf("%d-%s", time.Now().Unix(), filepath.Base(filename))
	return filepath.Join(uploadDir, fileName), nil
}

// uploadToS3 uploads a file to S3
func uploadToS3(ctx context.Context, s3Client *s3.Client, bucketName, filePath string, file multipart.File) error {
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &filePath,
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}
	return nil
}
