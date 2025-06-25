package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDriver   string
	DBDSN      string
	ServerPort string

	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOUseSSL    bool
	MinIOBucket    string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	return &Config{
		DBDriver:       os.Getenv("DB_DRIVER"),
		DBDSN:          os.Getenv("DB_DSN"),
		ServerPort:     os.Getenv("SERVER_PORT"),
		MinIOEndpoint:  os.Getenv("MINIO_ENDPOINT"),
		MinIOAccessKey: os.Getenv("MINIO_ACCESS_KEY_ID"),
		MinIOSecretKey: os.Getenv("MINIO_SECRET_ACCESS_KEY"),
		MinIOUseSSL:    os.Getenv("MINIO_USE_SSL") == "true",
		MinIOBucket:    os.Getenv("MINIO_BUCKET"),
	}
}
