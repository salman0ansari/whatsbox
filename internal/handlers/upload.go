package handlers

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/salman0ansari/whatsbox/internal/config"
	"github.com/salman0ansari/whatsbox/internal/database"
	"github.com/salman0ansari/whatsbox/internal/logging"
	"github.com/salman0ansari/whatsbox/internal/utils"
	"github.com/salman0ansari/whatsbox/internal/whatsapp"
	"go.uber.org/zap"
)

const (
	tusVersion    = "1.0.0"
	tusExtensions = "creation,termination"
)

// TusHandler handles chunked uploads using the tus protocol
type TusHandler struct {
	waClient   *whatsapp.Client
	uploadRepo *database.UploadRepository
	fileRepo   *database.FileRepository
	cfg        *config.Config
}

// NewTusHandler creates a new tus handler
func NewTusHandler(waClient *whatsapp.Client, cfg *config.Config) *TusHandler {
	// Ensure temp directory exists
	os.MkdirAll(cfg.TempDir, 0755)

	return &TusHandler{
		waClient:   waClient,
		uploadRepo: database.NewUploadRepository(),
		fileRepo:   database.NewFileRepository(),
		cfg:        cfg,
	}
}

// Options handles the OPTIONS request for tus protocol discovery
func (h *TusHandler) Options(c *fiber.Ctx) error {
	c.Set("Tus-Resumable", tusVersion)
	c.Set("Tus-Version", tusVersion)
	c.Set("Tus-Extension", tusExtensions)
	c.Set("Tus-Max-Size", strconv.FormatInt(h.cfg.MaxUploadSize, 10))
	return c.SendStatus(fiber.StatusNoContent)
}

// Create handles POST requests to create a new upload
func (h *TusHandler) Create(c *fiber.Ctx) error {
	// Verify tus version
	if c.Get("Tus-Resumable") != tusVersion {
		return c.Status(fiber.StatusPreconditionFailed).JSON(fiber.Map{
			"error":   "unsupported_version",
			"message": "Unsupported Tus-Resumable version",
		})
	}

	// Get upload length
	uploadLength, err := strconv.ParseInt(c.Get("Upload-Length"), 10, 64)
	if err != nil || uploadLength <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_length",
			"message": "Invalid or missing Upload-Length header",
		})
	}

	// Check max size
	if uploadLength > h.cfg.MaxUploadSize {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"error":   "file_too_large",
			"message": fmt.Sprintf("File exceeds maximum size of %d bytes", h.cfg.MaxUploadSize),
		})
	}

	// Parse metadata
	metadata := parseUploadMetadata(c.Get("Upload-Metadata"))
	filename := utils.SanitizeFilename(metadata["filename"])
	if filename == "" {
		filename = "unnamed_file"
	}

	// Generate upload ID
	uploadID, err := utils.GenerateShortID(12)
	if err != nil {
		logging.Error("Failed to generate upload ID", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "id_generation_failed",
			"message": "Failed to generate upload ID",
		})
	}

	// Create upload record
	upload := &database.Upload{
		ID:        uploadID,
		Filename:  sql.NullString{String: filename, Valid: true},
		FileSize:  sql.NullInt64{Int64: uploadLength, Valid: true},
		Offset:    0,
		Metadata:  sql.NullString{String: c.Get("Upload-Metadata"), Valid: true},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.uploadRepo.Create(upload); err != nil {
		logging.Error("Failed to create upload record", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "create_failed",
			"message": "Failed to create upload",
		})
	}

	// Create temp file
	tempPath := h.getTempPath(uploadID)
	file, err := os.Create(tempPath)
	if err != nil {
		logging.Error("Failed to create temp file", zap.Error(err))
		h.uploadRepo.Delete(uploadID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "temp_file_failed",
			"message": "Failed to create temporary file",
		})
	}
	file.Close()

	logging.Info("Upload created",
		zap.String("upload_id", uploadID),
		zap.String("filename", filename),
		zap.Int64("size", uploadLength),
	)

	// Return location
	location := fmt.Sprintf("/api/upload/%s", uploadID)
	c.Set("Location", location)
	c.Set("Tus-Resumable", tusVersion)
	return c.SendStatus(fiber.StatusCreated)
}

