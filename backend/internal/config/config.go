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
	DBUser        string
	DBPassword    string
	DBName        string
	DBHost        string
	DBPort        string
	DBURL         string
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
	// Centrifugo Config
	CentrifugoURL         string
	CentrifugoAPIKey      string
	CentrifugoTokenSecret string
}

func LoadConfig() *Config {
	_ = godotenv.Load() // Load .env file if exists

	return &Config{
		AppPort:               getEnv("APP_PORT", "5001"),
		AppEnv:                getEnv("APP_ENV", "development"),
		ClientURL:             getEnv("CLIENT_URL", "http://localhost:3000"),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		DBUser:                getEnv("DB_USER", "postgres"),
		DBPassword:            getEnv("DB_PASSWORD", "postgres"),
		DBName:                getEnv("DB_NAME", "video_downloader"),
		DBHost:                getEnv("DB_HOST", "localhost"),
		DBPort:                getEnv("DB_PORT", "5432"),
		DBURL:                 getEnv("DB_URL", ""),
		RedisAddr:             getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:         getEnv("REDIS_PASSWORD", ""),
		JWTSecret:             getEnv("JWT_SECRET", "secret"),
		JWTExpiryHour:         getEnv("JWT_EXPIRY_HOUR", "24"),
		MinioEndpoint:         getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey:        getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:        getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:           getEnv("MINIO_BUCKET", "videos"),
		MinioUseSSL:           getEnv("MINIO_USE_SSL", "false") == "true",
		BCryptCost:            getEnvInt("BCRYPT_COST", 10),
		EncryptionKey:         getEnv("ENCRYPTION_KEY", "secret"),
		CentrifugoURL:         getEnv("CENTRIFUGE_URL", "ws://infrastructure-centrifugo:8000/connection/websocket"),
		CentrifugoAPIKey:      getEnv("CENTRIFUGO_API_KEY", ""),
		CentrifugoTokenSecret: getEnv("CENTRIFUGO_TOKEN_SECRET", ""),
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
