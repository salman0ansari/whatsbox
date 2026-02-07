package middleware

import (
	"crypto/subtle"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/salman0ansari/whatsbox/internal/config"
)

const (
	authCookieName = "whatsbox_admin_session"
)

// AdminAuth creates an admin authentication middleware
func AdminAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If no admin password is set, deny all access
		if cfg.AdminPassword == "" {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":   "auth_not_configured",
				"message": "Admin authentication is not configured. Set ADMIN_PASSWORD environment variable.",
			})
		}

		// Get session token from cookie
		token := c.Cookies(authCookieName)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
		}

		// Validate JWT token
		claims := &jwt.RegisteredClaims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(cfg.AdminSessionSecret), nil
		})

		if err != nil || !parsedToken.Valid {
			// Clear invalid cookie
			c.Cookie(&fiber.Cookie{
				Name:     authCookieName,
				Value:    "",
				Expires:  time.Now().Add(-1 * time.Hour),
				HTTPOnly: true,
				Secure:   c.Protocol() == "https",
				SameSite: "Lax",
				Path:     "/",
			})
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Invalid or expired session",
			})
		}

		return c.Next()
	}
}

// Login handles admin login
func Login(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If no admin password is set, return error
		if cfg.AdminPassword == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "auth_disabled",
				"message": "Admin authentication is not configured. Set ADMIN_PASSWORD environment variable.",
			})
		}

		// Parse request body
		var req struct {
			Password string `json:"password"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "invalid_request",
				"message": "Invalid request body",
			})
		}

		// Validate password using constant-time comparison
		if subtle.ConstantTimeCompare([]byte(req.Password), []byte(cfg.AdminPassword)) != 1 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "invalid_credentials",
				"message": "Invalid password",
			})
		}

		// Generate JWT token
		claims := &jwt.RegisteredClaims{
			Subject:   "admin",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.AdminSessionMaxAge) * time.Second)),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(cfg.AdminSessionSecret))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "token_generation_failed",
				"message": "Failed to generate session token",
			})
		}

		// Set HTTP-only cookie
		c.Cookie(&fiber.Cookie{
			Name:     authCookieName,
			Value:    tokenString,
			Expires:  time.Now().Add(time.Duration(cfg.AdminSessionMaxAge) * time.Second),
			HTTPOnly: true,
			Secure:   c.Protocol() == "https",
			SameSite: "Lax",
			Path:     "/",
		})

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Logged in successfully",
		})
	}
}

// LogoutSession handles admin session logout
func LogoutSession() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Clear the session cookie
		c.Cookie(&fiber.Cookie{
			Name:     authCookieName,
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour),
			HTTPOnly: true,
			Secure:   c.Protocol() == "https",
			SameSite: "Lax",
			Path:     "/",
		})

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Logged out successfully",
		})
	}
}

// CheckAuth returns the current authentication status
func CheckAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If no admin password is set, authentication is required but not possible
		if cfg.AdminPassword == "" {
			return c.JSON(fiber.Map{
				"authenticated": false,
				"auth_required": true,
				"message":       "Admin authentication is not configured. Set ADMIN_PASSWORD environment variable.",
			})
		}

		// Get session token from cookie
		token := c.Cookies(authCookieName)
		if token == "" {
			return c.JSON(fiber.Map{
				"authenticated": false,
				"auth_required": true,
			})
		}

		// Validate JWT token
		claims := &jwt.RegisteredClaims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(cfg.AdminSessionSecret), nil
		})

		if err != nil || !parsedToken.Valid {
			return c.JSON(fiber.Map{
				"authenticated": false,
				"auth_required": true,
			})
		}

		return c.JSON(fiber.Map{
			"authenticated": true,
			"auth_required": true,
		})
	}
}
