// =============================================================================
// Package main is the entry point for API Gateway.
// =============================================================================
// API Gateway serves as the single entry point for all client requests,
// routing them to appropriate microservices via gRPC.
//
// Features:
// - HTTP REST API endpoint
// - JWT authentication
// - Rate limiting
// - Request validation
// - Error handling
// - Swagger documentation
//
// =============================================================================
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/xiiisorate/granula_api/api-gateway/internal/config"
	"github.com/xiiisorate/granula_api/api-gateway/internal/handlers"
	appgrpc "github.com/xiiisorate/granula_api/api-gateway/internal/grpc"
	"github.com/xiiisorate/granula_api/api-gateway/internal/middleware"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// @title Granula API
// @version 1.0
// @description REST API для планирования ремонта и перепланировки квартир
// @host api.granula.raitokyokai.tech
// @BasePath /api/v1
// @schemes https http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите "Bearer" и JWT токен

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.MustNew(logger.Config{
		Level:       cfg.LogLevel,
		ServiceName: "api-gateway",
		Format:      "json",
		Development: cfg.AppEnv != "production",
	})
	logger.SetGlobal(log)

	log.Info("Starting API Gateway",
		logger.String("env", cfg.AppEnv),
		logger.String("version", "1.0.0"),
	)

	// Create gRPC clients
	grpcClients, err := appgrpc.NewClients(cfg)
	if err != nil {
		log.Fatal("Failed to create gRPC clients", logger.Err(err))
	}
	defer grpcClients.Close()

	log.Info("Connected to backend services")

	// Create handlers
	authHandler := handlers.NewAuthHandler(grpcClients.AuthConn)
	userHandler := handlers.NewUserHandler(grpcClients.UserConn)
	notificationHandler := handlers.NewNotificationHandler(grpcClients.NotificationConn)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler:          middleware.ErrorHandler,
		DisableStartupMessage: true,
		ReadTimeout:           cfg.ReadTimeout,
		WriteTimeout:          cfg.WriteTimeout,
		IdleTimeout:           cfg.IdleTimeout,
	})

	// Global middleware
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.IsDevelopment(),
	}))
	app.Use(requestid.New())
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: cfg.CORSAllowCredentials,
	}))

	// ==========================================================================
	// Health & Service Status
	// ==========================================================================
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "api-gateway",
			"version": "1.0.0",
		})
	})

	app.Get("/ready", func(c *fiber.Ctx) error {
		// Check all backend services
		return c.JSON(fiber.Map{
			"status":   "ready",
			"services": grpcClients.HealthCheck(c.Context()),
		})
	})

	// ==========================================================================
	// API Routes
	// ==========================================================================
	api := app.Group("/api/v1")

	// --------------------------------------------------------------------------
	// Auth routes (public)
	// --------------------------------------------------------------------------
	auth := api.Group("/auth")
	{
		auth.Post("/register", authHandler.Register)
		auth.Post("/login", authHandler.Login)
		auth.Post("/refresh", authHandler.RefreshToken)
	}

	// Protected auth routes
	authProtected := api.Group("/auth")
	authProtected.Use(middleware.Auth(cfg))
	{
		authProtected.Post("/logout", authHandler.Logout)
		authProtected.Post("/logout-all", authHandler.LogoutAll)
	}

	// --------------------------------------------------------------------------
	// User routes (protected)
	// --------------------------------------------------------------------------
	users := api.Group("/users")
	users.Use(middleware.Auth(cfg))
	{
		users.Get("/me", userHandler.GetProfile)
		users.Patch("/me", userHandler.UpdateProfile)
		users.Put("/me/password", userHandler.ChangePassword)
		users.Delete("/me", userHandler.DeleteAccount)
	}

	// --------------------------------------------------------------------------
	// Notification routes (protected)
	// --------------------------------------------------------------------------
	notifications := api.Group("/notifications")
	notifications.Use(middleware.Auth(cfg))
	{
		notifications.Get("/", notificationHandler.GetNotifications)
		notifications.Get("/count", notificationHandler.GetUnreadCount)
		notifications.Post("/:id/read", notificationHandler.MarkAsRead)
		notifications.Post("/read-all", notificationHandler.MarkAllAsRead)
		notifications.Delete("/:id", notificationHandler.DeleteNotification)
		notifications.Delete("/", notificationHandler.DeleteAllRead)
	}

	// --------------------------------------------------------------------------
	// Workspace routes (protected) - placeholder
	// --------------------------------------------------------------------------
	workspaces := api.Group("/workspaces")
	workspaces.Use(middleware.Auth(cfg))
	{
		workspaces.Get("/", placeholderHandler("list workspaces"))
		workspaces.Post("/", placeholderHandler("create workspace"))
		workspaces.Get("/:id", placeholderHandler("get workspace"))
		workspaces.Patch("/:id", placeholderHandler("update workspace"))
		workspaces.Delete("/:id", placeholderHandler("delete workspace"))
	}

	// --------------------------------------------------------------------------
	// Scene routes (protected) - placeholder
	// --------------------------------------------------------------------------
	scenes := api.Group("/scenes")
	scenes.Use(middleware.Auth(cfg))
	{
		scenes.Get("/", placeholderHandler("list scenes"))
		scenes.Post("/", placeholderHandler("create scene"))
		scenes.Get("/:id", placeholderHandler("get scene"))
		scenes.Patch("/:id", placeholderHandler("update scene"))
		scenes.Delete("/:id", placeholderHandler("delete scene"))
	}

	// --------------------------------------------------------------------------
	// AI routes (protected) - placeholder
	// --------------------------------------------------------------------------
	ai := api.Group("/ai")
	ai.Use(middleware.Auth(cfg))
	{
		ai.Post("/recognize", placeholderHandler("recognize floor plan"))
		ai.Post("/generate", placeholderHandler("generate design"))
		ai.Post("/chat", placeholderHandler("AI chat"))
	}

	// ==========================================================================
	// Start Server
	// ==========================================================================

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		address := cfg.HTTPAddress()
		log.Info("API Gateway listening",
			logger.String("address", address),
			logger.Bool("swagger", cfg.SwaggerEnabled),
		)

		if err := app.Listen(address); err != nil {
			log.Fatal("Failed to start server", logger.Err(err))
		}
	}()

	<-quit
	log.Info("Shutting down API Gateway...")

	if err := app.Shutdown(); err != nil {
		log.Error("Error during shutdown", logger.Err(err))
	}

	log.Info("API Gateway stopped")
}

// placeholderHandler creates a placeholder handler for routes not yet implemented.
func placeholderHandler(action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":    "Endpoint pending implementation",
			"action":     action,
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}
