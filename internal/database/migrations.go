package database

import (
	"github.com/salman0ansari/whatsbox/internal/logging"
	"go.uber.org/zap"
)

func migrate() error {
	migrations := []string{
		// Files table
		`CREATE TABLE IF NOT EXISTS files (
			id              TEXT PRIMARY KEY,
			filename        TEXT NOT NULL,
			mime_type       TEXT NOT NULL,
			file_size       INTEGER NOT NULL,
			file_hash       TEXT NOT NULL,
			description     TEXT,
			
			direct_path     TEXT NOT NULL,
			media_key       BLOB NOT NULL,
			file_enc_hash   BLOB NOT NULL,
			
			password_hash   TEXT,
			max_downloads   INTEGER,
			download_count  INTEGER DEFAULT 0,
			
			created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at      DATETIME NOT NULL,
			
			status          TEXT DEFAULT 'active'
		)`,

		// Indexes for files
		`CREATE INDEX IF NOT EXISTS idx_files_hash ON files(file_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_files_expires_at ON files(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_files_status ON files(status)`,

		// Chunked uploads tracking
		`CREATE TABLE IF NOT EXISTS uploads (
			id              TEXT PRIMARY KEY,
			filename        TEXT,
			file_size       INTEGER,
			offset          INTEGER DEFAULT 0,
			metadata        TEXT,
			created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Hourly stats
		`CREATE TABLE IF NOT EXISTS stats_hourly (
			hour            DATETIME PRIMARY KEY,
			uploads         INTEGER DEFAULT 0,
			downloads       INTEGER DEFAULT 0,
			upload_bytes    INTEGER DEFAULT 0,
			download_bytes  INTEGER DEFAULT 0,
			failed_uploads  INTEGER DEFAULT 0,
			failed_downloads INTEGER DEFAULT 0,
			requests        INTEGER DEFAULT 0
		)`,

		// Daily stats
		`CREATE TABLE IF NOT EXISTS stats_daily (
			date            DATE PRIMARY KEY,
			uploads         INTEGER DEFAULT 0,
			downloads       INTEGER DEFAULT 0,
			upload_bytes    INTEGER DEFAULT 0,
			download_bytes  INTEGER DEFAULT 0
		)`,

		// File access log
		`CREATE TABLE IF NOT EXISTS access_log (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id         TEXT NOT NULL,
			action          TEXT NOT NULL,
			ip_address      TEXT,
			user_agent      TEXT,
			created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Indexes for access_log
		`CREATE INDEX IF NOT EXISTS idx_access_log_file_id ON access_log(file_id)`,
		`CREATE INDEX IF NOT EXISTS idx_access_log_created_at ON access_log(created_at)`,
	}

	for _, migration := range migrations {
		if _, err := DB.Exec(migration); err != nil {
			logging.Error("Migration failed", zap.Error(err), zap.String("sql", migration))
			return err
		}
	}

	// Run column migrations separately with error handling
	if err := migrateColumns(); err != nil {
		return err
	}

	logging.Info("Database migrations completed successfully")
	return nil
}

// migrateColumns handles ALTER TABLE migrations that may fail if column already exists
func migrateColumns() error {
	// Check if file_sha256 column exists
	var colCount int
	err := DB.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('files') WHERE name = 'file_sha256'`).Scan(&colCount)
	if err != nil {
		logging.Error("Failed to check if file_sha256 column exists", zap.Error(err))
		return err
	}

	if colCount == 0 {
		// Column doesn't exist, add it
		_, err = DB.Exec(`ALTER TABLE files ADD COLUMN file_sha256 BLOB`)
		if err != nil {
			logging.Error("Failed to add file_sha256 column", zap.Error(err))
			return err
		}
		logging.Info("Added file_sha256 column to files table")
	} else {
		logging.Debug("file_sha256 column already exists, skipping migration")
	}

	return nil
}
