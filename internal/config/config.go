package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	AWS      AWSConfig
	Upload   UploadConfig
	CDN      CDNConfig
	SMTP     SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret                string
	ExpiresIn             time.Duration
	RefreshTokenExpiresIn time.Duration
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	S3Bucket        string
	S3Endpoint      string
	SQSEndpoint     string
	EventQueueName  string
}

type UploadConfig struct {
	Path           string
	MaxFileSize    int64
	UploadProvider string
}

// CDNConfig holds the public base URL that serves uploaded assets.
// Only the storage key is persisted in the database; the public URL is
// rebuilt on every response so that changing CDN host never requires a
// data migration.
type CDNConfig struct {
	BaseURL string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	jwtExpiresIn, _ := time.ParseDuration(getEnv("JWT_EXPIRES_IN", "15m"))
	jwtRefreshTokenExpiresIn, _ := time.ParseDuration(getEnv("JWT_REFRESH_TOKEN_EXPIRES_IN", "168h"))

	maxFileSize, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "10485760"), 10, 64)
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "1025"))

	config := &Config{
		Server: ServerConfig{
			Port:    getEnv("SERVER_PORT", getEnv("PORT", "8080")),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "ecommerce"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:                getEnv("JWT_SECRET", "your-secret-key"),
			ExpiresIn:             jwtExpiresIn,
			RefreshTokenExpiresIn: jwtRefreshTokenExpiresIn,
		},
		AWS: AWSConfig{
			Region:          getEnv("AWS_REGION", "us-east-1"),
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			S3Bucket:        getEnv("AWS_S3_BUCKET_NAME", ""),
			S3Endpoint:      getEnv("AWS_S3_ENDPOINT", ""),
			SQSEndpoint:     getEnv("AWS_SQS_ENDPOINT", getEnv("AWS_S3_ENDPOINT", "")),
			EventQueueName:  getEnv("AWS_EVENT_QUEUE_NAME", ""),
		},
		Upload: UploadConfig{
			Path:           getEnv("UPLOAD_DIR", "./uploads"),
			MaxFileSize:    maxFileSize,
			UploadProvider: getEnv("UPLOAD_PROVIDER", "local"),
		},
		CDN: CDNConfig{
			BaseURL: strings.TrimRight(getEnv("CDN_BASE_URL", "http://localhost:8081/uploads"), "/"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "localhost"),
			Port:     smtpPort,
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@shop.com"),
		},
	}
	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
