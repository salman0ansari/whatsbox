package database

import (
	"database/sql"
	"time"
)

// File represents a stored file
type File struct {
	ID            string
	Filename      string
	MimeType      string
	FileSize      int64
	FileHash      string
	Description   sql.NullString
	DirectPath    string
	MediaKey      []byte
	FileEncHash   []byte
	PasswordHash  sql.NullString
	MaxDownloads  sql.NullInt64
	DownloadCount int64
	CreatedAt     time.Time
	ExpiresAt     time.Time
	Status        string
}

// Upload represents an in-progress chunked upload
type Upload struct {
	ID        string
	Filename  sql.NullString
	FileSize  sql.NullInt64
	Offset    int64
	Metadata  sql.NullString
	CreatedAt time.Time
	UpdatedAt time.Time
}

// StatsHourly represents hourly aggregated stats
type StatsHourly struct {
	Hour            time.Time
	Uploads         int64
	Downloads       int64
	UploadBytes     int64
	DownloadBytes   int64
	FailedUploads   int64
	FailedDownloads int64
	Requests        int64
}

// StatsDaily represents daily aggregated stats
type StatsDaily struct {
	Date          time.Time
	Uploads       int64
	Downloads     int64
	UploadBytes   int64
	DownloadBytes int64
}

// AccessLog represents a file access log entry
type AccessLog struct {
	ID        int64
	FileID    string
	Action    string
	IPAddress sql.NullString
	UserAgent sql.NullString
	CreatedAt time.Time
}

// FileRepository handles file database operations
type FileRepository struct{}

func NewFileRepository() *FileRepository {
	return &FileRepository{}
}

