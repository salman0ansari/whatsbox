package stats

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/salman0ansari/whatsbox/internal/database"
	"github.com/salman0ansari/whatsbox/internal/logging"
	"go.uber.org/zap"
)

// Collector tracks real-time statistics for the service
type Collector struct {
	// Counters (use atomic for thread safety)
	uploadsTotal    int64
	downloadsTotal  int64
	bytesUploaded   int64
	bytesDownloaded int64

	// Gauges (current values)
	activeUploads   int64
	activeDownloads int64

	// Error counters
	uploadErrors   int64
	downloadErrors int64

	// Session stats (reset on restart)
	startTime time.Time

	// For persisting to database
	statsRepo *database.StatsRepository
	mu        sync.RWMutex
}

// Stats represents a snapshot of current statistics
type Stats struct {
	// Counters
	UploadsTotal    int64 `json:"uploads_total"`
	DownloadsTotal  int64 `json:"downloads_total"`
	BytesUploaded   int64 `json:"bytes_uploaded"`
	BytesDownloaded int64 `json:"bytes_downloaded"`

	// Gauges
	ActiveUploads   int64 `json:"active_uploads"`
	ActiveDownloads int64 `json:"active_downloads"`

	// Errors
	UploadErrors   int64 `json:"upload_errors"`
	DownloadErrors int64 `json:"download_errors"`

	// Session info
	UptimeSeconds int64  `json:"uptime_seconds"`
	StartTime     string `json:"start_time"`
}

// Global collector instance
var (
	collector *Collector
	once      sync.Once
)

// Init initializes the global stats collector
func Init() *Collector {
	once.Do(func() {
		collector = &Collector{
			startTime: time.Now(),
			statsRepo: database.NewStatsRepository(),
		}
		logging.Info("Stats collector initialized")
	})
	return collector
}

// Get returns the global collector instance
func Get() *Collector {
	if collector == nil {
		return Init()
	}
	return collector
}

// IncrementUploads increments the upload counter
func (c *Collector) IncrementUploads() {
	atomic.AddInt64(&c.uploadsTotal, 1)
}

// IncrementDownloads increments the download counter
func (c *Collector) IncrementDownloads() {
	atomic.AddInt64(&c.downloadsTotal, 1)
}

// AddBytesUploaded adds to the bytes uploaded counter
func (c *Collector) AddBytesUploaded(bytes int64) {
	atomic.AddInt64(&c.bytesUploaded, bytes)
}

// AddBytesDownloaded adds to the bytes downloaded counter
func (c *Collector) AddBytesDownloaded(bytes int64) {
	atomic.AddInt64(&c.bytesDownloaded, bytes)
}

// IncrementActiveUploads increments active upload count
func (c *Collector) IncrementActiveUploads() {
	atomic.AddInt64(&c.activeUploads, 1)
}

// DecrementActiveUploads decrements active upload count
func (c *Collector) DecrementActiveUploads() {
	atomic.AddInt64(&c.activeUploads, -1)
}

// IncrementActiveDownloads increments active download count
func (c *Collector) IncrementActiveDownloads() {
	atomic.AddInt64(&c.activeDownloads, 1)
}

// DecrementActiveDownloads decrements active download count
func (c *Collector) DecrementActiveDownloads() {
	atomic.AddInt64(&c.activeDownloads, -1)
}

// IncrementUploadErrors increments upload error count
func (c *Collector) IncrementUploadErrors() {
	atomic.AddInt64(&c.uploadErrors, 1)
}

// IncrementDownloadErrors increments download error count
func (c *Collector) IncrementDownloadErrors() {
	atomic.AddInt64(&c.downloadErrors, 1)
}

// GetStats returns a snapshot of current statistics
func (c *Collector) GetStats() *Stats {
	return &Stats{
		UploadsTotal:    atomic.LoadInt64(&c.uploadsTotal),
		DownloadsTotal:  atomic.LoadInt64(&c.downloadsTotal),
		BytesUploaded:   atomic.LoadInt64(&c.bytesUploaded),
		BytesDownloaded: atomic.LoadInt64(&c.bytesDownloaded),
		ActiveUploads:   atomic.LoadInt64(&c.activeUploads),
		ActiveDownloads: atomic.LoadInt64(&c.activeDownloads),
		UploadErrors:    atomic.LoadInt64(&c.uploadErrors),
		DownloadErrors:  atomic.LoadInt64(&c.downloadErrors),
		UptimeSeconds:   int64(time.Since(c.startTime).Seconds()),
		StartTime:       c.startTime.Format(time.RFC3339),
	}
}

// GetActiveTransfers returns total active transfers (uploads + downloads)
func (c *Collector) GetActiveTransfers() int64 {
	return atomic.LoadInt64(&c.activeUploads) + atomic.LoadInt64(&c.activeDownloads)
}

// FlushHourly persists hourly stats to the database
func (c *Collector) FlushHourly() error {
	stats := c.GetStats()
	hourStart := time.Now().Truncate(time.Hour)

	hourlyStats := &database.StatsHourly{
		Hour:            hourStart,
		Uploads:         stats.UploadsTotal,
		Downloads:       stats.DownloadsTotal,
		UploadBytes:     stats.BytesUploaded,
		DownloadBytes:   stats.BytesDownloaded,
		FailedUploads:   stats.UploadErrors,
		FailedDownloads: stats.DownloadErrors,
		Requests:        0, // Can be tracked via middleware if needed
	}

	if err := c.statsRepo.SaveHourly(hourlyStats); err != nil {
		logging.Error("Failed to save hourly stats", zap.Error(err))
		return err
	}

	logging.Debug("Hourly stats flushed", zap.Time("hour", hourStart))
	return nil
}

// Reset resets all counters (typically after flushing)
func (c *Collector) Reset() {
	atomic.StoreInt64(&c.uploadsTotal, 0)
	atomic.StoreInt64(&c.downloadsTotal, 0)
	atomic.StoreInt64(&c.bytesUploaded, 0)
	atomic.StoreInt64(&c.bytesDownloaded, 0)
	atomic.StoreInt64(&c.uploadErrors, 0)
	atomic.StoreInt64(&c.downloadErrors, 0)
	// Note: active counters are not reset as they represent current state
}
