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
	Log    *LogModel
}

func Migrate(db *gorm.DB) {
}

func InitGorm(dialector gorm.Dialector, config *gorm.Config) (*Models, error) {
	db, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	Migrate(db)
	err = db.AutoMigrate(&ImageMetadata{}, &Tag{}, &User{}, &AuditEntry{}, &EntryType{})
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate database: %w", err)
	}

	tags := &TagModel{db: db}
	users := &UserModel{db: db}
	log := &LogModel{db: db}
	models := &Models{
		db:     db,
		Users:  users,
		Log:    log,
		Tags:   tags,
		Images: &ImageModel{Tags: tags, Db: db},
	}

	return models, nil
}

func MustInitDB() *Models {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	models, err := InitGorm(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(fmt.Sprint("Failed to initialize database: %w", err))
	}

	return models
}
