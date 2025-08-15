package models

import (
	"server/internal/env"

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

func MustInitStorageSqlite() *Models {
	db, err := gorm.Open(sqlite.Open(env.GetEnvString("DB_PATH")), &gorm.Config{})
	db.AutoMigrate(&Image{}, &Tag{}, &User{}, &AuditEntry{}, &EntryType{})

	if err != nil {
		panic("failed to connect database")
	}

	tags := &TagModel{db: db}
	users := &UserModel{db: db}
	log := &AuditLog{db: db}
	models := &Models{
		db:     db,
		Users:  users,
		Log:    log,
		Images: &ImageModel{Tags: tags, db: db},
	}

	return models
}
