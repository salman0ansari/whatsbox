package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/salman0ansari/whatsbox/internal/config"
	"github.com/salman0ansari/whatsbox/internal/database"
	"github.com/salman0ansari/whatsbox/internal/frontend"
	"github.com/salman0ansari/whatsbox/internal/handlers"
	"github.com/salman0ansari/whatsbox/internal/jobs"
	"github.com/salman0ansari/whatsbox/internal/logging"
	"github.com/salman0ansari/whatsbox/internal/middleware"
	"github.com/salman0ansari/whatsbox/internal/stats"
	"github.com/salman0ansari/whatsbox/internal/whatsapp"
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

	// Security check: Ensure admin password is configured
	if cfg.AdminPassword == "" {
		logging.Warn("SECURITY WARNING: ADMIN_PASSWORD is not set! Admin panel will be inaccessible.")
		logging.Warn("Please set ADMIN_PASSWORD environment variable to enable admin access.")
	}

	// Setup database
	if err := database.Setup(cfg); err != nil {
		logging.Fatal("Failed to setup database", zap.Error(err))
	}
	defer database.Close()

	// Setup WhatsApp client
	waClient, err := whatsapp.NewClient(cfg)
	if err != nil {
		logging.Fatal("Failed to create WhatsApp client", zap.Error(err))
	}
	defer waClient.Close()

	// Connect to WhatsApp if already logged in
	if err := waClient.Connect(context.Background()); err != nil {
		logging.Error("Failed to connect to WhatsApp", zap.Error(err))
	}

	// Start auto-reconnect
	waClient.AutoReconnect()

	// Initialize stats collector
	stats.Init()

	// Start background job scheduler
	scheduler := jobs.NewScheduler(cfg)
	scheduler.Start()
	defer scheduler.Stop()

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

	// Health handlers
	healthHandler := handlers.NewHealthHandler(waClient.IsConnected)
	app.Get("/health", healthHandler.Health)
	app.Get("/ready", healthHandler.Ready)

	// API routes
	api := app.Group("/api")

	// Public status endpoint (for frontend header)
	api.Get("/status", healthHandler.Status)

	// Admin routes
	adminHandler := handlers.NewAdminHandler(waClient)
	admin := api.Group("/admin")

	// Auth routes (no auth required)
	admin.Post("/login", middleware.Login(cfg))
	admin.Get("/me", middleware.CheckAuth(cfg))

	// Protected admin routes
	adminProtected := admin.Group("")
	adminProtected.Use(middleware.AdminAuth(cfg))
	adminProtected.Get("/qr", adminHandler.GetQR)
	adminProtected.Get("/status", adminHandler.GetStatus)
	adminProtected.Post("/logout", adminHandler.Logout)
	adminProtected.Post("/logout-session", middleware.LogoutSession())

	// Stats routes (protected)
	statsHandler := handlers.NewStatsHandler()
	adminProtected.Get("/stats", statsHandler.GetStats)
	adminProtected.Get("/stats/hourly", statsHandler.GetHourlyStats)
	adminProtected.Get("/stats/daily", statsHandler.GetDailyStats)

	// File routes
	fileHandler := handlers.NewFileHandler(waClient, cfg)
	files := api.Group("/files")
	files.Post("/", fileHandler.Upload)
	files.Get("/:id", fileHandler.Get)
	files.Get("/:id/download", fileHandler.Download)

	// Protected file routes (admin only)
	filesProtected := files.Group("", middleware.AdminAuth(cfg))
	filesProtected.Get("/", fileHandler.List)
	filesProtected.Delete("/:id", fileHandler.Delete)

	// Tus chunked upload routes
	tusHandler := handlers.NewTusHandler(waClient, cfg)
	upload := api.Group("/upload")
	upload.Options("/", tusHandler.Options)
	upload.Post("/", tusHandler.Create)
	upload.Head("/:id", tusHandler.Head)
	upload.Patch("/:id", tusHandler.Patch)
	upload.Delete("/:id", tusHandler.Delete)

	// Serve embedded frontend (SPA with fallback to index.html)
	app.Use("/", frontend.Handler())

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

	// Wait for active transfers to complete (with timeout)
	collector := stats.Get()
	waitStart := time.Now()
	for {
		active := collector.GetActiveTransfers()
		if active == 0 {
			break
		}
		if time.Since(waitStart) > cfg.ShutdownTimeout {
			logging.Warn("Shutdown timeout reached with active transfers",
				zap.Int64("active_transfers", active))
			break
		}
		logging.Info("Waiting for active transfers to complete",
			zap.Int64("active_transfers", active))
		time.Sleep(1 * time.Second)
	}

	// Disconnect WhatsApp
	waClient.Disconnect()

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
