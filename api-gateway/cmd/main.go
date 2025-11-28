// Package main is the entry point for API Gateway.
package main

import (
	"fmt"

	"github.com/xiiisorate/granula_api/api-gateway/internal/config"
	"github.com/xiiisorate/granula_api/api-gateway/internal/handlers"
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
// @schemes https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите "Bearer" и JWT токен

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Init(logger.Config{
		Level:       cfg.LogLevel,
		ServiceName: "api-gateway",
		Pretty:      cfg.AppEnv != "production",
	})

	logger.Info("Starting API Gateway")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(fiberlogger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-ID",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// API routes
	api := app.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	{
		auth.Post("/register", handlers.Register(cfg))
		auth.Post("/login", handlers.Login(cfg))
		auth.Post("/refresh", handlers.RefreshToken(cfg))
	}

	// Protected auth routes
	authProtected := api.Group("/auth")
	authProtected.Use(middleware.Auth(cfg))
	{
		authProtected.Post("/logout", handlers.Logout(cfg))
		authProtected.Post("/logout-all", handlers.LogoutAll(cfg))
	}

	// User routes (protected)
	users := api.Group("/users")
	users.Use(middleware.Auth(cfg))
	{
		users.Get("/me", handlers.GetProfile(cfg))
		users.Patch("/me", handlers.UpdateProfile(cfg))
		users.Put("/me/password", handlers.ChangePassword(cfg))
		users.Delete("/me", handlers.DeleteAccount(cfg))
	}

	// Notification routes (protected)
	notifications := api.Group("/notifications")
	notifications.Use(middleware.Auth(cfg))
	{
		notifications.Get("/", handlers.GetNotifications(cfg))
		notifications.Get("/count", handlers.GetUnreadCount(cfg))
		notifications.Post("/:id/read", handlers.MarkAsRead(cfg))
		notifications.Post("/read-all", handlers.MarkAllAsRead(cfg))
		notifications.Delete("/:id", handlers.DeleteNotification(cfg))
		notifications.Delete("/", handlers.DeleteAllRead(cfg))
	}

	// Start server
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	logger.Info(fmt.Sprintf("API Gateway listening on %s", address))

	if err := app.Listen(address); err != nil {
		logger.Fatal("Failed to start server", err)
	}
}

