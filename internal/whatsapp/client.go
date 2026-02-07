package whatsapp

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/salman0ansari/whatsbox/internal/config"
	"github.com/salman0ansari/whatsbox/internal/logging"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.uber.org/zap"
)

// Client wraps the whatsmeow client with additional functionality
type Client struct {
	client    *whatsmeow.Client
	container *sqlstore.Container
	cfg       *config.Config

	mu             sync.RWMutex
	connected      bool
	connectedAt    time.Time
	reconnectCount int64

	qrChan   chan string
	qrCancel context.CancelFunc

	// QR code caching to prevent excessive reconnections
	cachedQR     *QRCode
	cachedQRTime time.Time
	qrGenerating bool
}

// Status represents the WhatsApp connection status
type Status struct {
	Connected      bool      `json:"connected"`
	LoggedIn       bool      `json:"logged_in"`
	ConnectedAt    time.Time `json:"connected_at,omitempty"`
	ReconnectCount int64     `json:"reconnect_count"`
	PhoneNumber    string    `json:"phone_number,omitempty"`
	PushName       string    `json:"push_name,omitempty"`
}

// QRCode represents a QR code response
type QRCode struct {
	Code    string `json:"code"`
	Image   string `json:"image"` // Base64 PNG
	Timeout int    `json:"timeout"`
}

// zapLogWrapper wraps zap logger for whatsmeow
type zapLogWrapper struct {
	logger *zap.Logger
}

func (z *zapLogWrapper) Debugf(msg string, args ...interface{}) {
	z.logger.Debug(fmt.Sprintf(msg, args...))
}

func (z *zapLogWrapper) Infof(msg string, args ...interface{}) {
	z.logger.Info(fmt.Sprintf(msg, args...))
}

func (z *zapLogWrapper) Warnf(msg string, args ...interface{}) {
	z.logger.Warn(fmt.Sprintf(msg, args...))
}

func (z *zapLogWrapper) Errorf(msg string, args ...interface{}) {
	z.logger.Error(fmt.Sprintf(msg, args...))
}

func (z *zapLogWrapper) Sub(module string) waLog.Logger {
	return &zapLogWrapper{logger: z.logger.With(zap.String("module", module))}
}

// NewClient creates a new WhatsApp client
func NewClient(cfg *config.Config) (*Client, error) {
	ctx := context.Background()

	// Ensure session directory exists
	sessionDir := filepath.Dir(cfg.WASessionPath)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	// Create logger wrapper
	waLogger := &zapLogWrapper{logger: logging.Logger.With(zap.String("component", "whatsmeow"))}

	// Create database container for session storage
	container, err := sqlstore.New(ctx, "sqlite3", cfg.WASessionPath+"?_journal_mode=WAL&_foreign_keys=on", waLogger.Sub("store"))
	if err != nil {
		return nil, fmt.Errorf("failed to create session store: %w", err)
	}

	// Get or create device store
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get device store: %w", err)
	}

	// Create whatsmeow client
	waClient := whatsmeow.NewClient(deviceStore, waLogger.Sub("client"))

	// Set client properties
	osName := "WhatsBox"
	platformType := waCompanionReg.DeviceProps_CHROME
	requireFullSync := false
	store.DeviceProps.Os = &osName
	store.DeviceProps.PlatformType = &platformType
	store.DeviceProps.RequireFullSync = &requireFullSync

	client := &Client{
		client:    waClient,
		container: container,
		cfg:       cfg,
	}

	// Set up event handler
	waClient.AddEventHandler(client.eventHandler)

	return client, nil
}

// Connect connects to WhatsApp
func (c *Client) Connect(ctx context.Context) error {
	if c.client.Store.ID == nil {
		// Not logged in, need QR code
		logging.Info("WhatsApp not logged in, QR code required")
		return nil
	}

	// Already have session, connect
	err := c.client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	return nil
}

