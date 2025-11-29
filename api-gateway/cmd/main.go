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
	appgrpc "github.com/xiiisorate/granula_api/api-gateway/internal/grpc"
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

	// ==========================================================================
	// Create Handlers
	// ==========================================================================
	authHandler := handlers.NewAuthHandler(grpcClients.AuthConn, grpcClients.UserConn)
	userHandler := handlers.NewUserHandler(grpcClients.UserConn, grpcClients.AuthConn)
	notificationHandler := handlers.NewNotificationHandler(grpcClients.NotificationConn)
	workspaceHandler := handlers.NewWorkspaceHandler(grpcClients.WorkspaceConn)
	sceneHandler := handlers.NewSceneHandler(grpcClients.SceneConn)
	aiHandler := handlers.NewAIHandler(grpcClients.AIConn)

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
		return c.JSON(fiber.Map{
			"status":   "ready",
			"services": grpcClients.HealthCheck(c.Context()),
		})
	})

	// ==========================================================================
	// Swagger Documentation
	// ==========================================================================
	swaggerHandler := handlers.NewSwaggerHandler("./docs/swagger.yaml")
	app.Get("/swagger", swaggerHandler.ServeUI)
	app.Get("/swagger/spec.yaml", swaggerHandler.ServeSpec)
	app.Get("/swagger/spec.json", swaggerHandler.ServeSpecJSON)
	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger", fiber.StatusMovedPermanently)
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
	// Workspace routes (protected)
	// --------------------------------------------------------------------------
	workspaces := api.Group("/workspaces")
	workspaces.Use(middleware.Auth(cfg))
	{
		workspaces.Get("/", workspaceHandler.ListWorkspaces)
		workspaces.Post("/", workspaceHandler.CreateWorkspace)
		workspaces.Get("/:id", workspaceHandler.GetWorkspace)
		workspaces.Patch("/:id", workspaceHandler.UpdateWorkspace)
		workspaces.Delete("/:id", workspaceHandler.DeleteWorkspace)

		// Workspace members
		workspaces.Get("/:id/members", workspaceHandler.GetMembers)
		workspaces.Post("/:id/members", workspaceHandler.AddMember)
		workspaces.Delete("/:id/members/:memberId", workspaceHandler.RemoveMember)

		// Workspace scenes
		workspaces.Get("/:workspace_id/scenes", sceneHandler.ListScenes)
		workspaces.Post("/:workspace_id/scenes", sceneHandler.CreateScene)
	}

	// --------------------------------------------------------------------------
	// Scene routes (protected)
	// --------------------------------------------------------------------------
	scenes := api.Group("/scenes")
	scenes.Use(middleware.Auth(cfg))
	{
		scenes.Get("/:id", sceneHandler.GetScene)
		scenes.Patch("/:id", sceneHandler.UpdateScene)
		scenes.Delete("/:id", sceneHandler.DeleteScene)
		scenes.Get("/:id/compliance", sceneHandler.CheckCompliance)
	}

	// --------------------------------------------------------------------------
	// AI routes (protected)
	// --------------------------------------------------------------------------
	ai := api.Group("/ai")
	ai.Use(middleware.Auth(cfg))
	{
		// Recognition
		ai.Post("/recognize", aiHandler.RecognizeFloorPlan)
		ai.Get("/recognize/:job_id/status", aiHandler.GetRecognitionStatus)

		// Generation
		ai.Post("/generate", aiHandler.GenerateVariants)
		ai.Get("/generate/:job_id/status", aiHandler.GetGenerationStatus)

		// Chat
		ai.Post("/chat", aiHandler.SendChatMessage)
		ai.Get("/chat/history", aiHandler.GetChatHistory)
		ai.Delete("/chat/history", aiHandler.ClearChatHistory)
	}

	// --------------------------------------------------------------------------
	// Placeholder routes for services not yet fully integrated
	// --------------------------------------------------------------------------
	// FloorPlan routes
	floorplans := api.Group("/floor-plans")
	floorplans.Use(middleware.Auth(cfg))
	{
		floorplans.Get("/", placeholderHandler("list floor plans"))
		floorplans.Post("/", createPlaceholderHandler("upload floor plan"))
		floorplans.Get("/:id", placeholderHandler("get floor plan"))
		floorplans.Patch("/:id", placeholderHandler("update floor plan"))
		floorplans.Delete("/:id", placeholderHandler("delete floor plan"))
		floorplans.Post("/:id/reprocess", placeholderHandler("reprocess floor plan"))
		floorplans.Post("/:id/create-scene", createPlaceholderHandler("create scene from floor plan"))
	}

	// Branch routes
	branches := api.Group("/scenes/:scene_id/branches")
	branches.Use(middleware.Auth(cfg))
	{
		branches.Get("/", placeholderHandler("list branches"))
		branches.Post("/", createPlaceholderHandler("create branch"))
		branches.Get("/:id", placeholderHandler("get branch"))
		branches.Patch("/:id", placeholderHandler("update branch"))
		branches.Delete("/:id", placeholderHandler("delete branch"))
		branches.Post("/:id/activate", placeholderHandler("activate branch"))
		branches.Post("/:id/duplicate", createPlaceholderHandler("duplicate branch"))
		branches.Get("/:id/compare/:target_id", placeholderHandler("compare branches"))
		branches.Post("/:id/merge", placeholderHandler("merge branch"))
	}

	// Compliance routes
	compliance := api.Group("/compliance")
	compliance.Use(middleware.Auth(cfg))
	{
		compliance.Post("/check", placeholderHandler("check compliance"))
		compliance.Post("/check-operation", placeholderHandler("check operation compliance"))
		compliance.Get("/rules", placeholderHandler("list compliance rules"))
		compliance.Get("/rules/:id", placeholderHandler("get compliance rule"))
	}

	// Request routes (expert requests)
	requests := api.Group("/requests")
	requests.Use(middleware.Auth(cfg))
	{
		requests.Get("/", placeholderHandler("list requests"))
		requests.Post("/", createPlaceholderHandler("create request"))
		requests.Get("/:id", placeholderHandler("get request"))
		requests.Patch("/:id", placeholderHandler("update request"))
		requests.Delete("/:id", placeholderHandler("cancel request"))
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

// createPlaceholderHandler creates a placeholder handler that returns 201 for create operations.
func createPlaceholderHandler(action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message":    "Endpoint pending implementation",
			"action":     action,
			"request_id": c.GetRespHeader("X-Request-ID"),
		})
	}
}
