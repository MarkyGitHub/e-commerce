package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	AppEnv string
	Server ServerConfig
	DB     DBConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
}

// DBConfig holds database configuration
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN returns the PostgreSQL connection string
func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// Load loads configuration from environment variables.
// If envFile is provided, it will load that file first.
// Environment variables take precedence over .env file values.
func Load(envFile string) (*Config, error) {
	// Load .env file if provided
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("failed to load env file %s: %w", envFile, err)
		}
	}

	return &Config{
		AppEnv: getEnv("APP_ENV", "local"),
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5434"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "example"),
			DBName:   getEnv("DB_NAME", "postgres"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}, nil
}

// LoadFromAppEnv loads the appropriate .env file based on APP_ENV environment variable
// or the provided default environment name.
func LoadFromAppEnv(configDir string, defaultEnv string) (*Config, error) {
	env := getEnv("APP_ENV", defaultEnv)
	envFile := fmt.Sprintf("%s/%s.env", configDir, env)

	// Check if file exists, if not just load from environment
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return Load("")
	}

	return Load(envFile)
}

// GetSessionTTLMinutes returns the session TTL in minutes
func GetSessionTTLMinutes() int {
	ttl := getEnv("SESSION_TTL_MINUTES", "60")
	if val, err := strconv.Atoi(ttl); err == nil {
		return val
	}
	return 60
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
