package config

import (
	"ecommerce/utils"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	AWS      AWSConfig
	Upload   UploadConfig
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
}

type UploadConfig struct {
	Path        string
	MaxFileSize int64
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	jwtExpiresIn, _ := time.ParseDuration(utils.GetEnv("JWT_EXPIRES_IN", "15m"))
	jwtRefreshTokenExpiresIn, _ := time.ParseDuration(utils.GetEnv("JWT_REFRESH_TOKEN_EXPIRES_IN", "7d"))

	maxFileSize, _ := strconv.ParseInt(utils.GetEnv("MAX_UPLOAD_SIZE", "10485760"), 10, 64)

	config := &Config{
		Server: ServerConfig{
			Port:    utils.GetEnv("SERVER_PORT", "8080"),
			GinMode: utils.GetEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     utils.GetEnv("DB_HOST", "localhost"),
			Port:     utils.GetEnv("DB_PORT", "5432"),
			User:     utils.GetEnv("DB_USER", "postgres"),
			Password: utils.GetEnv("DB_PASSWORD", "postgres"),
			Name:     utils.GetEnv("DB_NAME", "ecommerce"),
			SSLMode:  utils.GetEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:                utils.GetEnv("JWT_SECRET", "your-secret-key"),
			ExpiresIn:             jwtExpiresIn,
			RefreshTokenExpiresIn: jwtRefreshTokenExpiresIn,
		},
		AWS: AWSConfig{
			Region:          utils.GetEnv("AWS_REGION", "us-east-1"),
			AccessKeyID:     utils.GetEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: utils.GetEnv("AWS_SECRET_ACCESS_KEY", ""),
			S3Bucket:        utils.GetEnv("AWS_S3_BUCKET_NAME", ""),
			S3Endpoint:      utils.GetEnv("AWS_S3_ENDPOINT", ""),
		},
		Upload: UploadConfig{
			Path:        utils.GetEnv("UPLOAD_DIR", "./uploads"),
			MaxFileSize: maxFileSize,
		},
	}
	return config, nil
}
