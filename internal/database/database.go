package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/salman0ansari/whatsbox/internal/config"
)

var DB *sql.DB

func Setup(cfg *config.Config) error {
	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.DatabasePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return err
	}

	var err error
	DB, err = sql.Open("sqlite3", cfg.DatabasePath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return err
	}

	// Test connection
	if err := DB.Ping(); err != nil {
		return err
	}

	// Run migrations
	if err := migrate(); err != nil {
		return err
	}

	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
