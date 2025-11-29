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

	// New handlers for FloorPlan, Branch, Compliance, Request
	floorPlanHandler := handlers.NewFloorPlanHandler(grpcClients.FloorPlanConn)
	branchHandler := handlers.NewBranchHandler(grpcClients.BranchConn)
	complianceHandler := handlers.NewComplianceHandler(grpcClients.ComplianceConn)
	requestHandler := handlers.NewRequestHandler(grpcClients.RequestConn)

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
	// CORS configuration
	// Note: AllowCredentials cannot be true when AllowOrigins is "*"
	// For development, we allow all origins without credentials
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID,X-Requested-With",
		ExposeHeaders:    "Content-Length,Content-Type,X-Request-ID",
		AllowCredentials: false, // Must be false when AllowOrigins is "*"
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
	// FloorPlan routes (protected)
	// --------------------------------------------------------------------------
	floorplans := api.Group("/floor-plans")
	floorplans.Use(middleware.Auth(cfg))
	{
		floorplans.Get("/", floorPlanHandler.List)
		floorplans.Post("/", floorPlanHandler.Upload)
		floorplans.Get("/:id", floorPlanHandler.Get)
		floorplans.Patch("/:id", floorPlanHandler.Update)
		floorplans.Delete("/:id", floorPlanHandler.Delete)
		floorplans.Post("/:id/recognize", floorPlanHandler.StartRecognition)
		floorplans.Get("/:id/recognition-status", floorPlanHandler.GetRecognitionStatus)
		floorplans.Get("/:id/download-url", floorPlanHandler.GetDownloadURL)
	}

	// --------------------------------------------------------------------------
	// Branch routes (protected)
	// --------------------------------------------------------------------------
	branches := api.Group("/scenes/:scene_id/branches")
	branches.Use(middleware.Auth(cfg))
	{
		branches.Get("/", branchHandler.List)
		branches.Post("/", branchHandler.Create)
		branches.Get("/:id", branchHandler.Get)
		branches.Patch("/:id", branchHandler.Update)
		branches.Delete("/:id", branchHandler.Delete)
		branches.Post("/:id/activate", branchHandler.Activate)
		branches.Post("/:id/duplicate", branchHandler.Duplicate)
		branches.Get("/:id/compare/:target_id", branchHandler.Compare)
		branches.Post("/:id/merge", branchHandler.Merge)
	}

	// --------------------------------------------------------------------------
	// Compliance routes (protected)
	// --------------------------------------------------------------------------
	compliance := api.Group("/compliance")
	compliance.Use(middleware.Auth(cfg))
	{
		compliance.Post("/check", complianceHandler.Check)
		compliance.Post("/check-operation", complianceHandler.CheckOperation)
		compliance.Get("/rules", complianceHandler.GetRules)
		compliance.Get("/rules/:id", complianceHandler.GetRule)
		compliance.Post("/report", complianceHandler.GenerateReport)
	}

	// --------------------------------------------------------------------------
	// Request routes (expert requests, protected)
	// --------------------------------------------------------------------------
	requests := api.Group("/requests")
	requests.Use(middleware.Auth(cfg))
	{
		requests.Get("/", requestHandler.List)
		requests.Post("/", requestHandler.Create)
		requests.Get("/:id", requestHandler.Get)
		requests.Patch("/:id", requestHandler.Update)
		requests.Delete("/:id", requestHandler.Delete)
		requests.Post("/:id/cancel", requestHandler.Cancel)
		requests.Post("/:id/submit", requestHandler.Submit)
		requests.Post("/:id/documents", requestHandler.AddDocument)
		requests.Get("/:id/documents", requestHandler.GetDocuments)
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
