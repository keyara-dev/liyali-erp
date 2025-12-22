package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"github.com/liyali/liyali-gateway/config"
	"github.com/liyali/liyali-gateway/middleware"
	"github.com/liyali/liyali-gateway/routes"
)

func init() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil && os.Getenv("APP_ENV") == "" {
		log.Println("Note: .env file not found, using environment variables")
	}

	// Set default values if not provided
	if os.Getenv("DB_PORT") == "" {
		os.Setenv("DB_PORT", "5432")
	}
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "localhost")
	}
	if os.Getenv("DB_USER") == "" {
		os.Setenv("DB_USER", "postgres")
	}
	if os.Getenv("DB_SSL_MODE") == "" {
		os.Setenv("DB_SSL_MODE", "disable")
	}
	if os.Getenv("APP_PORT") == "" {
		os.Setenv("APP_PORT", "8080")
	}
	if os.Getenv("FRONTEND_URL") == "" {
		os.Setenv("FRONTEND_URL", "http://localhost:3000")
	}
}

func main() {
	// Initialize database
	config.InitDatabase()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Liyali Gateway Backend",
	})

	// Middleware
	app.Use(middleware.ErrorHandlingMiddleware())
	app.Use(middleware.LoggerMiddleware())
	app.Use(middleware.CORSMiddleware())

	// Setup routes
	routes.SetupRoutes(app)

	// Start server
	port := os.Getenv("APP_PORT")
	log.Printf("🚀 Starting Liyali Gateway Backend on port %s", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
