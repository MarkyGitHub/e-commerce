package db

import (
	"testing"

	"github.com/glebarez/sqlite"
)

func TestConnect_HelperWithSQLiteDialector(t *testing.T) {
	gormDB, err := connect(sqlite.Open("file:db_connect_test?mode=memory&cache=shared"))
	if err != nil {
		t.Fatalf("expected connect() to succeed with sqlite dialector: %v", err)
	}

	if err := gormDB.Exec("SELECT 1").Error; err != nil {
		t.Fatalf("expected db to accept queries: %v", err)
	}
}
