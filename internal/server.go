package internal

import (
	"job-scraper/internal/config"
	"job-scraper/internal/db"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// echo-swagger middleware

// CustomValidator implements echo.Validator interface
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates structs using validator package
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func StartServer() {
	// Configure log level based on environment variable
	logLevel := slog.LevelInfo // default to Info
	if os.Getenv("LOG_LEVEL") == "debug" || os.Getenv("LOG_LEVEL") == "DEBUG" {
		logLevel = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	logger.Debug("Logger initialized", "level", logLevel.String())

	// Load config
	config.LoadSecrets()

	// Connect to PostgreSQL
	dsn := config.GetSecrets().DatabaseDSN
	if dsn == "" {
		logger.Error("Database DSN not set in config.yaml")
		os.Exit(1)
	}
	db.ConnectDatabase(dsn)

	// Install pg_trgm extension for trigram operations
	if err := db.DB.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		logger.Error("Failed to install pg_trgm extension", "error", err)
		panic("Extension installation failed")
	}
	logger.Info("pg_trgm extension installed")

	// Auto-migrate models (add all models here as your app grows)
	if err := db.DB.AutoMigrate(&db.Companies{}, &db.Jobs{}); err != nil {
		logger.Error("AutoMigrate failed", "error", err)
		panic("Automigration Failed")
	}
	logger.Info("Auto Migration Successful")

	// Create GIN indexes manually for trigram operations
	if err := db.DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_role_trgm ON jobs USING gin (job_role gin_trgm_ops)`).Error; err != nil {
		logger.Error("Failed to create GIN index on job_role", "error", err)
		panic("Index creation failed")
	}
	logger.Info("GIN index on job_role created")

	if err := db.DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_job_details_trgm ON jobs USING gin (job_details gin_trgm_ops)`).Error; err != nil {
		logger.Error("Failed to create GIN index on job_details", "error", err)
		panic("Index creation failed")
	}
	logger.Info("GIN index on job_details created")

	e := echo.New()
	// Middleware
	// e.Use(echomiddleware.Logger())
	e.Use(middleware.CORS())
	e.Use(echomiddleware.Recover())
	e.Validator = &CustomValidator{validator: validator.New()}
	attachPaths(e)

	e.GET("/health", func(c echo.Context) error {
		slog.Info("Health check endpoint hit")
		return c.JSON(http.StatusOK, echo.Map{
			"status": "Healthy",
		})
	})

	// Start server
	port := os.Getenv("PORT")
	logger.Info("API Port", "port", port)
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}

func attachPaths(e *echo.Echo) {
	// Public routes
	e.POST("/add_scrape_company/workday", SubmitWorkdayCompanyToScrape)
	e.POST("/add_scrape_company/greenhouse", SubmitGreenhouseCompanyToScrape)
	e.POST("/add_scrape_company/oraclecloud", SubmitOracleCloudCompanyToScrape)
	e.GET("/start_scrape", SubmitScrapeRequest)

	// API routes for frontend
	api := e.Group("/api")
	api.GET("/jobs/search", SearchJobs)
	api.GET("/jobs/latest", GetLatestJobs)
	api.GET("/jobs/today", GetTodaysJobs)
	api.GET("/jobs/all", GetAllJobs)
	api.GET("/companies", GetCompanies)
	api.PUT("/companies/:name", UpdateCompany)
	api.DELETE("/companies/:name", DeleteCompany)
	api.DELETE("/jobs/cleanup", DeleteOldJobs)

	// Redirect root to /ui
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/ui")
	})

	// Custom handler for /ui routes that serves static files or index.html for SPA routing
	e.GET("/ui*", func(c echo.Context) error {
		path := c.Request().URL.Path

		// Remove /ui prefix to get the file path
		filePath := strings.TrimPrefix(path, "/ui")
		if filePath == "" || filePath == "/" {
			filePath = "/index.html"
		}

		// Build full path to file
		fullPath := filepath.Join("frontend/dist", filePath)

		// Check if file exists
		if _, err := os.Stat(fullPath); err == nil {
			// File exists, serve it
			return c.File(fullPath)
		}

		// File doesn't exist, serve index.html for SPA routing
		return c.File("frontend/dist/index.html")
	})
}
