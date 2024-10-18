package config

import (
	"fmt"
	"os"
	"path/filepath"
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
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}

	environmentPath := filepath.Join(dir, "./../.env")

	err = godotenv.Load(environmentPath)
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
