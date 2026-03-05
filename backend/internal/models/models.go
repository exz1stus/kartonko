package models

import (
	"server/internal/env"

	"fmt"

	"gorm.io/driver/sqlite"
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

func MustInitStorageSqlite() *Models {
	dbPath := env.GetEnvString("DB_PATH")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err.Error()))
	}
	Migrate(db)
	err = db.AutoMigrate(&Image{}, &Tag{}, &User{}, &AuditEntry{}, &EntryType{})
	if err != nil {
		panic(fmt.Sprintf("failed to auto migrate database: %v", err.Error()))
	}

	tags := &TagModel{db: db}
	err = tags.RegisterAutoTags()
	if err != nil {
		panic(fmt.Sprintf("failed to register auto tags: %v", err.Error()))
	}

	users := &UserModel{db: db}
	log := &AuditLog{db: db}
	models := &Models{
		db:     db,
		Users:  users,
		Log:    log,
		Tags:   tags,
		Images: &ImageModel{Tags: tags, db: db},
	}

	return models
}
