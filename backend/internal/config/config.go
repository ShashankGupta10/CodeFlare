package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	ServerPort         int
	LogLevel           string
	JWTSecret          string
	CloudflareApiToken string
	CloudflareZoneId   string
	DeleteSecretPhrase string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading environment file")
	}

	value, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		fmt.Println("Error converting server port to int")
	}

	return &Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		ServerPort:         value,
		LogLevel:           os.Getenv("LOG_LEVEL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		CloudflareApiToken: os.Getenv("CLOUDFLARE_API_TOKEN"),
		CloudflareZoneId:   os.Getenv("CLOUDFLARE_ZONE_ID"),
		DeleteSecretPhrase: os.Getenv("DELETE_SECRET_PHRASE"),
	}
}
