package appconfig

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the configuration for the application.
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
	RedisHost          string
	RedisPort          int
	RedisPassword      string
	RedisDB            int
	RedisPoolSize      int
	RedisMinIdleConns  int
	RedisTTL           time.Duration
	RedisRateLimit     int
	Env                string
}

// LoadConfig reads environment variables and constructs a Config struct.
// It returns a pointer to Config and an error if any required variable is missing or invalid.
func LoadConfig() (*Config, error) {
	// Helper function to get environment variables with fallbacks.
	getEnv := func(key string, fallback string) string {
		if value, exists := os.LookupEnv(key); exists {
			return value
		}
		return fallback
	}

	// Parse SMTP_PORT with error handling.
	smtpPortStr := getEnv("SMTP_PORT", "587")
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	// Parse Redis_PORT with error handling.
	redisPortStr := getEnv("REDIS_PORT", "6379")
	redisPort, err := strconv.Atoi(redisPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT: %w", err)
	}

	// Parse Redis_DB with error handling.
	redisDBStr := getEnv("REDIS_DB", "0")
	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	// Parse Redis_PoolSize with error handling.
	redisPoolSizeStr := getEnv("REDIS_POOL_SIZE", "10")
	redisPoolSize, err := strconv.Atoi(redisPoolSizeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_POOL_SIZE: %w", err)
	}

	// Parse Redis_MinIdleConns with error handling.
	redisMinIdleConnsStr := getEnv("REDIS_MIN_IDLE_CONNS", "5")
	redisMinIdleConns, err := strconv.Atoi(redisMinIdleConnsStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_MIN_IDLE_CONNS: %w", err)
	}

	// Parse Redis_TTL with error handling -> 10 minutes by default
	redisTTLStr := getEnv("REDIS_TTL", "10m")
	redisTTL, err := time.ParseDuration(redisTTLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_TTL: %w", err)
	}

	// Parse Redis_RateLimit with error handling.
	redisRateLimitStr := getEnv("REDIS_RATE_LIMIT", "100")
	redisRateLimit, err := strconv.Atoi(redisRateLimitStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_RATE_LIMIT: %w", err)
	}

	// Construct the database URL.
	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnv("POSTGRES_USER", "postgres"),
		getEnv("POSTGRES_PASSWORD", "password"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("POSTGRES_DB", "webline"),
	)

	// Parse Server Port.
	serverPort := getEnv("BACKEND_PORT", "8080")

	// Initialize Config struct.
	cfg := &Config{
		Port:               serverPort,
		DbUrl:              dbUrl,
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
		SMTPPort:           smtpPort,
		SMTPUsername:       getEnv("SMTP_USERNAME", ""),
		SMTPPassword:       getEnv("SMTP_PASSWORD", ""),
		FromEmail:          getEnv("FROM_EMAIL", ""),
		FromName:           getEnv("FROM_NAME", ""),
		ToEmail:            getEnv("TO_EMAIL", ""),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
		BackendURL:         getEnv("BACKEND_URL", "http://localhost:8080"),
		RedisHost:          getEnv("REDIS_HOST", "localhost"),
		RedisPort:          redisPort,
		RedisPassword:      getEnv("REDIS_PASSWORD", "password"),
		RedisDB:            redisDB,
		RedisPoolSize:      redisPoolSize,
		RedisMinIdleConns:  redisMinIdleConns,
		RedisTTL:           time.Duration(redisTTL),
		RedisRateLimit:     redisRateLimit,
		Env:                getEnv("ENV", "local"),
	}

	return cfg, nil
}
