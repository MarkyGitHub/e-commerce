package testutil

import (
	"os"
	"testing"
)

func TestGetEnv_DefaultAndOverride(t *testing.T) {
	key := "TESTUTIL_GETENV_TEST"
	_ = os.Unsetenv(key)

	if got := getEnv(key, "default"); got != "default" {
		t.Fatalf("expected default value, got %q", got)
	}

	t.Setenv(key, "override")
	if got := getEnv(key, "default"); got != "override" {
		t.Fatalf("expected override value, got %q", got)
	}
}

func TestSetupTestDB_SQLiteDefault(t *testing.T) {
	// Ensure we exercise the default (SQLite) code path.
	_ = os.Unsetenv("TEST_DB_DIALECT")

	db := SetupTestDB(t)
	if db == nil {
		t.Fatalf("expected db to be non-nil")
	}

	// A very small smoke check that the connection is usable.
	if err := db.Exec("SELECT 1").Error; err != nil {
		t.Fatalf("expected sqlite db to accept queries: %v", err)
	}
}

func TestSetupTestDB_PostgresMode_SkipsIfUnavailable(t *testing.T) {
	t.Setenv("TEST_DB_DIALECT", "postgres")

	// If Postgres is not available, this will call t.Skipf.
	// That's OK: it still exercises the branch and doesn't fail the suite.
	_ = SetupTestDB(t)
}
