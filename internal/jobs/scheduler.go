package jobs

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/salman0ansari/whatsbox/internal/config"
	"github.com/salman0ansari/whatsbox/internal/database"
	"github.com/salman0ansari/whatsbox/internal/logging"
	"github.com/salman0ansari/whatsbox/internal/stats"
	"go.uber.org/zap"
)

// Scheduler manages background jobs
type Scheduler struct {
	cfg           *config.Config
	collector     *stats.Collector
	fileRepo      *database.FileRepository
	uploadRepo    *database.UploadRepository
	statsRepo     *database.StatsRepository
	accessLogRepo *database.AccessLogRepository

	stopCh  chan struct{}
	wg      sync.WaitGroup
	running bool
	mu      sync.Mutex
}

// NewScheduler creates a new job scheduler
func NewScheduler(cfg *config.Config) *Scheduler {
	return &Scheduler{
		cfg:           cfg,
		collector:     stats.Get(),
		fileRepo:      database.NewFileRepository(),
		uploadRepo:    database.NewUploadRepository(),
		statsRepo:     database.NewStatsRepository(),
		accessLogRepo: database.NewAccessLogRepository(),
		stopCh:        make(chan struct{}),
	}
}

// Start starts all background jobs
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	logging.Info("Starting background job scheduler")

	// Start individual job goroutines
	s.wg.Add(4)
	go s.runExpiredFilesJob()
	go s.runIncompleteUploadsJob()
	go s.runStatsAggregationJob()
	go s.runAccessLogCleanupJob()
}

// Stop gracefully stops all background jobs
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	logging.Info("Stopping background job scheduler")
	close(s.stopCh)
	s.wg.Wait()
	logging.Info("Background job scheduler stopped")
}

// runExpiredFilesJob marks expired files every hour
func (s *Scheduler) runExpiredFilesJob() {
	defer s.wg.Done()

	// Run immediately on startup
	s.markExpiredFiles()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.markExpiredFiles()
		}
	}
}

func (s *Scheduler) markExpiredFiles() {
	count, err := s.fileRepo.MarkExpired()
	if err != nil {
		logging.Error("Failed to mark expired files", zap.Error(err))
		return
	}
	if count > 0 {
		logging.Info("Marked expired files", zap.Int64("count", count))
	}
}

// runIncompleteUploadsJob cleans up incomplete uploads every 6 hours
func (s *Scheduler) runIncompleteUploadsJob() {
	defer s.wg.Done()

	// Run immediately on startup
	s.cleanIncompleteUploads()

	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.cleanIncompleteUploads()
		}
	}
}

func (s *Scheduler) cleanIncompleteUploads() {
	// Delete uploads older than 24 hours
	before := time.Now().Add(-24 * time.Hour)
	count, err := s.uploadRepo.DeleteOld(before)
	if err != nil {
		logging.Error("Failed to delete old uploads", zap.Error(err))
		return
	}
	if count > 0 {
		logging.Info("Deleted incomplete uploads", zap.Int64("count", count))
	}

	// Clean temp files that don't have corresponding upload records
	s.cleanOrphanedTempFiles()
}

func (s *Scheduler) cleanOrphanedTempFiles() {
	files, err := os.ReadDir(s.cfg.TempDir)
	if err != nil {
		if !os.IsNotExist(err) {
			logging.Error("Failed to read temp directory", zap.Error(err))
		}
		return
	}

	var cleaned int
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Get file info to check age
		info, err := file.Info()
		if err != nil {
			continue
		}

		// Skip files less than 24 hours old
		if time.Since(info.ModTime()) < 24*time.Hour {
			continue
		}

		// Remove orphaned temp file
		path := filepath.Join(s.cfg.TempDir, file.Name())
		if err := os.Remove(path); err != nil {
			logging.Error("Failed to remove orphaned temp file",
				zap.String("path", path),
				zap.Error(err))
			continue
		}
		cleaned++
	}

	if cleaned > 0 {
		logging.Info("Cleaned orphaned temp files", zap.Int("count", cleaned))
	}
}

// runStatsAggregationJob aggregates hourly stats to daily at midnight
func (s *Scheduler) runStatsAggregationJob() {
	defer s.wg.Done()

	// Flush current stats on startup
	s.collector.FlushHourly()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			// Flush stats before shutting down
			s.collector.FlushHourly()
			return
		case <-ticker.C:
			s.aggregateStats()
		}
	}
}

func (s *Scheduler) aggregateStats() {
	// Flush current hourly stats
	if err := s.collector.FlushHourly(); err != nil {
		logging.Error("Failed to flush hourly stats", zap.Error(err))
	}

	// Reset counters after flush
	s.collector.Reset()

	// At midnight, aggregate previous day's hourly stats to daily
	now := time.Now()
	if now.Hour() == 0 {
		yesterday := now.Add(-24 * time.Hour)
		if err := s.statsRepo.AggregateHourlyToDaily(yesterday); err != nil {
			logging.Error("Failed to aggregate hourly to daily stats", zap.Error(err))
		} else {
			logging.Info("Aggregated daily stats", zap.Time("date", yesterday.Truncate(24*time.Hour)))
		}

		// Clean up old hourly stats (keep 7 days)
		oldBefore := now.Add(-7 * 24 * time.Hour)
		if count, err := s.statsRepo.DeleteOldHourly(oldBefore); err != nil {
			logging.Error("Failed to delete old hourly stats", zap.Error(err))
		} else if count > 0 {
			logging.Info("Deleted old hourly stats", zap.Int64("count", count))
		}
	}
}

// runAccessLogCleanupJob cleans old access logs daily
func (s *Scheduler) runAccessLogCleanupJob() {
	defer s.wg.Done()

	// Run at startup
	s.cleanAccessLogs()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.cleanAccessLogs()
		}
	}
}

func (s *Scheduler) cleanAccessLogs() {
	// Keep access logs for 30 days
	before := time.Now().Add(-30 * 24 * time.Hour)
	count, err := s.accessLogRepo.DeleteOld(before)
	if err != nil {
		logging.Error("Failed to delete old access logs", zap.Error(err))
		return
	}
	if count > 0 {
		logging.Info("Deleted old access logs", zap.Int64("count", count))
	}
}
