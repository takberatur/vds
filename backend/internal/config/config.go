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
	OutboundProxyURL      string
	YTDLPImperersonate    string
	ProxyForAll           bool
	ProxyIncludeHosts     string
	ProxyExcludeHosts     string
	TwitchMaxSeconds      int
	YTDLPJSRuntime        string
	DisableCookiesFile    bool
	CookiesFilePath       string
	YoutubeUseCookies     string
	YoutubePlayerClient   string
	YoutubeCustomDisabled bool

	// Telegram bot
	TelegramBotToken      string
	TelegramChatID        string
	TelegramNotifications bool
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
		MinioUseSSL:           getEnvBool("MINIO_USE_SSL", false),
		BCryptCost:            getEnvInt("BCRYPT_COST", 10),
		EncryptionKey:         getEnv("ENCRYPTION_KEY", "secret"),
		CentrifugoURL:         getEnv("CENTRIFUGE_URL", "ws://infrastructure-centrifugo:8000/connection/websocket"),
		CentrifugoAPIKey:      getEnv("CENTRIFUGO_API_KEY", ""),
		CentrifugoTokenSecret: getEnv("CENTRIFUGO_TOKEN_SECRET", ""),
		OutboundProxyURL:      getEnv("OUTBOUND_PROXY_URL", ""),
		YTDLPImperersonate:    getEnv("YTDLP_IMPERSONATE", ""),
		ProxyForAll:           getEnvBool("PROXY_FOR_ALL", false),
		ProxyIncludeHosts:     getEnv("PROXY_INCLUDE_HOSTS", ""),
		ProxyExcludeHosts:     getEnv("PROXY_EXCLUDE_HOSTS", ""),
		TwitchMaxSeconds:      getEnvInt("TWITCH_MAX_SECONDS", 300),
		YTDLPJSRuntime:        getEnv("YTDLP_JS_RUNTIME", ""),
		DisableCookiesFile:    getEnvBool("DISABLE_COOKIES_FILE", false),
		CookiesFilePath:       getEnv("COOKIES_FILE_PATH", "/app/cookies.txt"),
		YoutubeUseCookies:     getEnv("YOUTUBE_USE_COOKIES", ""),
		YoutubePlayerClient:   getEnv("YOUTUBE_PLAYER_CLIENT", ""),
		YoutubeCustomDisabled: getEnvBool("YOUTUBE_CUSTOM_DISABLED", false),
		// Telegram bot
		TelegramBotToken:      getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:        getEnv("TELEGRAM_CHAT_ID", ""),
		TelegramNotifications: getEnvBool("TELEGRAM_NOTIFICATIONS", false),
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

func getEnvBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}
