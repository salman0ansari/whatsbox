package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/salman0ansari/whatsbox/internal/logging"
	"github.com/salman0ansari/whatsbox/internal/whatsapp"
	"go.uber.org/zap"
)

// AdminHandler handles admin-related endpoints
type AdminHandler struct {
	waClient *whatsapp.Client
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(waClient *whatsapp.Client) *AdminHandler {
	return &AdminHandler{
		waClient: waClient,
	}
}

// GetQR returns a QR code for WhatsApp login
func (h *AdminHandler) GetQR(c *fiber.Ctx) error {
	if h.waClient.IsLoggedIn() {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":   "already_logged_in",
			"message": "WhatsApp is already logged in. Use /api/admin/logout first.",
		})
	}

	// Keep QR pairing session alive beyond this HTTP request.
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Minute)

	qr, err := h.waClient.GetQR(ctx)
	if err != nil {
		logging.Error("Failed to get QR code", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "qr_generation_failed",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"qr_code": qr.Image,
		"timeout": qr.Timeout,
	})
}

// GetStatus returns the WhatsApp connection status
func (h *AdminHandler) GetStatus(c *fiber.Ctx) error {
	status := h.waClient.GetStatus()
	return c.JSON(status)
}

// Logout logs out from WhatsApp
func (h *AdminHandler) Logout(c *fiber.Ctx) error {
	if !h.waClient.IsLoggedIn() {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":   "not_logged_in",
			"message": "WhatsApp is not logged in",
		})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	if err := h.waClient.Logout(ctx); err != nil {
		logging.Error("Failed to logout", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "logout_failed",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}
