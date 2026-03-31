package database

import (
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Database wraps the sql.DB connection pool
type Database struct {
	DB *gorm.DB
}

// internal/database/database.go
func NewDatabase(dsn string) (*Database, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	log.Println("Connected to database (GORM)")

	// Optional: configure pool (similar to before)
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(10 * time.Minute)

	// Ping to verify
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &Database{DB: db}, nil
}

// Close safely closes the connection pool
func (d *Database) Close() error {
	if d.DB == nil {
		return nil
	}
 sqlDB, _ := d.DB.DB()
   return  sqlDB.Close()

}