// Create inserts a new file record
func (r *FileRepository) Create(f *File) error {
	_, err := DB.Exec(`
		INSERT INTO files (id, filename, mime_type, file_size, file_hash, description,
			direct_path, media_key, file_enc_hash, password_hash, max_downloads,
			download_count, created_at, expires_at, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		f.ID, f.Filename, f.MimeType, f.FileSize, f.FileHash, f.Description,
		f.DirectPath, f.MediaKey, f.FileEncHash, f.PasswordHash, f.MaxDownloads,
		f.DownloadCount, f.CreatedAt, f.ExpiresAt, f.Status)
	return err
}

// GetByID retrieves a file by its ID
func (r *FileRepository) GetByID(id string) (*File, error) {
	f := &File{}
	err := DB.QueryRow(`
		SELECT id, filename, mime_type, file_size, file_hash, description,
			direct_path, media_key, file_enc_hash, password_hash, max_downloads,
			download_count, created_at, expires_at, status
		FROM files WHERE id = ?`, id).Scan(
		&f.ID, &f.Filename, &f.MimeType, &f.FileSize, &f.FileHash, &f.Description,
		&f.DirectPath, &f.MediaKey, &f.FileEncHash, &f.PasswordHash, &f.MaxDownloads,
		&f.DownloadCount, &f.CreatedAt, &f.ExpiresAt, &f.Status)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// GetByHash retrieves an active file by its hash (for deduplication)
func (r *FileRepository) GetByHash(hash string) (*File, error) {
	f := &File{}
	err := DB.QueryRow(`
		SELECT id, filename, mime_type, file_size, file_hash, description,
			direct_path, media_key, file_enc_hash, password_hash, max_downloads,
			download_count, created_at, expires_at, status
		FROM files WHERE file_hash = ? AND status = 'active'`, hash).Scan(
		&f.ID, &f.Filename, &f.MimeType, &f.FileSize, &f.FileHash, &f.Description,
		&f.DirectPath, &f.MediaKey, &f.FileEncHash, &f.PasswordHash, &f.MaxDownloads,
		&f.DownloadCount, &f.CreatedAt, &f.ExpiresAt, &f.Status)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// List retrieves all files with pagination
func (r *FileRepository) List(limit, offset int) ([]*File, error) {
	rows, err := DB.Query(`
		SELECT id, filename, mime_type, file_size, file_hash, description,
			direct_path, media_key, file_enc_hash, password_hash, max_downloads,
			download_count, created_at, expires_at, status
		FROM files
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*File
	for rows.Next() {
		f := &File{}
		err := rows.Scan(
			&f.ID, &f.Filename, &f.MimeType, &f.FileSize, &f.FileHash, &f.Description,
			&f.DirectPath, &f.MediaKey, &f.FileEncHash, &f.PasswordHash, &f.MaxDownloads,
			&f.DownloadCount, &f.CreatedAt, &f.ExpiresAt, &f.Status)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

// IncrementDownloadCount increments the download counter
func (r *FileRepository) IncrementDownloadCount(id string) error {
	_, err := DB.Exec(`UPDATE files SET download_count = download_count + 1 WHERE id = ?`, id)
	return err
}

// UpdateStatus updates the file status
func (r *FileRepository) UpdateStatus(id, status string) error {
	_, err := DB.Exec(`UPDATE files SET status = ? WHERE id = ?`, status, id)
	return err
}

// Delete soft-deletes a file by setting status to 'deleted'
func (r *FileRepository) Delete(id string) error {
	return r.UpdateStatus(id, "deleted")
}

// MarkExpired marks all files past expiry as expired
func (r *FileRepository) MarkExpired() (int64, error) {
	result, err := DB.Exec(`
		UPDATE files SET status = 'expired'
		WHERE status = 'active' AND expires_at < CURRENT_TIMESTAMP`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Count returns total file count by status
func (r *FileRepository) Count(status string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM files`
	if status != "" {
		query += ` WHERE status = ?`
		err := DB.QueryRow(query, status).Scan(&count)
		return count, err
	}
	err := DB.QueryRow(query).Scan(&count)
	return count, err
}

// TotalSize returns total size of all active files
func (r *FileRepository) TotalSize() (int64, error) {
	var size sql.NullInt64
	err := DB.QueryRow(`SELECT SUM(file_size) FROM files WHERE status = 'active'`).Scan(&size)
	if err != nil {
		return 0, err
	}
	return size.Int64, nil
}

// UploadRepository handles upload database operations
type UploadRepository struct{}

func NewUploadRepository() *UploadRepository {
	return &UploadRepository{}
}

// Create inserts a new upload record
func (r *UploadRepository) Create(u *Upload) error {
	_, err := DB.Exec(`
		INSERT INTO uploads (id, filename, file_size, offset, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Filename, u.FileSize, u.Offset, u.Metadata, u.CreatedAt, u.UpdatedAt)
	return err
}

// GetByID retrieves an upload by its ID
func (r *UploadRepository) GetByID(id string) (*Upload, error) {
	u := &Upload{}
	err := DB.QueryRow(`
		SELECT id, filename, file_size, offset, metadata, created_at, updated_at
		FROM uploads WHERE id = ?`, id).Scan(
		&u.ID, &u.Filename, &u.FileSize, &u.Offset, &u.Metadata, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateOffset updates the upload offset
func (r *UploadRepository) UpdateOffset(id string, offset int64) error {
	_, err := DB.Exec(`
		UPDATE uploads SET offset = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, offset, id)
	return err
}

// Delete removes an upload record
func (r *UploadRepository) Delete(id string) error {
	_, err := DB.Exec(`DELETE FROM uploads WHERE id = ?`, id)
	return err
}

// DeleteOld removes uploads older than the given duration
func (r *UploadRepository) DeleteOld(before time.Time) (int64, error) {
	result, err := DB.Exec(`DELETE FROM uploads WHERE created_at < ?`, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// AccessLogRepository handles access log database operations
type AccessLogRepository struct{}

func NewAccessLogRepository() *AccessLogRepository {
	return &AccessLogRepository{}
}

// Create inserts a new access log entry
func (r *AccessLogRepository) Create(log *AccessLog) error {
	_, err := DB.Exec(`
		INSERT INTO access_log (file_id, action, ip_address, user_agent, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		log.FileID, log.Action, log.IPAddress, log.UserAgent, log.CreatedAt)
	return err
}

// CountByFileID counts access logs for a specific file
func (r *AccessLogRepository) CountByFileID(fileID string) (int64, error) {
	var count int64
	err := DB.QueryRow(`SELECT COUNT(*) FROM access_log WHERE file_id = ?`, fileID).Scan(&count)
	return count, err
}

// DeleteOld removes access logs older than the given time
func (r *AccessLogRepository) DeleteOld(before time.Time) (int64, error) {
	result, err := DB.Exec(`DELETE FROM access_log WHERE created_at < ?`, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// StatsRepository handles stats database operations
type StatsRepository struct{}

func NewStatsRepository() *StatsRepository {
	return &StatsRepository{}
}

// SaveHourly saves or updates hourly stats
func (r *StatsRepository) SaveHourly(s *StatsHourly) error {
	_, err := DB.Exec(`
		INSERT INTO stats_hourly (hour, uploads, downloads, upload_bytes, download_bytes, 
			failed_uploads, failed_downloads, requests)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(hour) DO UPDATE SET
			uploads = uploads + excluded.uploads,
			downloads = downloads + excluded.downloads,
			upload_bytes = upload_bytes + excluded.upload_bytes,
			download_bytes = download_bytes + excluded.download_bytes,
			failed_uploads = failed_uploads + excluded.failed_uploads,
			failed_downloads = failed_downloads + excluded.failed_downloads,
			requests = requests + excluded.requests`,
		s.Hour, s.Uploads, s.Downloads, s.UploadBytes, s.DownloadBytes,
		s.FailedUploads, s.FailedDownloads, s.Requests)
	return err
}

// GetHourlyStats retrieves hourly stats for a given time range
func (r *StatsRepository) GetHourlyStats(start, end time.Time) ([]*StatsHourly, error) {
	rows, err := DB.Query(`
		SELECT hour, uploads, downloads, upload_bytes, download_bytes, 
			failed_uploads, failed_downloads, requests
		FROM stats_hourly
		WHERE hour >= ? AND hour < ?
		ORDER BY hour DESC`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*StatsHourly
	for rows.Next() {
		s := &StatsHourly{}
		if err := rows.Scan(&s.Hour, &s.Uploads, &s.Downloads, &s.UploadBytes,
			&s.DownloadBytes, &s.FailedUploads, &s.FailedDownloads, &s.Requests); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// SaveDaily saves or updates daily stats
func (r *StatsRepository) SaveDaily(s *StatsDaily) error {
	_, err := DB.Exec(`
		INSERT INTO stats_daily (date, uploads, downloads, upload_bytes, download_bytes)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(date) DO UPDATE SET
			uploads = excluded.uploads,
			downloads = excluded.downloads,
			upload_bytes = excluded.upload_bytes,
			download_bytes = excluded.download_bytes`,
		s.Date, s.Uploads, s.Downloads, s.UploadBytes, s.DownloadBytes)
	return err
}

// GetDailyStats retrieves daily stats for a given time range
func (r *StatsRepository) GetDailyStats(start, end time.Time) ([]*StatsDaily, error) {
	rows, err := DB.Query(`
		SELECT date, uploads, downloads, upload_bytes, download_bytes
		FROM stats_daily
		WHERE date >= ? AND date < ?
		ORDER BY date DESC`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*StatsDaily
	for rows.Next() {
		s := &StatsDaily{}
		if err := rows.Scan(&s.Date, &s.Uploads, &s.Downloads, &s.UploadBytes, &s.DownloadBytes); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// AggregateHourlyToDaily aggregates hourly stats to daily
func (r *StatsRepository) AggregateHourlyToDaily(date time.Time) error {
	startOfDay := date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	_, err := DB.Exec(`
		INSERT INTO stats_daily (date, uploads, downloads, upload_bytes, download_bytes)
		SELECT ?, SUM(uploads), SUM(downloads), SUM(upload_bytes), SUM(download_bytes)
		FROM stats_hourly
		WHERE hour >= ? AND hour < ?
		ON CONFLICT(date) DO UPDATE SET
			uploads = excluded.uploads,
			downloads = excluded.downloads,
			upload_bytes = excluded.upload_bytes,
			download_bytes = excluded.download_bytes`,
		startOfDay, startOfDay, endOfDay)
	return err
}

// DeleteOldHourly removes hourly stats older than the given time
func (r *StatsRepository) DeleteOldHourly(before time.Time) (int64, error) {
	result, err := DB.Exec(`DELETE FROM stats_hourly WHERE hour < ?`, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
