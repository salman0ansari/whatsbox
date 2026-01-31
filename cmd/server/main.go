package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/salman0ansari/whatsbox/internal/config"
	"github.com/salman0ansari/whatsbox/internal/database"
	"github.com/salman0ansari/whatsbox/internal/handlers"
	"github.com/salman0ansari/whatsbox/internal/logging"
	"github.com/salman0ansari/whatsbox/internal/middleware"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logging
	if err := logging.Setup(cfg); err != nil {
		fmt.Printf("Failed to setup logging: %v\n", err)
		os.Exit(1)
	}
	defer logging.Sync()

	logging.Info("Starting WhatsBox server",
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
	)

	// Setup database
	if err := database.Setup(cfg); err != nil {
		logging.Fatal("Failed to setup database", zap.Error(err))
	}
	defer database.Close()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit:             int(cfg.MaxUploadSize),
		DisableStartupMessage: true,
		ErrorHandler:          errorHandler,
	})

	// Middleware
	app.Use(middleware.RequestID())
	app.Use(middleware.Recovery())
	app.Use(middleware.Logger())
	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowMethods:  "GET,POST,PUT,PATCH,DELETE,OPTIONS,HEAD",
		AllowHeaders:  "Origin,Content-Type,Accept,Authorization,X-Request-ID,X-Password,Upload-Length,Upload-Offset,Tus-Resumable,Upload-Metadata",
		ExposeHeaders: "Upload-Offset,Upload-Length,Tus-Version,Tus-Resumable,Tus-Max-Size,Tus-Extension,Location,X-Request-ID",
	}))

	// Health check placeholder function (will be updated when WhatsApp client is added)
	waConnected := func() bool {
		return false // Will be replaced in Phase 2
	}

	// Health handlers
	healthHandler := handlers.NewHealthHandler(waConnected)
	app.Get("/health", healthHandler.Health)
	app.Get("/ready", healthHandler.Ready)

	// API routes will be added in subsequent phases
	api := app.Group("/api")
	_ = api // Placeholder

	// Start server in goroutine
	go func() {
		addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
		logging.Info("Server listening", zap.String("address", addr))
		if err := app.Listen(addr); err != nil {
			logging.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logging.Info("Shutting down server...")

	// Create shutdown context with timeout
	_, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// Shutdown Fiber
	if err := app.ShutdownWithTimeout(cfg.ShutdownTimeout); err != nil {
		logging.Error("Server forced to shutdown", zap.Error(err))
	}

	logging.Info("Server stopped")
}

// errorHandler handles errors returned by handlers
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	requestID := middleware.GetRequestID(c)
	logging.Error("Request error",
		zap.String("request_id", requestID),
		zap.Int("status", code),
		zap.Error(err),
	)

	return c.Status(code).JSON(fiber.Map{
		"error":      message,
		"request_id": requestID,
	})
}
