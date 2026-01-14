package testutil

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSetupTestDB_PostgresMode_CanBeExercisedWithoutPostgres(t *testing.T) {
	// Force postgres path.
	t.Setenv("TEST_DB_DIALECT", "postgres")

	old := openPostgres
	t.Cleanup(func() { openPostgres = old })

	openPostgres = func(string) (*gorm.DB, error) {
		// Return a sqlite DB instead so we exercise the Postgres branch (including cleanup)
		// without requiring an external Postgres instance.
		return gorm.Open(sqlite.Open("file:testutil_postgres_path?mode=memory&cache=shared"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	}

	db := SetupTestDB(t)
	if db == nil {
		t.Fatalf("expected db")
	}
	if err := db.Exec("SELECT 1").Error; err != nil {
		t.Fatalf("expected db to accept queries: %v", err)
	}
}
