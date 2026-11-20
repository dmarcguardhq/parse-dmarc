//go:build cgo

package storage

import (
	"database/sql"
	"fmt"

	sqlite3 "github.com/mattn/go-sqlite3"
)

// sqliteDriver is go-sqlite3 with temp_store set on every connection: the
// driver has no DSN key for it. Temporary tables and sort spills stay in
// memory, so nothing is written outside the database's directory and the
// root filesystem can be read-only.
const sqliteDriver = "sqlite3_parse_dmarc"

// database/sql drivers register by name, once, at program start.
func init() {
	sql.Register(sqliteDriver, &sqlite3.SQLiteDriver{
		ConnectHook: func(c *sqlite3.SQLiteConn) error {
			_, err := c.Exec("PRAGMA temp_store = MEMORY", nil)
			return err
		},
	})
}

// NewStorage creates a new storage instance
func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open(sqliteDriver, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &Storage{db: db}
	if err := storage.init(); err != nil {
		return nil, fmt.Errorf("initialize database schema: %w", err)
	}

	return storage, nil
}

// init initializes database schema
func (s *Storage) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		report_id TEXT UNIQUE NOT NULL,
		org_name TEXT NOT NULL,
		email TEXT,
		domain TEXT NOT NULL,
		date_begin INTEGER NOT NULL,
		date_end INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		policy_p TEXT,
		policy_sp TEXT,
		policy_pct INTEGER,
		total_messages INTEGER,
		compliant_messages INTEGER,
		raw_report TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		report_id INTEGER NOT NULL,
		source_ip TEXT NOT NULL,
		count INTEGER NOT NULL,
		disposition TEXT,
		dkim_result TEXT,
		spf_result TEXT,
		header_from TEXT,
		envelope_from TEXT,
		dkim_domains TEXT,
		spf_domains TEXT,
		FOREIGN KEY (report_id) REFERENCES reports(id)
	);

	CREATE INDEX IF NOT EXISTS idx_reports_date_begin ON reports(date_begin);
	CREATE INDEX IF NOT EXISTS idx_reports_domain ON reports(domain);
	CREATE INDEX IF NOT EXISTS idx_records_report_id ON records(report_id);
	CREATE INDEX IF NOT EXISTS idx_records_source_ip ON records(source_ip);
	`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}

	return nil
}
