package database

import (
	"fmt"
	"os"

	"server/internal/board"
	"server/internal/image"
	"server/internal/log"
	"server/internal/processing"
	"server/internal/tag"
	"server/internal/user"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitGorm(dialector gorm.Dialector, config *gorm.Config) (*gorm.DB, error) {
	if config == nil {
		config = &gorm.Config{}
	}
	config.TranslateError = true
	db, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.AutoMigrate(
		&image.ImageMetadata{},
		&processing.Job{},
		&tag.Tag{},
		&user.User{},
		&log.AuditEntry{},
		&log.AuditGroup{},
		&board.Board{},
		&board.BoardItem{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate database: %w", err)
	}
	if err := log.BackfillGroups(db); err != nil {
		return nil, fmt.Errorf("failed to group existing audit entries: %w", err)
	}
	return db, nil
}

func MustInitPostgreDB() *gorm.DB {
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
