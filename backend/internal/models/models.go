package models

import (
	"os"

	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Models struct {
	db *gorm.DB

	Users  *UserModel
	Tags   *TagModel
	Images *ImageModel
	Log    *AuditLog
}

func Migrate(db *gorm.DB) {
}

func MustInitDB() *Models {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to postgres: %w", err))
	}

	Migrate(db)
	err = db.AutoMigrate(&ImageMetadata{}, &Tag{}, &User{}, &AuditEntry{}, &EntryType{})
	if err != nil {
		panic(fmt.Sprintf("failed to auto migrate database: %v", err.Error()))
	}

	tags := &TagModel{db: db}
	users := &UserModel{db: db}
	log := &AuditLog{db: db}
	models := &Models{
		db:     db,
		Users:  users,
		Log:    log,
		Tags:   tags,
		Images: &ImageModel{Tags: tags, Db: db},
	}

	return models
}
