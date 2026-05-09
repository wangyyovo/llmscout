package storage

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS routes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		type TEXT NOT NULL CHECK(type IN ('prefix','exact')),
		path TEXT NOT NULL,
		target_url TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		route_name TEXT NOT NULL,
		method TEXT NOT NULL,
		path TEXT NOT NULL,
		request_url TEXT NOT NULL DEFAULT '',
		target_url TEXT NOT NULL DEFAULT '',
		protocol TEXT NOT NULL DEFAULT 'REST',
		status_code INTEGER,
		latency_ms INTEGER,
		req_headers TEXT,
		req_body TEXT,
		resp_headers TEXT,
		resp_body TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_logs_route_name ON logs(route_name);
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);`
	_, err := db.Exec(schema)
	if err != nil {
		return err
	}
	// Migration: add new columns to existing logs table
	db.Exec("ALTER TABLE logs ADD COLUMN request_url TEXT NOT NULL DEFAULT ''")
	db.Exec("ALTER TABLE logs ADD COLUMN target_url TEXT NOT NULL DEFAULT ''")
	return nil
}