// GetQRChannel returns a channel that will receive QR codes for login
func (c *Client) GetQRChannel(ctx context.Context) (<-chan QRCode, error) {
	if c.client.Store.ID != nil {
		return nil, fmt.Errorf("already logged in")
	}

	// Cancel any existing QR channel
	c.mu.Lock()
	if c.qrCancel != nil {
		c.qrCancel()
	}
	qrCtx, cancel := context.WithCancel(ctx)
	c.qrCancel = cancel
	c.mu.Unlock()

	qrChan, _ := c.client.GetQRChannel(qrCtx)
	resultChan := make(chan QRCode, 1)

	go func() {
		defer close(resultChan)
		for evt := range qrChan {
			if evt.Event == "code" {
				// Generate QR code image
				png, err := qrcode.Encode(evt.Code, qrcode.Medium, 256)
				if err != nil {
					logging.Error("Failed to generate QR code image", zap.Error(err))
					continue
				}

				qr := QRCode{
					Code:    evt.Code,
					Image:   base64.StdEncoding.EncodeToString(png),
					Timeout: int(evt.Timeout.Seconds()),
				}

				select {
				case resultChan <- qr:
				case <-qrCtx.Done():
					return
				}
			} else if evt.Event == "success" {
				logging.Info("QR code login successful")
				return
			}
		}
	}()

	// Start connection to generate QR
	err := c.client.Connect()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect for QR: %w", err)
	}

	return resultChan, nil
}

// GetQR returns a single QR code for login with caching
func (c *Client) GetQR(ctx context.Context) (*QRCode, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we have a cached QR code that's still valid
	if c.cachedQR != nil && time.Since(c.cachedQRTime) < time.Duration(c.cachedQR.Timeout)*time.Second {
		logging.Debug("Returning cached QR code")
		// Update remaining timeout
		qr := *c.cachedQR
		elapsed := time.Since(c.cachedQRTime).Seconds()
		remainingTimeout := c.cachedQR.Timeout - int(elapsed)
		if remainingTimeout < 0 {
			remainingTimeout = 0
		}
		qr.Timeout = remainingTimeout
		return &qr, nil
	}

	// If we're already generating a QR code, wait for it
	if c.qrGenerating {
		c.mu.Unlock()
		// Wait up to 10 seconds for generation to complete
		for i := 0; i < 100; i++ {
			time.Sleep(100 * time.Millisecond)
			c.mu.RLock()
			if !c.qrGenerating && c.cachedQR != nil {
				qr := *c.cachedQR
				c.mu.RUnlock()
				return &qr, nil
			}
			c.mu.RUnlock()
		}
		return nil, fmt.Errorf("timeout waiting for QR code generation")
	}

	// Start generating new QR code
	c.qrGenerating = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.qrGenerating = false
		c.mu.Unlock()
	}()

	qrChan, err := c.GetQRChannel(ctx)
	if err != nil {
		return nil, err
	}

	select {
	case qr, ok := <-qrChan:
		if !ok {
			return nil, fmt.Errorf("QR channel closed")
		}
		// Cache the QR code
		c.mu.Lock()
		c.cachedQR = &qr
		c.cachedQRTime = time.Now()
		c.mu.Unlock()
		return &qr, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for QR code")
	}
}

// Disconnect disconnects from WhatsApp
func (c *Client) Disconnect() {
	c.client.Disconnect()
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()
}

// Logout logs out and clears the session
func (c *Client) Logout(ctx context.Context) error {
	if c.client.Store.ID == nil {
		return fmt.Errorf("not logged in")
	}

	err := c.client.Logout(ctx)
	if err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	c.mu.Lock()
	c.connected = false
	c.cachedQR = nil // Clear cached QR on logout
	c.mu.Unlock()

	logging.Info("Logged out from WhatsApp")
	return nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected && c.client.IsConnected()
}

// IsLoggedIn returns whether there's a stored session
func (c *Client) IsLoggedIn() bool {
	return c.client.Store.ID != nil
}

// GetStatus returns the current connection status
func (c *Client) GetStatus() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := Status{
		Connected:      c.connected && c.client.IsConnected(),
		LoggedIn:       c.client.Store.ID != nil,
		ConnectedAt:    c.connectedAt,
		ReconnectCount: c.reconnectCount,
	}

	if c.client.Store.ID != nil {
		status.PhoneNumber = c.client.Store.ID.User
		if c.client.Store.PushName != "" {
			status.PushName = c.client.Store.PushName
		}
	}

	return status
}

// GetClient returns the underlying whatsmeow client for direct operations
func (c *Client) GetClient() *whatsmeow.Client {
	return c.client
}

// Close closes the client and database connection
func (c *Client) Close() error {
	c.Disconnect()
	return c.container.Close()
}
