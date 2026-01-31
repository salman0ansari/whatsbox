package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/salman0ansari/whatsbox/internal/logging"
	"go.uber.org/zap"
)

// Logger logs incoming requests with timing information
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get request ID
		requestID := GetRequestID(c)

		// Determine log level based on status
		status := c.Response().StatusCode()

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.Get("User-Agent")),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
		}

		switch {
		case status >= 500:
			logging.Error("Request failed", fields...)
		case status >= 400:
			logging.Warn("Request error", fields...)
		default:
			logging.Info("Request completed", fields...)
		}

		return err
	}
}
