package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	waConnected func() bool
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(waConnectedFunc func() bool) *HealthHandler {
	return &HealthHandler{
		waConnected: waConnectedFunc,
	}
}

// Health returns basic liveness check
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

// Ready returns readiness check (including WhatsApp connection status)
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	waConnected := false
	if h.waConnected != nil {
		waConnected = h.waConnected()
	}

	if !waConnected {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":   "not_ready",
			"whatsapp": "disconnected",
		})
	}

	return c.JSON(fiber.Map{
		"status":   "ready",
		"whatsapp": "connected",
	})
}

// Status returns public connection status (for frontend)
func (h *HealthHandler) Status(c *fiber.Ctx) error {
	waConnected := false
	if h.waConnected != nil {
		waConnected = h.waConnected()
	}

	return c.JSON(fiber.Map{
		"connected": waConnected,
	})
}
