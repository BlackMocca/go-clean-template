package config

import (
	"os"

	"github.com/go-pg/pg/v9"
	"github.com/joho/godotenv"
)

type Config struct {
	PGORM *pg.DB
}

func init() {
	if err := godotenv.Load(); err != nil {
		panic("No .env file found")
	}
}

func NewConfig() *Config {
	return &Config{
		PGORM: NewPsqlConnection(),
	}
}

func (c *Config) GetEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}
