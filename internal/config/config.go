package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config enthält die gesamte Anwendungskonfiguration
type Config struct {
	AppEnv string
	Server ServerConfig
	DB     DBConfig
}

// ServerConfig enthält Server-Konfiguration
type ServerConfig struct {
	Port string
}

// DBConfig enthält Datenbank-Konfiguration
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN gibt den PostgreSQL-Verbindungsstring zurück
func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// Load lädt Konfiguration aus Umgebungsvariablen.
// Falls envFile angegeben, wird diese Datei zuerst geladen.
// Umgebungsvariablen haben Vorrang vor .env-Werten.
// In prod muss DB_PASSWORD gesetzt sein (Umgebung oder Secret Manager) —
// es gibt keinen Default und keinen Commit von Produktionsgeheimnissen.
func Load(envFile string) (*Config, error) {
	// .env-Datei laden falls angegeben
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("failed to load env file %s: %w", envFile, err)
		}
	}

	appEnv := getEnv("APP_ENV", "local")
	passwordDefault := "example"
	if appEnv == "prod" {
		passwordDefault = ""
	}

	cfg := &Config{
		AppEnv: appEnv,
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5434"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", passwordDefault),
			DBName:   getEnv("DB_NAME", "postgres"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.AppEnv == "prod" && c.DB.Password == "" {
		return fmt.Errorf("DB_PASSWORD must be set via environment or secret manager in production")
	}
	return nil
}

// LoadFromAppEnv lädt die passende .env-Datei basierend auf APP_ENV
// oder dem angegebenen Standard-Umgebungsnamen.
func LoadFromAppEnv(configDir string, defaultEnv string) (*Config, error) {
	env := getEnv("APP_ENV", defaultEnv)
	envFile := fmt.Sprintf("%s/%s.env", configDir, env)

	// Prüfen ob Datei existiert, sonst aus Umgebung laden
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return Load("")
	}

	return Load(envFile)
}

// GetSessionTTLMinutes gibt die Session-Gültigkeitsdauer in Minuten zurück
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
