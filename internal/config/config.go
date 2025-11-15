package config

import (
	"os"
)

/*

Application configuration for loading parameters from environment variables.
Config struct contains settings for database connection and server operation:
- DBHost, DBPort, DBUser, DBPassword, DBName - database connection parameters
- ServerPort - port for HTTP server
- LogLevel - logging level (e.g., debug, info, error)

Load() function creates a config by reading values from environment variables.

*/

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
	LogLevel   string
}

func Load() (*Config, error) {
	return &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		ServerPort: os.Getenv("SERVER_PORT"),
		LogLevel:   os.Getenv("LOG_LEVEL"),
	}, nil
}
