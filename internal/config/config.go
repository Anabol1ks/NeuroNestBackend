package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	YandexClientID     string
	YandexClientSecret string
	YandexRedirectURL  string
	UploadsPath        string
	BaseURL            string
	IAMtoken           string
	CatalogID          string
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file", err)
	}
	YandexClientID = os.Getenv("YANDEX_CLIENT_ID")
	YandexClientSecret = os.Getenv("YANDEX_CLIENT_SECRET")
	YandexRedirectURL = os.Getenv("YANDEX_REDIRECT_URL")
	UploadsPath = os.Getenv("UPLOADS_PATH")
	BaseURL = os.Getenv("BASE_URL")
	IAMtoken = os.Getenv("IAM_TOKEN")
	CatalogID = os.Getenv("CATALOG_ID")
}
