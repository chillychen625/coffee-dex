package storage

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// SQLiteDB wraps a *sql.DB opened against a SQLite file.
type SQLiteDB struct {
	db *sql.DB
}

// NewSQLiteDB opens (or creates) a SQLite database at path, configures WAL
// mode and foreign keys, then creates all application tables.
func NewSQLiteDB(path string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	s := &SQLiteDB{db: db}

	if err := s.init(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize SQLite database: %w", err)
	}

	return s, nil
}

// DB returns the underlying *sql.DB.
func (s *SQLiteDB) DB() *sql.DB {
	return s.db
}

// Close closes the underlying database connection.
func (s *SQLiteDB) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SQLiteDB) init() error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
	}
	for _, p := range pragmas {
		if _, err := s.db.Exec(p); err != nil {
			return fmt.Errorf("failed to set pragma %q: %w", p, err)
		}
	}

	ddl := []string{
		`CREATE TABLE IF NOT EXISTS coffees (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			origin TEXT,
			roaster TEXT,
			variety TEXT,
			roast_level TEXT,
			processing_method TEXT,
			roast_date TEXT,
			is_finished INTEGER DEFAULT 0,
			created_at TEXT,
			updated_at TEXT
		)`,

		`CREATE TABLE IF NOT EXISTS brews (
			id TEXT PRIMARY KEY,
			coffee_id TEXT NOT NULL,
			tasting_notes TEXT,
			tasting_traits TEXT,
			rating INTEGER,
			recipe TEXT,
			dripper TEXT,
			end_time_minutes INTEGER,
			end_time_seconds INTEGER,
			created_at TEXT,
			FOREIGN KEY (coffee_id) REFERENCES coffees(id) ON DELETE CASCADE
		)`,

		`CREATE INDEX IF NOT EXISTS idx_brews_coffee_id ON brews(coffee_id)`,

		`CREATE TABLE IF NOT EXISTS brewers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			pokeball_type TEXT NOT NULL,
			recipes TEXT,
			created_at TEXT
		)`,

		`CREATE TABLE IF NOT EXISTS pokemons (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			sprite_path TEXT NOT NULL,
			base_stats TEXT NOT NULL,
			description TEXT
		)`,

		`CREATE TABLE IF NOT EXISTS coffee_pokemon (
			id TEXT PRIMARY KEY,
			coffee_id TEXT NOT NULL,
			pokemon_id INTEGER NOT NULL,
			nickname TEXT,
			level INTEGER DEFAULT 1,
			mapping_confidence REAL,
			llm_description TEXT,
			trait_mapping TEXT,
			created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
			FOREIGN KEY (coffee_id) REFERENCES coffees(id),
			FOREIGN KEY (pokemon_id) REFERENCES pokemons(id),
			UNIQUE (pokemon_id)
		)`,
	}

	for _, stmt := range ddl {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute DDL: %w\nStatement: %s", err, stmt)
		}
	}

	// Safe migrations — ignore errors if column already exists.
	s.db.Exec("ALTER TABLE brews ADD COLUMN is_learning INTEGER DEFAULT 0")
	s.db.Exec("ALTER TABLE coffees ADD COLUMN finished_at TEXT")

	return nil
}
