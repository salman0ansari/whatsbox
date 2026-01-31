package whatsapp

import (
	"context"
	"fmt"
	"io"

	"github.com/salman0ansari/whatsbox/internal/logging"
	"go.mau.fi/whatsmeow"
	"go.uber.org/zap"
)

// UploadResponse contains the result of uploading a file to WhatsApp
type UploadResponse struct {
	DirectPath  string
	MediaKey    []byte
	FileEncHash []byte
	FileSHA256  []byte
	FileLength  uint64
}

// Upload uploads a file to WhatsApp servers
func (c *Client) Upload(ctx context.Context, data []byte, mediaType whatsmeow.MediaType) (*UploadResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to WhatsApp")
	}

	logging.Debug("Uploading file to WhatsApp",
		zap.Int("size", len(data)),
		zap.String("media_type", string(mediaType)),
	)

	resp, err := c.client.Upload(ctx, data, mediaType)
	if err != nil {
		logging.Error("Failed to upload to WhatsApp", zap.Error(err))
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	logging.Info("File uploaded to WhatsApp",
		zap.String("direct_path", resp.DirectPath),
		zap.Uint64("file_length", resp.FileLength),
	)

	return &UploadResponse{
		DirectPath:  resp.DirectPath,
		MediaKey:    resp.MediaKey,
		FileEncHash: resp.FileEncSHA256,
		FileSHA256:  resp.FileSHA256,
		FileLength:  resp.FileLength,
	}, nil
}

// UploadFromReader uploads a file from a reader to WhatsApp servers
func (c *Client) UploadFromReader(ctx context.Context, reader io.Reader, mediaType whatsmeow.MediaType) (*UploadResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to WhatsApp")
	}

	// Read all data from reader
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	return c.Upload(ctx, data, mediaType)
}
