package database

import (
	"fmt"
	"os"
	"server/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
}

func InitGorm(dialector gorm.Dialector, config *gorm.Config) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	Migrate(db)
	err = db.AutoMigrate(
		&models.ImageMetadata{},
		&models.Tag{},
		&models.User{},
		&models.AuditEntry{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate database: %w", err)
	}

	return db, nil
}

func MustInitDB() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	db, err := InitGorm(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("Failed to initialize database: %w", err))
	}
	return db
}
