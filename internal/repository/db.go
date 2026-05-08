package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

const dbFilename = "data.db"

// DataDir returns the platform-appropriate directory for application data.
// macOS: ~/Library/Application Support/todo-app
// Windows: %APPDATA%\todo-app
// Linux: ~/.config/todo-app
func DataDir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		base = home
	}
	return filepath.Join(base, "todo-app")
}

// Open creates dataDir if needed, opens data.db inside it, applies PRAGMAs,
// runs the schema migration, and returns the ready-to-use *sql.DB.
func Open(dataDir string) (*sql.DB, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("repository.Open: create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, dbFilename)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("repository.Open: open db: %w", err)
	}

	// SQLite works best with a single writer; WAL allows concurrent reads.
	db.SetMaxOpenConns(1)

	if err := applyPragmas(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("repository.Open: pragmas: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("repository.Open: migrate: %w", err)
	}

	return db, nil
}

func applyPragmas(db *sql.DB) error {
	// journal_mode=WAL returns the active mode; we ignore the result and
	// verify it in tests.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("journal_mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("foreign_keys: %w", err)
	}
	return nil
}

const schema = `
CREATE TABLE IF NOT EXISTS tasks (
    id          TEXT PRIMARY KEY,
    date        TEXT NOT NULL,
    text        TEXT NOT NULL,
    done        INTEGER NOT NULL DEFAULT 0,
    position    INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_date ON tasks(date);
`

func migrate(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("schema: %w", err)
	}
	return nil
}
