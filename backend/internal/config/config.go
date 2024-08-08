package config

import "os"

type Config struct {
	Port      string
	JWTSecret string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "shhhhh"),
	}
}

func getEnv(key, def string) string {
	if val, yes := os.LookupEnv(key); yes {
		return val
	}
	return def
}
