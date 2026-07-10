package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewCreatesSchema(t *testing.T) {
	// Path points into a not-yet-existing directory to exercise MkdirAll.
	path := filepath.Join(t.TempDir(), "nested", "state.db")

	s, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = s.Close() }()

	if s.DB() == nil {
		t.Fatal("DB() returned nil")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("database file not created: %v", err)
	}

	wantTables := []string{
		"hosts",
		"projects",
		"snapshots",
		"discovery_snapshots",
		"project_policies",
		"upgrade_runs",
		"upgrade_run_events",
		"backups",
	}
	for _, tbl := range wantTables {
		var name string
		err := s.DB().QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", tbl,
		).Scan(&name)
		if err != nil {
			t.Errorf("expected migration to create table %q: %v", tbl, err)
		}
	}
}

func TestNewRequiresPath(t *testing.T) {
	if _, err := New(""); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestNewIsIdempotent(t *testing.T) {
	// Migrations use IF NOT EXISTS, so opening the same file twice must not fail.
	path := filepath.Join(t.TempDir(), "state.db")

	s1, err := New(path)
	if err != nil {
		t.Fatalf("first New: %v", err)
	}
	_ = s1.Close()

	s2, err := New(path)
	if err != nil {
		t.Fatalf("second New: %v", err)
	}
	_ = s2.Close()
}

func TestBeginTx(t *testing.T) {
	s, err := New(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = s.Close() }()

	tx, err := s.BeginTx(context.Background())
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
}

func TestCloseNilDBIsSafe(t *testing.T) {
	s := &Store{}
	if err := s.Close(); err != nil {
		t.Fatalf("Close with nil db should be a no-op: %v", err)
	}
}

func TestSplitSQL(t *testing.T) {
	got := splitSQL("A;  B ;;\n C;")
	want := []string{"A;", "B;", "C;"}

	if len(got) != len(want) {
		t.Fatalf("splitSQL returned %d statements (%v), want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("statement %d = %q, want %q", i, got[i], want[i])
		}
	}
}
