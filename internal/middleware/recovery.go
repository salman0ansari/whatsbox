package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/salman0ansari/whatsbox/internal/logging"
	"go.uber.org/zap"
)

// Recovery recovers from panics and logs them
func Recovery() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				requestID := GetRequestID(c)
				logging.Error("Panic recovered",
					zap.String("request_id", requestID),
					zap.Any("panic", r),
					zap.String("path", c.Path()),
				)
				_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "internal_server_error",
					"message": "An unexpected error occurred",
				})
			}
		}()
		return c.Next()
	}
}
