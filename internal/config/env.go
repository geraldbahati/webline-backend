package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	DbUrl           string
	DefaultPageSize int32
	DefaultPage     int32
}

func LoadConfig() Config {
	if err := godotenv.Load(".env"); err != nil {
		// Added this for testing and debugging reasons
		// To be removed
		log.Fatal("Error loading .env file:", err)
	}

	return Config{
		Port: getEnv("PORT", "8080"),
		DbUrl: fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASSWORD", "password"),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_NAME", "webline"),
		),
		DefaultPageSize: 100,
		DefaultPage:     1,
	}
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
