package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            int
	BaseURL         string
	ShortCodeLength int
	DB              DBConfig
}

type DBConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

// LoadConfig loads configuration from environment (.env) variables
func LoadConfig() *Config {
	// try loading from parent dir
	parentEnv := filepath.Join("..", ".env")
	if err := godotenv.Load(parentEnv); err != nil {
		// fallback to current dir
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: .env file not found, using environment variables or defaults")
		}
	}

	port := getEnvAsInt("PORT", 8080)
	baseURL := getEnv("BASE_URL", fmt.Sprintf("http://localhost:%d", port))

	return &Config{
		Port:            port,
		BaseURL:         baseURL,
		ShortCodeLength: getEnvAsInt("SHORT_CODE_LENGTH", 7),
		DB: DBConfig{
			Host:     getEnv("DATABASE_HOST", "127.0.0.1"),
			Port:     getEnvAsInt("DATABASE_PORT", 3306),
			Database: getEnv("DATABASE_NAME", "urls"),
			User:     getEnv("DATABASE_USER", "url_shorten_service"),
			Password: getEnv("DATABASE_PASSWORD", "123"),
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}
	if intValue, err := strconv.Atoi(strValue); err == nil {
		return intValue
	}
	return defaultValue
}

// GetDSN returns the MySQL connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		c.DB.User,
		c.DB.Password,
		c.DB.Host,
		c.DB.Port,
		c.DB.Database,
	)
}
