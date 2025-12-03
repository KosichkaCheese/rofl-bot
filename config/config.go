package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DB_URL     string
	BOT_TOKEN  string
	JWT_SECRET string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	config := &Config{
		DB_URL:     os.Getenv("DATABASE_URL"),
		BOT_TOKEN:  os.Getenv("BOT_TOKEN"),
		JWT_SECRET: os.Getenv("JWT_SECRET"),
	}

	return config, nil
}
