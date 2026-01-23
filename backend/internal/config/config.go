package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort       string
	AppEnv        string
	ClientURL     string
	DatabaseURL   string
	RedisAddr     string
	RedisPassword string
	JWTSecret     string
	JWTExpiryHour string
	// MinIO Config
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
	// Encryption Config
	BCryptCost    int
	EncryptionKey string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		AppPort:       getEnv("APP_PORT", "3000"),
		AppEnv:        getEnv("APP_ENV", "development"),
		ClientURL:     getEnv("CLIENT_URL", "http://localhost:5173"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/dbname"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		JWTSecret:     getEnv("JWT_SECRET", "super-secret-key"),
		JWTExpiryHour: getEnv("JWT_EXPIRY_HOUR", "24"),

		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:    getEnv("MINIO_BUCKET", "video-downloader"),
		MinioUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
		// Encryption Config
		BCryptCost:    getEnvInt("BCRYPT_COST", 10),
		EncryptionKey: getEnv("ENCRYPTION_KEY", "ef36bf2f945e74c2bdc2480e5e726b25629ba9886824b059d0aba3196c1d1f0f"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}
