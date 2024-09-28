package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/ouz/gobackend/api/routes"
	"github.com/ouz/gobackend/config"
	"github.com/ouz/gobackend/database"
	"github.com/ouz/gobackend/middleware"
	"github.com/ouz/gobackend/pkg/auth"
	"github.com/ouz/gobackend/pkg/user"
	"github.com/ouz/gobackend/util"
	"gorm.io/gorm"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed to start: %v", err)
	}
}

func run() error {
	if err := config.LoadConfig(); err != nil {
		return err
	}

	db, err := database.ConnectDB()
	if err != nil {
		return err
	}
	defer database.CloseDatabaseConnection(db)

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, userService)

	app := createFiberApp()
	
	setupHealthChecks(app, db)

	setupMiddlewares(app)

	setupAPIRoutes(app, userService, authService)

	app.Use(notFoundHandler)

	go startServer(app)

	// Graceful shutdown
	return handleGracefulShutdown(app)
}

func createFiberApp() *fiber.App {
	return fiber.New(fiber.Config{
		IdleTimeout:  5 * time.Second,
		ErrorHandler: util.ErrorHandler,
	})
}

func setupMiddlewares(app *fiber.App) {
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(limiter.New(limiter.Config{
		Max: 100,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(&fiber.Map{
				"status":  "fail",
				"message": "You have requested too many in a single time-frame! Please wait another minute!",
			})
		},
	}))
	app.Use(middleware.ClientSecret())
}

func setupHealthChecks(app *fiber.App, db *gorm.DB) {
	app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe: func(c *fiber.Ctx) bool {
			return true
		},
		LivenessEndpoint: "/live",
		ReadinessProbe: func(c *fiber.Ctx) bool {
			return database.IsReady(db)
		},
		ReadinessEndpoint: "/ready",
	}))
}

func setupAPIRoutes(app *fiber.App, userService user.Service, authService auth.Service) {
	api := app.Group("/api/v1")
	routes.SetUpUserRoutes(api, userService, authService)
}

func notFoundHandler(c *fiber.Ctx) error {
	return c.SendStatus(404)
}

func startServer(app *fiber.App) {
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func handleGracefulShutdown(app *fiber.App) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		return err
	}

	log.Info("Server gracefully stopped")
	return nil
}
