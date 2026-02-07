package handlers

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/salman0ansari/whatsbox/internal/config"
	"github.com/salman0ansari/whatsbox/internal/database"
	"github.com/salman0ansari/whatsbox/internal/logging"
	"github.com/salman0ansari/whatsbox/internal/utils"
	"github.com/salman0ansari/whatsbox/internal/whatsapp"
	"go.uber.org/zap"
)

// FileHandler handles file-related endpoints
type FileHandler struct {
	waClient *whatsapp.Client
	fileRepo *database.FileRepository
	logRepo  *database.AccessLogRepository
	cfg      *config.Config
}

// NewFileHandler creates a new file handler
func NewFileHandler(waClient *whatsapp.Client, cfg *config.Config) *FileHandler {
	return &FileHandler{
		waClient: waClient,
		fileRepo: database.NewFileRepository(),
		logRepo:  database.NewAccessLogRepository(),
		cfg:      cfg,
	}
}

// FileResponse represents a file in API responses
type FileResponse struct {
	ID                string    `json:"id"`
	Filename          string    `json:"filename"`
	MimeType          string    `json:"mime_type"`
	FileSize          int64     `json:"file_size"`
	Description       string    `json:"description,omitempty"`
	DownloadURL       string    `json:"download_url"`
	PasswordProtected bool      `json:"password_protected"`
	MaxDownloads      *int64    `json:"max_downloads,omitempty"`
	DownloadCount     int64     `json:"download_count"`
	CreatedAt         time.Time `json:"created_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	Status            string    `json:"status"`
	Duplicate         bool      `json:"duplicate,omitempty"`
}

// Upload handles file uploads
func (h *FileHandler) Upload(c *fiber.Ctx) error {
	// Check WhatsApp connection
	if !h.waClient.IsConnected() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":   "whatsapp_not_connected",
			"message": "WhatsApp is not connected. Please scan QR code first.",
		})
	}

	// Get file from form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "missing_file",
			"message": "No file provided",
		})
	}

	// Sanitize filename to prevent path traversal
	fileHeader.Filename = utils.SanitizeFilename(fileHeader.Filename)

	// Check file size
	if fileHeader.Size > h.cfg.MaxUploadSize {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"error":   "file_too_large",
			"message": "File exceeds maximum upload size of 2GB",
		})
	}

	// Open file
	file, err := fileHeader.Open()
	if err != nil {
		logging.Error("Failed to open uploaded file", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "file_open_failed",
			"message": "Failed to open uploaded file",
		})
	}
	defer file.Close()

	// Read file content
	fileData, err := io.ReadAll(file)
	if err != nil {
		logging.Error("Failed to read uploaded file", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "file_read_failed",
			"message": "Failed to read uploaded file",
		})
	}

	// Calculate file hash for tracking purposes
	fileHash := utils.HashFile(fileData)

	// Get optional metadata from form
	description := c.FormValue("description", "")
	password := c.FormValue("password", "")
	maxDownloadsStr := c.FormValue("max_downloads", "")
	expiresInStr := c.FormValue("expires_in", "")

	// Parse max downloads
	var maxDownloads sql.NullInt64
	if maxDownloadsStr != "" {
		if val, err := strconv.ParseInt(maxDownloadsStr, 10, 64); err == nil && val > 0 {
			maxDownloads = sql.NullInt64{Int64: val, Valid: true}
		}
	}

	// Calculate expiry time
	expiryDays := h.cfg.DefaultExpiryDays
	if expiresInStr != "" {
		if seconds, err := strconv.ParseInt(expiresInStr, 10, 64); err == nil && seconds > 0 {
			days := int(seconds / 86400)
			if days > 0 && days <= h.cfg.MaxExpiryDays {
				expiryDays = days
			}
		}
	}
	expiresAt := time.Now().Add(time.Duration(expiryDays) * 24 * time.Hour)

	// Hash password if provided
	var passwordHash sql.NullString
	if password != "" {
		hash, err := utils.HashPassword(password)
		if err != nil {
			logging.Error("Failed to hash password", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "password_hash_failed",
				"message": "Failed to process password",
			})
		}
		passwordHash = sql.NullString{String: hash, Valid: true}
	}

	// Detect MIME type
	mimeType := http.DetectContentType(fileData)
	if mimeType == "application/octet-stream" {
		// Try to use the content type from the form
		mimeType = fileHeader.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}

	// Get correct media type for WhatsApp
	mediaType := utils.GetMediaType(mimeType)

	// Upload to WhatsApp
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Minute)
	defer cancel()

	uploadResp, err := h.waClient.Upload(ctx, fileData, mediaType)
	if err != nil {
		logging.Error("Failed to upload to WhatsApp", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "upload_failed",
			"message": "Failed to upload file to storage",
		})
	}

	// Generate short ID
	fileID, err := utils.GenerateShortID(h.cfg.ShortIDLength)
	if err != nil {
		logging.Error("Failed to generate file ID", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "id_generation_failed",
			"message": "Failed to generate file ID",
		})
	}

	// Create file record
	dbFile := &database.File{
		ID:            fileID,
		Filename:      fileHeader.Filename,
		MimeType:      mimeType,
		FileSize:      fileHeader.Size,
		FileHash:      fileHash,
		Description:   sql.NullString{String: description, Valid: description != ""},
		DirectPath:    uploadResp.DirectPath,
		MediaKey:      uploadResp.MediaKey,
		FileEncHash:   uploadResp.FileEncHash,
		FileSHA256:    uploadResp.FileSHA256,
		PasswordHash:  passwordHash,
		MaxDownloads:  maxDownloads,
		DownloadCount: 0,
		CreatedAt:     time.Now(),
		ExpiresAt:     expiresAt,
		Status:        "active",
	}

	if err := h.fileRepo.Create(dbFile); err != nil {
		logging.Error("Failed to save file record", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "save_failed",
			"message": "Failed to save file record",
		})
	}

	logging.Info("File uploaded successfully",
		zap.String("file_id", fileID),
		zap.String("filename", fileHeader.Filename),
		zap.Int64("size", fileHeader.Size),
	)

	return c.Status(fiber.StatusCreated).JSON(h.toFileResponse(dbFile, false))
}

// List returns all files
func (h *FileHandler) List(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)

	if limit > 1000 {
		limit = 1000
	}

	files, err := h.fileRepo.List(limit, offset)
	if err != nil {
		logging.Error("Failed to list files", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "list_failed",
			"message": "Failed to list files",
		})
	}

	responses := make([]FileResponse, len(files))
	for i, f := range files {
		responses[i] = h.toFileResponse(f, false)
	}

	return c.JSON(fiber.Map{
		"files":  responses,
		"limit":  limit,
		"offset": offset,
		"count":  len(responses),
	})
}

// Get returns a single file's metadata
func (h *FileHandler) Get(c *fiber.Ctx) error {
	fileID := c.Params("id")
	if fileID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "missing_id",
			"message": "File ID is required",
		})
	}

	file, err := h.fileRepo.GetByID(fileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "not_found",
				"message": "File not found",
			})
		}
		logging.Error("Failed to get file", zap.Error(err), zap.String("file_id", fileID))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "get_failed",
			"message": "Failed to get file",
		})
	}

	return c.JSON(h.toFileResponse(file, false))
}

// Download handles file downloads
func (h *FileHandler) Download(c *fiber.Ctx) error {
	fileID := c.Params("id")
	if fileID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "missing_id",
			"message": "File ID is required",
		})
	}

	// Get file metadata
	file, err := h.fileRepo.GetByID(fileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "not_found",
				"message": "File not found",
			})
		}
		logging.Error("Failed to get file", zap.Error(err), zap.String("file_id", fileID))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "get_failed",
			"message": "Failed to get file",
		})
	}

	// Check if file is expired
	if file.Status == "expired" || time.Now().After(file.ExpiresAt) {
		return c.Status(fiber.StatusGone).JSON(fiber.Map{
			"error":      "file_expired",
			"message":    "This file has expired and is no longer available",
			"expired_at": file.ExpiresAt,
		})
	}

	// Check if file is deleted
	if file.Status == "deleted" {
		return c.Status(fiber.StatusGone).JSON(fiber.Map{
			"error":   "file_deleted",
			"message": "This file has been deleted",
		})
	}

	// Check download limit - will be validated atomically during download

	// Check password if required
	if file.PasswordHash.Valid {
		password := c.Get("X-Password", "")
		if password == "" {
			password = c.Query("password", "")
		}
		if password == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "password_required",
				"message": "This file is password protected. Provide password via X-Password header or password query parameter.",
			})
		}
		if !utils.CheckPassword(password, file.PasswordHash.String) {
			// Log failed attempt
			h.logRepo.Create(&database.AccessLog{
				FileID:    fileID,
				Action:    "password_fail",
				IPAddress: sql.NullString{String: c.IP(), Valid: true},
				UserAgent: sql.NullString{String: c.Get("User-Agent"), Valid: true},
				CreatedAt: time.Now(),
			})
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "invalid_password",
				"message": "Incorrect password",
			})
		}
	}

	// Check WhatsApp connection
	if !h.waClient.IsConnected() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":   "whatsapp_not_connected",
			"message": "WhatsApp is not connected. Cannot download file.",
		})
	}

	// Download from WhatsApp
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Minute)
	defer cancel()

	// Get correct media type for WhatsApp
	mediaType := utils.GetMediaType(file.MimeType)

	downloadReq := &whatsapp.DownloadRequest{
		DirectPath:  file.DirectPath,
		MediaKey:    file.MediaKey,
		FileEncHash: file.FileEncHash,
		FileSHA256:  file.FileSHA256,
		FileLength:  uint64(file.FileSize),
		MediaType:   mediaType,
	}

	data, err := h.waClient.Download(ctx, downloadReq)
	if err != nil {
		logging.Error("Failed to download from WhatsApp", zap.Error(err), zap.String("file_id", fileID))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "download_failed",
			"message": "Failed to download file from storage",
		})
	}

	// Increment download count atomically
	if err := h.fileRepo.IncrementDownloadCountAtomically(fileID); err != nil {
		if err.Error() == "download limit reached" {
			return c.Status(fiber.StatusGone).JSON(fiber.Map{
				"error":         "download_limit_reached",
				"message":       "This file has reached its maximum download count",
				"max_downloads": file.MaxDownloads.Int64,
			})
		}
		logging.Warn("Failed to increment download count", zap.Error(err), zap.String("file_id", fileID))
	}

	// Log access
	h.logRepo.Create(&database.AccessLog{
		FileID:    fileID,
		Action:    "download",
		IPAddress: sql.NullString{String: c.IP(), Valid: true},
		UserAgent: sql.NullString{String: c.Get("User-Agent"), Valid: true},
		CreatedAt: time.Now(),
	})

	logging.Info("File downloaded",
		zap.String("file_id", fileID),
		zap.String("ip", c.IP()),
	)

	// Set headers and return file
	c.Set("Content-Type", file.MimeType)
	c.Set("Content-Disposition", "attachment; filename=\""+file.Filename+"\"")
	c.Set("Content-Length", strconv.FormatInt(file.FileSize, 10))

	return c.Send(data)
}

// Delete soft-deletes a file
func (h *FileHandler) Delete(c *fiber.Ctx) error {
	fileID := c.Params("id")
	if fileID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "missing_id",
			"message": "File ID is required",
		})
	}

	// Check if file exists
	file, err := h.fileRepo.GetByID(fileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "not_found",
				"message": "File not found",
			})
		}
		logging.Error("Failed to get file", zap.Error(err), zap.String("file_id", fileID))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "get_failed",
			"message": "Failed to get file",
		})
	}

	if file.Status == "deleted" {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":   "already_deleted",
			"message": "File has already been deleted",
		})
	}

	if err := h.fileRepo.Delete(fileID); err != nil {
		logging.Error("Failed to delete file", zap.Error(err), zap.String("file_id", fileID))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "delete_failed",
			"message": "Failed to delete file",
		})
	}

	logging.Info("File deleted", zap.String("file_id", fileID))

	return c.JSON(fiber.Map{
		"message": "File deleted successfully",
		"id":      fileID,
	})
}

// toFileResponse converts a database file to an API response
func (h *FileHandler) toFileResponse(f *database.File, duplicate bool) FileResponse {
	resp := FileResponse{
		ID:                f.ID,
		Filename:          f.Filename,
		MimeType:          f.MimeType,
		FileSize:          f.FileSize,
		DownloadURL:       "/api/files/" + f.ID + "/download",
		PasswordProtected: f.PasswordHash.Valid,
		DownloadCount:     f.DownloadCount,
		CreatedAt:         f.CreatedAt,
		ExpiresAt:         f.ExpiresAt,
		Status:            f.Status,
		Duplicate:         duplicate,
	}

	if f.Description.Valid {
		resp.Description = f.Description.String
	}

	if f.MaxDownloads.Valid {
		resp.MaxDownloads = &f.MaxDownloads.Int64
	}

	return resp
}
