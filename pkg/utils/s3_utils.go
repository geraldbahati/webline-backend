package utils

import "fmt"

// ConstructS3URL constructs the full S3 URL for a given file path,
// using the specified bucket name and region.
func ConstructS3URL(bucketName, region, filePath string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, region, filePath)
}
