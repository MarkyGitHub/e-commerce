package db

import (
	"testing"

	"github.com/glebarez/sqlite"
)

func TestMigrate_SQLite(t *testing.T) {
	gormDB, err := connect(sqlite.Open("file:db_migrate_test?mode=memory&cache=shared"))
	if err != nil {
		t.Fatalf("connect() error: %v", err)
	}

	if err := Migrate(gormDB); err != nil {
		t.Fatalf("Migrate() error: %v", err)
	}
}
