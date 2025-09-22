package db

import (
	_ "database/sql"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase(dsn string) {
	var database *gorm.DB
	var err error

	// Retry connection with exponential backoff
	maxRetries := 30
	initialDelay := 1 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			// Test the connection
			sqlDB, sqlErr := database.DB()
			if sqlErr == nil {
				if pingErr := sqlDB.Ping(); pingErr == nil {
					log.Printf("Successfully connected to database on attempt %d", attempt)
					DB = database

					// Set connection pool settings
					sqlDB.SetMaxIdleConns(10)   // max idle connections
					sqlDB.SetMaxOpenConns(1000) // max open connections
					return
				}
			}
		}

		if attempt == maxRetries {
			log.Fatal("Failed to connect to database after all retries!", err)
		}

		delay := time.Duration(attempt) * initialDelay
		log.Printf("Failed to connect to database (attempt %d/%d): %v. Retrying in %v...", attempt, maxRetries, err, delay)
		time.Sleep(delay)
	}
}
