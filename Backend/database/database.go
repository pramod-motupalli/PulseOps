package database

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "database/db")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS monitors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		method TEXT NOT NULL DEFAULT 'GET',
		interval_seconds INTEGER NOT NULL DEFAULT 60,
		timeout_seconds INTEGER NOT NULL DEFAULT 5,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		CHECK (interval_seconds >= 5),
		CHECK (timeout_seconds >= 1)
	);

	CREATE TABLE IF NOT EXISTS check_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		monitor_id INTEGER NOT NULL,
		status_code INTEGER NOT NULL DEFAULT 0,
		response_time_ms INTEGER NOT NULL DEFAULT 0,
		is_up INTEGER NOT NULL DEFAULT 0,
		error_message TEXT,
		checked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_monitors_user_id
		ON monitors(user_id);

	CREATE INDEX IF NOT EXISTS idx_check_results_monitor_time
		ON check_results(monitor_id, checked_at);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}