// Head handles HEAD requests to get upload offset
func (h *TusHandler) Head(c *fiber.Ctx) error {
	uploadID := c.Params("id")
	if uploadID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "missing_id",
			"message": "Upload ID is required",
		})
	}

	upload, err := h.uploadRepo.GetByID(uploadID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "not_found",
				"message": "Upload not found",
			})
		}
		logging.Error("Failed to get upload", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "get_failed",
			"message": "Failed to get upload",
		})
	}

	c.Set("Tus-Resumable", tusVersion)
	c.Set("Upload-Offset", strconv.FormatInt(upload.Offset, 10))
	if upload.FileSize.Valid {
		c.Set("Upload-Length", strconv.FormatInt(upload.FileSize.Int64, 10))
	}

	return c.SendStatus(fiber.StatusOK)
}

// Patch handles PATCH requests to upload chunks
func (h *TusHandler) Patch(c *fiber.Ctx) error {
	// Verify tus version
	if c.Get("Tus-Resumable") != tusVersion {
		return c.Status(fiber.StatusPreconditionFailed).JSON(fiber.Map{
			"error":   "unsupported_version",
			"message": "Unsupported Tus-Resumable version",
		})
	}

	uploadID := c.Params("id")
	if uploadID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "missing_id",
			"message": "Upload ID is required",
		})
	}

	// Get upload record
	upload, err := h.uploadRepo.GetByID(uploadID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "not_found",
				"message": "Upload not found",
			})
		}
		logging.Error("Failed to get upload", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "get_failed",
			"message": "Failed to get upload",
		})
	}

	// Verify offset
	clientOffset, err := strconv.ParseInt(c.Get("Upload-Offset"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_offset",
			"message": "Invalid Upload-Offset header",
		})
	}

	if clientOffset != upload.Offset {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":          "offset_mismatch",
			"message":        "Upload-Offset does not match current offset",
			"current_offset": upload.Offset,
		})
	}

	// Verify content type
	contentType := c.Get("Content-Type")
	if contentType != "application/offset+octet-stream" {
		return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
			"error":   "invalid_content_type",
			"message": "Content-Type must be application/offset+octet-stream",
		})
	}

	// Open temp file for appending
	tempPath := h.getTempPath(uploadID)
	file, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logging.Error("Failed to open temp file", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "temp_file_failed",
			"message": "Failed to open temporary file",
		})
	}
	defer file.Close()

	// Write chunk to file
	body := c.Body()
	bytesWritten, err := file.Write(body)
	if err != nil {
		logging.Error("Failed to write chunk", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "write_failed",
			"message": "Failed to write chunk",
		})
	}

	// Update offset
	newOffset := upload.Offset + int64(bytesWritten)
	if err := h.uploadRepo.UpdateOffset(uploadID, newOffset); err != nil {
		logging.Error("Failed to update offset", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "update_failed",
			"message": "Failed to update upload offset",
		})
	}

	logging.Debug("Chunk uploaded",
		zap.String("upload_id", uploadID),
		zap.Int("bytes", bytesWritten),
		zap.Int64("new_offset", newOffset),
	)

	// Check if upload is complete
	if upload.FileSize.Valid && newOffset >= upload.FileSize.Int64 {
		// Upload complete - process the file
		go h.processCompletedUpload(uploadID, upload)
	}

	c.Set("Tus-Resumable", tusVersion)
	c.Set("Upload-Offset", strconv.FormatInt(newOffset, 10))
	return c.SendStatus(fiber.StatusNoContent)
}

// Delete handles DELETE requests to cancel an upload
func (h *TusHandler) Delete(c *fiber.Ctx) error {
	uploadID := c.Params("id")
	if uploadID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "missing_id",
			"message": "Upload ID is required",
		})
	}

	// Check if upload exists
	_, err := h.uploadRepo.GetByID(uploadID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "not_found",
				"message": "Upload not found",
			})
		}
		logging.Error("Failed to get upload", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "get_failed",
			"message": "Failed to get upload",
		})
	}

	// Delete temp file
	tempPath := h.getTempPath(uploadID)
	os.Remove(tempPath)

	// Delete upload record
	if err := h.uploadRepo.Delete(uploadID); err != nil {
		logging.Error("Failed to delete upload", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "delete_failed",
			"message": "Failed to delete upload",
		})
	}

	logging.Info("Upload cancelled", zap.String("upload_id", uploadID))

	c.Set("Tus-Resumable", tusVersion)
	return c.SendStatus(fiber.StatusNoContent)
}

