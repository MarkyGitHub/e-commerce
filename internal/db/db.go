package db

import (
	"fmt"

	"github.com/MarkyGitHub/e-commerce/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect stellt eine Verbindung zur PostgreSQL-Datenbank her
func Connect(cfg config.DBConfig) (*gorm.DB, error) {
	return connect(postgres.Open(cfg.DSN()))
}

// ConnectWithDSN stellt eine Verbindung mit DSN-String her
func ConnectWithDSN(dsn string) (*gorm.DB, error) {
	return connect(postgres.Open(dsn))
}

func connect(dialector gorm.Dialector) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
