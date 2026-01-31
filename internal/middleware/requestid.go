package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestID adds a unique request ID to each request
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()[:8]
		}
		c.Set("X-Request-ID", requestID)
		c.Locals("request_id", requestID)
		return c.Next()
	}
}

// GetRequestID retrieves the request ID from context
func GetRequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals("request_id").(string); ok {
		return id
	}
	return ""
}
