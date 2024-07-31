package appconfig

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DbUrl              string
	DefaultPageSize    int32
	DefaultPage        int32
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	AWSBucketName      string
	BusinessShortCode  string
	Passkey            string
	CallbackURL        string
	ConsumerKey        string
	ConsumerSecret     string
	AccountReference   string
	SMTPHost           string
	SMTPPort           int
	SMTPUsername       string
	SMTPPassword       string
	FromEmail          string
	FromName           string
	ToEmail            string
	FrontendURL        string
	BackendURL         string
}

func LoadConfig() Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	port, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return Config{
		Port: getEnv("PORT", "8080"),
		DbUrl: fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			getEnv("POSTGRES_USER", "postgres"),
			getEnv("POSTGRES_PASSWORD", "password"),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("POSTGRES_DB", "webline"),
		),
		DefaultPageSize:    100,
		DefaultPage:        1,
		AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
		AWSBucketName:      getEnv("AWS_BUCKET_NAME", "your_bucket_name"),
		BusinessShortCode:  getEnv("BUSINESS_SHORTCODE", ""),
		Passkey:            getEnv("PASSKEY", ""),
		CallbackURL:        getEnv("CALLBACK_URL", ""),
		ConsumerKey:        getEnv("CONSUMER_KEY", ""),
		ConsumerSecret:     getEnv("CONSUMER_SECRET", ""),
		AccountReference:   getEnv("ACCOUNT_REFERENCE", ""),
		SMTPHost:           getEnv("SMTP_HOST", ""),
		SMTPPort:           port,
		SMTPUsername:       getEnv("SMTP_USERNAME", ""),
		SMTPPassword:       getEnv("SMTP_PASSWORD", ""),
		FromEmail:          getEnv("FROM_EMAIL", ""),
		FromName:           getEnv("FROM_NAME", ""),
		ToEmail:            getEnv("TO_EMAIL", ""),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
		BackendURL:         getEnv("BACKEND_URL", "http://localhost:8080"),
	}
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
