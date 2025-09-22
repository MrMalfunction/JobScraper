package internal

import (
	"job-scraper/internal/config"
	"job-scraper/internal/db"
	"net/http"
	"os"

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
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load config
	config.LoadSecrets()

	// Connect to PostgreSQL
	dsn := config.GetSecrets().DatabaseDSN
	if dsn == "" {
		logger.Error("Database DSN not set in config.yaml")
		os.Exit(1)
	}
	db.ConnectDatabase(dsn)

	// Auto-migrate models (add all models here as your app grows)
	err := db.DB.AutoMigrate(&db.Companies{}, &db.Jobs{})
	if err != nil {
		logger.Error("AutoMigrate failed", "error", err)
		panic("Automigration Failed")
	}
	logger.Info("Auto Migration Successful")

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
	e.Logger.Fatal(e.Start(":8080"))
}

func attachPaths(e *echo.Echo) {
	// Public routes

	e.POST("/add_scrape_company/workday", SubmitWorkdayCompanyToScrape)
	e.GET("/start_scrape", SubmitScrapeRequest)
}
