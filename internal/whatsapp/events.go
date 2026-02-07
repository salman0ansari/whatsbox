package whatsapp

import (
	"time"

	"github.com/salman0ansari/whatsbox/internal/logging"
	"go.mau.fi/whatsmeow/types/events"
	"go.uber.org/zap"
)

// eventHandler handles events from the WhatsApp client
func (c *Client) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Connected:
		c.mu.Lock()
		c.connected = true
		c.connectedAt = time.Now()
		c.mu.Unlock()
		logging.Info("WhatsApp connected")

	case *events.Disconnected:
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
		logging.Warn("WhatsApp disconnected")

	case *events.LoggedOut:
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
		logging.Warn("WhatsApp logged out",
			zap.Bool("on_connect", v.OnConnect),
			zap.String("reason", v.Reason.String()),
		)

	case *events.StreamReplaced:
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
		logging.Warn("WhatsApp stream replaced (logged in elsewhere)")

	case *events.TemporaryBan:
		logging.Error("WhatsApp temporary ban",
			zap.String("code", v.Code.String()),
			zap.Duration("duration", v.Expire),
		)

	case *events.ConnectFailure:
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
		logging.Error("WhatsApp connection failure",
			zap.String("reason", v.Reason.String()),
		)

	case *events.ClientOutdated:
		logging.Error("WhatsApp client outdated, update required")

	case *events.StreamError:
		logging.Error("WhatsApp stream error",
			zap.String("code", v.Code),
		)

	case *events.PairSuccess:
		c.mu.Lock()
		c.cachedQR = nil // Clear cached QR on successful login
		c.mu.Unlock()
		logging.Info("WhatsApp pairing successful",
			zap.String("id", v.ID.String()),
		)

	case *events.PairError:
		logging.Error("WhatsApp pairing error",
			zap.Error(v.Error),
		)

	case *events.QR:
		// QR events are handled separately via GetQRChannel
		logging.Debug("QR code event received")

	case *events.HistorySync:
		// We don't need chat history for file storage
		logging.Debug("History sync event received (ignored)")

	case *events.PushName:
		logging.Debug("Push name updated", zap.String("name", v.NewPushName))

	default:
		// Ignore other events (messages, receipts, etc.)
	}
}

// AutoReconnect attempts to reconnect when disconnected
func (c *Client) AutoReconnect() {
	go func() {
		for {
			if !c.IsConnected() && c.IsLoggedIn() {
				logging.Info("Attempting to reconnect to WhatsApp...")

				c.mu.Lock()
				c.reconnectCount++
				c.mu.Unlock()

				err := c.client.Connect()
				if err != nil {
					logging.Error("Reconnection failed", zap.Error(err))
					time.Sleep(30 * time.Second)
					continue
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()
}
