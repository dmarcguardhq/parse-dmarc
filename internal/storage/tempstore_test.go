package storage

import (
	"path/filepath"
	"testing"
)

// Temporary data stays in memory, so the process writes nothing outside the
// database directory.
func TestSQLiteTempStoreIsMemory(t *testing.T) {
	s, err := NewStorage(filepath.Join(t.TempDir(), "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	var mode int
	if err := s.db.QueryRow("PRAGMA temp_store").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != 2 {
		t.Fatalf("temp_store = %d, want 2 (MEMORY)", mode)
	}
}
