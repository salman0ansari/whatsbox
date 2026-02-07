package whatsapp

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/salman0ansari/whatsbox/internal/logging"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.uber.org/zap"
)

// DownloadRequest contains the parameters needed to download a file
type DownloadRequest struct {
	DirectPath  string
	MediaKey    []byte
	FileEncHash []byte
	FileSHA256  []byte
	FileLength  uint64
	MimeType    string
}

// Download downloads a file from WhatsApp servers using the proper message-based approach
func (c *Client) Download(ctx context.Context, req *DownloadRequest) ([]byte, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to WhatsApp")
	}

	logging.Debug("Downloading file from WhatsApp",
		zap.String("direct_path", req.DirectPath),
		zap.String("mime_type", req.MimeType),
		zap.Uint64("file_length", req.FileLength),
	)

	var data []byte
	var err error

	// Detect media type and download with appropriate message object
	if isImageType(req.MimeType) {
		msg := &waE2E.ImageMessage{
			DirectPath:    &req.DirectPath,
			MediaKey:      req.MediaKey,
			Mimetype:      &req.MimeType,
			FileEncSHA256: req.FileEncHash,
			FileSHA256:    req.FileSHA256,
			FileLength:    &req.FileLength,
		}
		data, err = c.client.Download(ctx, msg)
	} else if isVideoType(req.MimeType) {
		msg := &waE2E.VideoMessage{
			DirectPath:    &req.DirectPath,
			MediaKey:      req.MediaKey,
			Mimetype:      &req.MimeType,
			FileEncSHA256: req.FileEncHash,
			FileSHA256:    req.FileSHA256,
			FileLength:    &req.FileLength,
		}
		data, err = c.client.Download(ctx, msg)
	} else if isAudioType(req.MimeType) {
		msg := &waE2E.AudioMessage{
			DirectPath:    &req.DirectPath,
			MediaKey:      req.MediaKey,
			Mimetype:      &req.MimeType,
			FileEncSHA256: req.FileEncHash,
			FileSHA256:    req.FileSHA256,
			FileLength:    &req.FileLength,
		}
		data, err = c.client.Download(ctx, msg)
	} else {
		// Default to document for all other types
		msg := &waE2E.DocumentMessage{
			DirectPath:    &req.DirectPath,
			MediaKey:      req.MediaKey,
			Mimetype:      &req.MimeType,
			FileEncSHA256: req.FileEncHash,
			FileSHA256:    req.FileSHA256,
			FileLength:    &req.FileLength,
		}
		data, err = c.client.Download(ctx, msg)
	}

	if err != nil {
		logging.Error("Failed to download from WhatsApp", zap.Error(err))
		return nil, fmt.Errorf("download failed: %w", err)
	}

	logging.Info("File downloaded from WhatsApp",
		zap.String("direct_path", req.DirectPath),
		zap.Int("size", len(data)),
	)

	return data, nil
}

// DownloadToWriter downloads a file and writes it to the provided writer
func (c *Client) DownloadToWriter(ctx context.Context, req *DownloadRequest, w io.Writer) error {
	data, err := c.Download(ctx, req)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// Helper functions to detect media types
func isImageType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

func isVideoType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "video/")
}

func isAudioType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "audio/")
}
