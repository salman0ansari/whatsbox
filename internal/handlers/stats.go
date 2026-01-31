package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/salman0ansari/whatsbox/internal/database"
	"github.com/salman0ansari/whatsbox/internal/logging"
	"github.com/salman0ansari/whatsbox/internal/stats"
	"go.uber.org/zap"
)

// StatsHandler handles stats-related endpoints
type StatsHandler struct {
	collector *stats.Collector
	statsRepo *database.StatsRepository
	fileRepo  *database.FileRepository
}

// NewStatsHandler creates a new stats handler
func NewStatsHandler() *StatsHandler {
	return &StatsHandler{
		collector: stats.Get(),
		statsRepo: database.NewStatsRepository(),
		fileRepo:  database.NewFileRepository(),
	}
}

// GetStats returns current real-time statistics
func (h *StatsHandler) GetStats(c *fiber.Ctx) error {
	currentStats := h.collector.GetStats()

	// Get file counts from database
	totalFiles, _ := h.fileRepo.Count("")
	activeFiles, _ := h.fileRepo.Count("active")
	expiredFiles, _ := h.fileRepo.Count("expired")
	deletedFiles, _ := h.fileRepo.Count("deleted")
	totalSize, _ := h.fileRepo.TotalSize()

	return c.JSON(fiber.Map{
		"realtime": currentStats,
		"storage": fiber.Map{
			"total_files":   totalFiles,
			"active_files":  activeFiles,
			"expired_files": expiredFiles,
			"deleted_files": deletedFiles,
			"total_bytes":   totalSize,
		},
	})
}

// GetHourlyStats returns hourly statistics for the last N hours
func (h *StatsHandler) GetHourlyStats(c *fiber.Ctx) error {
	hours := c.QueryInt("hours", 24)
	if hours < 1 || hours > 168 { // Max 1 week
		hours = 24
	}

	end := time.Now().Truncate(time.Hour).Add(time.Hour)
	start := end.Add(-time.Duration(hours) * time.Hour)

	hourlyStats, err := h.statsRepo.GetHourlyStats(start, end)
	if err != nil {
		logging.Error("Failed to get hourly stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "stats_failed",
			"message": "Failed to retrieve hourly statistics",
		})
	}

	// Convert to response format
	data := make([]fiber.Map, len(hourlyStats))
	for i, s := range hourlyStats {
		data[i] = fiber.Map{
			"hour":             s.Hour.Format(time.RFC3339),
			"uploads":          s.Uploads,
			"downloads":        s.Downloads,
			"upload_bytes":     s.UploadBytes,
			"download_bytes":   s.DownloadBytes,
			"failed_uploads":   s.FailedUploads,
			"failed_downloads": s.FailedDownloads,
			"requests":         s.Requests,
		}
	}

	return c.JSON(fiber.Map{
		"period": fiber.Map{
			"start": start.Format(time.RFC3339),
			"end":   end.Format(time.RFC3339),
			"hours": hours,
		},
		"data": data,
	})
}

// GetDailyStats returns daily statistics for the last N days
func (h *StatsHandler) GetDailyStats(c *fiber.Ctx) error {
	days := c.QueryInt("days", 30)
	if days < 1 || days > 365 {
		days = 30
	}

	end := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
	start := end.Add(-time.Duration(days) * 24 * time.Hour)

	dailyStats, err := h.statsRepo.GetDailyStats(start, end)
	if err != nil {
		logging.Error("Failed to get daily stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "stats_failed",
			"message": "Failed to retrieve daily statistics",
		})
	}

	// Convert to response format
	data := make([]fiber.Map, len(dailyStats))
	for i, s := range dailyStats {
		data[i] = fiber.Map{
			"date":           s.Date.Format("2006-01-02"),
			"uploads":        s.Uploads,
			"downloads":      s.Downloads,
			"upload_bytes":   s.UploadBytes,
			"download_bytes": s.DownloadBytes,
		}
	}

	return c.JSON(fiber.Map{
		"period": fiber.Map{
			"start": start.Format("2006-01-02"),
			"end":   end.Format("2006-01-02"),
			"days":  days,
		},
		"data": data,
	})
}
