package image

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestImageQueriesLoadTagsWithoutCountColumn(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"CREATE TABLE image_metadata (id integer primary key, created_at datetime, updated_at datetime, deleted_at datetime, filename text, hash text, format text, width integer, height integer, user_id integer)",
		"CREATE TABLE tags (id integer primary key, created_at datetime, updated_at datetime, deleted_at datetime, name text, user_id integer)",
		"CREATE TABLE image_tags (image_metadata_id integer, tag_id integer)",
		"INSERT INTO image_metadata (id, filename, hash) VALUES (1, 'first', 'hash')",
		"INSERT INTO tags (id, name) VALUES (2, 'example')",
		"INSERT INTO image_tags (image_metadata_id, tag_id) VALUES (1, 2)",
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}

	repo := NewImageRepository(db)
	image, err := repo.Get(context.Background(), 1)
	if err != nil || len(image.Tags) != 1 || image.Tags[0].Name != "example" {
		t.Fatalf("get with tags: %+v, %v", image, err)
	}

	images, err := repo.Search(context.Background(), &Query{Tags: []string{"example"}})
	if err != nil || len(images) != 1 || len(images[0].Tags) != 1 || images[0].Tags[0].Name != "example" {
		t.Fatalf("search with tags: %+v, %v", images, err)
	}
}
