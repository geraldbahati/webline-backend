package utils

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/disintegration/imaging"
)

const maxUploadSize = 10 << 20 // 10 MB

// UploadFileToS3 uploads a single file to S3 after optimizing the image.
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

	// Optimize the image before uploading
	optimizedReader, err := optimizeImage(file, handler)
	if err != nil {
		return "", fmt.Errorf("failed to optimize image: %w", err)
	}

	if err := uploadToS3(ctx, s3Client, bucketName, filePath, optimizedReader); err != nil {
		return "", err
	}

	return filePath, nil
}

// UploadCustomFileToS3 uploads a single file to S3 after optimizing the file.
func UploadCustomFileToS3(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, s3Client *s3.Client, bucketName, uploadDir string) (string, error) {
	filePath, err := generateFilePath(uploadDir, fileHeader.Filename)
	if err != nil {
		return "", err
	}

	optimizedReader, err := optimizeImage(file, fileHeader)
	if err != nil {
		return "", fmt.Errorf("failed to optimize image: %w", err)
	}

	if err := uploadToS3(ctx, s3Client, bucketName, filePath, optimizedReader); err != nil {
		return "", err
	}

	return filePath, nil
}

// UploadMultipleFilesToS3 uploads multiple files to S3
func UploadMultipleFilesToS3(ctx context.Context, files []*multipart.FileHeader, s3Client *s3.Client, bucketName, uploadDir string) ([]string, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	uploadedFiles := make([]string, 0, len(files))
	errChan := make(chan error, len(files))

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

			mu.Lock()
			uploadedFiles = append(uploadedFiles, filePath)
			mu.Unlock()
		}(fileHeader)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return nil, err
		}
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

// uploadToS3 uploads a file to S3 given an io.Reader.
func uploadToS3(ctx context.Context, s3Client *s3.Client, bucketName, filePath string, reader io.Reader) error {
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &filePath,
		Body:   reader,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}
	return nil
}

// optimizeImage reads the provided image, decodes it, resizes it if its width exceeds maxWidth,
// and then re-encodes it as JPEG with quality 80.
func optimizeImage(file multipart.File, fileHeader *multipart.FileHeader) (io.Reader, error) {
	// Read the entire file into memory
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Create a reader from the data
	imgReader := bytes.NewReader(data)
	img, _, err := image.Decode(imgReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Define the maximum width for optimization
	maxWidth := 1024
	if img.Bounds().Dx() > maxWidth {
		// Resize the image while preserving the aspect ratio
		img = imaging.Resize(img, maxWidth, 0, imaging.Lanczos)
	}

	// Encode the image to JPEG with quality 80
	var outBuffer bytes.Buffer
	if err := jpeg.Encode(&outBuffer, img, &jpeg.Options{Quality: 80}); err != nil {
		return nil, fmt.Errorf("failed to encode optimized image: %w", err)
	}

	return &outBuffer, nil
}