// processCompletedUpload handles the completed upload asynchronously
func (h *TusHandler) processCompletedUpload(uploadID string, upload *database.Upload) {
	logging.Info("Processing completed upload", zap.String("upload_id", uploadID))

	tempPath := h.getTempPath(uploadID)
	defer func() {
		// Clean up temp file and upload record
		os.Remove(tempPath)
		h.uploadRepo.Delete(uploadID)
	}()

	// Check WhatsApp connection
	if !h.waClient.IsConnected() {
		logging.Error("WhatsApp not connected, cannot process upload", zap.String("upload_id", uploadID))
		return
	}

	// Read file
	fileData, err := os.ReadFile(tempPath)
	if err != nil {
		logging.Error("Failed to read temp file", zap.Error(err), zap.String("upload_id", uploadID))
		return
	}

	// Calculate hash for tracking purposes
	fileHash := utils.HashFile(fileData)

	// Parse metadata
	metadata := parseUploadMetadata(upload.Metadata.String)
	filename := utils.SanitizeFilename(metadata["filename"])
	if filename == "" {
		filename = "unnamed_file"
	}
	description := metadata["description"]
	password := metadata["password"]

	// Calculate expiry
	expiryDays := h.cfg.DefaultExpiryDays
	if expStr := metadata["expires_in"]; expStr != "" {
		if seconds, err := strconv.ParseInt(expStr, 10, 64); err == nil && seconds > 0 {
			days := int(seconds / 86400)
			if days > 0 && days <= h.cfg.MaxExpiryDays {
				expiryDays = days
			}
		}
	}
	expiresAt := time.Now().Add(time.Duration(expiryDays) * 24 * time.Hour)

	// Parse max downloads
	var maxDownloads sql.NullInt64
	if maxStr := metadata["max_downloads"]; maxStr != "" {
		if val, err := strconv.ParseInt(maxStr, 10, 64); err == nil && val > 0 {
			maxDownloads = sql.NullInt64{Int64: val, Valid: true}
		}
	}

	// Hash password if provided
	var passwordHash sql.NullString
	if password != "" {
		hash, err := utils.HashPassword(password)
		if err != nil {
			logging.Error("Failed to hash password", zap.Error(err))
		} else {
			passwordHash = sql.NullString{String: hash, Valid: true}
		}
	}

	// Detect MIME type
	mimeType := http.DetectContentType(fileData)
	if mimeType == "application/octet-stream" {
		mimeType = "application/octet-stream"
	}

	// Get correct media type for WhatsApp
	mediaType := utils.GetMediaType(mimeType)

	// Upload to WhatsApp
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	uploadResp, err := h.waClient.Upload(ctx, fileData, mediaType)
	if err != nil {
		logging.Error("Failed to upload to WhatsApp", zap.Error(err), zap.String("upload_id", uploadID))
		return
	}

	// Generate file ID
	fileID, err := utils.GenerateShortID(h.cfg.ShortIDLength)
	if err != nil {
		logging.Error("Failed to generate file ID", zap.Error(err))
		return
	}

	// Create file record
	dbFile := &database.File{
		ID:            fileID,
		Filename:      filename,
		MimeType:      mimeType,
		FileSize:      int64(len(fileData)),
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
		logging.Error("Failed to save file record", zap.Error(err), zap.String("upload_id", uploadID))
		return
	}

	logging.Info("Chunked upload completed successfully",
		zap.String("upload_id", uploadID),
		zap.String("file_id", fileID),
		zap.String("filename", filename),
		zap.Int("size", len(fileData)),
	)
}

// getTempPath returns the temp file path for an upload
func (h *TusHandler) getTempPath(uploadID string) string {
	return filepath.Join(h.cfg.TempDir, uploadID+".tmp")
}

// parseUploadMetadata parses the Upload-Metadata header
func parseUploadMetadata(header string) map[string]string {
	metadata := make(map[string]string)
	if header == "" {
		return metadata
	}

	pairs := strings.Split(header, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		parts := strings.SplitN(pair, " ", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value, err := base64.StdEncoding.DecodeString(strings.TrimSpace(parts[1]))
		if err != nil {
			continue
		}

		metadata[key] = string(value)
	}

	return metadata
}
