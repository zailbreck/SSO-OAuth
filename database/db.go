package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres" // Driver GORM untuk PostgreSQL
	"gorm.io/gorm"
)

// InitDB initializes and returns a GORM database connection pool
func InitDB() (*gorm.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error opening database with gorm: %w", err)
	}

	// Set connection pool settings (optional but recommended)
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Ping aGORM tidak memiliki Ping() langsung, tapi operasi pertama akan memvalidasi koneksi.
	// Jika Anda ingin memastikan koneksi, Anda bisa menggunakan sqlDB.Ping()
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database using GORM!")
	return db, nil
}

// CloseDB sekarang menerima *gorm.DB
func CloseDB(db *gorm.DB) {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("Error getting underlying sql.DB for closing: %v\n", err)
			return
		}
		err = sqlDB.Close()
		if err != nil {
			log.Printf("Error closing database connection: %v\n", err)
		} else {
			log.Println("Database connection closed.")
		}
	}
}
