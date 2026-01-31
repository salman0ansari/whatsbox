package whatsapp

import (
	"context"
	"fmt"
	"io"

	"github.com/salman0ansari/whatsbox/internal/logging"
	"go.mau.fi/whatsmeow"
	"go.uber.org/zap"
)

// DownloadRequest contains the parameters needed to download a file
type DownloadRequest struct {
	DirectPath  string
	MediaKey    []byte
	FileEncHash []byte
	FileSHA256  []byte
	FileLength  uint64
	MediaType   whatsmeow.MediaType
}

// Download downloads a file from WhatsApp servers
func (c *Client) Download(ctx context.Context, req *DownloadRequest) ([]byte, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to WhatsApp")
	}

	logging.Debug("Downloading file from WhatsApp",
		zap.String("direct_path", req.DirectPath),
		zap.Uint64("file_length", req.FileLength),
	)

	// Get media type string for the download
	mediaTypeStr := string(req.MediaType)

	data, err := c.client.DownloadMediaWithPath(
		ctx,
		req.DirectPath,
		req.FileEncHash,
		req.FileSHA256,
		req.MediaKey,
		int(req.FileLength),
		req.MediaType,
		mediaTypeStr,
	)
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
