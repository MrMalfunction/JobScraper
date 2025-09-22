package db

import (
	_ "database/sql"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase(dsn string) {
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database!", err)
	}
	DB = database

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get generic database object!", err)
	}
	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)   // max idle connections
	sqlDB.SetMaxOpenConns(1000) // max open connections
}
