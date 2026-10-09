package tag

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSearchPrefixOrdersByVisibleImageCount(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"CREATE TABLE tags (id integer primary key, name text, user_id integer, deleted_at datetime)",
		"CREATE TABLE image_metadata (id integer primary key, deleted_at datetime)",
		"CREATE TABLE image_tags (tag_id integer, image_metadata_id integer)",
		"INSERT INTO tags (id, name, user_id) VALUES (1, 'a-two', 1), (2, 'b-one', 1), (3, 'c-zero', 1), (4, 'd-zero', 1)",
		"INSERT INTO image_metadata (id) VALUES (10), (11), (12)",
		"INSERT INTO image_metadata (id, deleted_at) VALUES (13, CURRENT_TIMESTAMP)",
		"INSERT INTO image_tags (tag_id, image_metadata_id) VALUES (1, 10), (1, 11), (2, 12), (3, 13)",
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}

	repo := NewTagRepository(db)
	tags, err := repo.SearchPrefix(context.Background(), "", 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 4 || tags[0].ID != 1 || tags[0].ImageCount != 2 || tags[1].ID != 2 || tags[1].ImageCount != 1 || tags[2].ID != 3 || tags[2].ImageCount != 0 || tags[3].ID != 4 || tags[3].ImageCount != 0 {
		t.Fatalf("unexpected tags: %+v", tags)
	}

	page, err := repo.SearchPrefix(context.Background(), "", 2, 2)
	if err != nil || len(page) != 2 || page[0].ID != 3 || page[1].ID != 4 {
		t.Fatalf("unexpected second page: %+v, %v", page, err)
	}

	match, err := repo.SearchPrefix(context.Background(), "B-", 0, 10)
	if err != nil || len(match) != 1 || match[0].ID != 2 {
		t.Fatalf("unexpected prefix results: %+v, %v", match, err)
	}

	tag, err := repo.Get(context.Background(), 1)
	if err != nil || tag.ImageCount != 2 {
		t.Fatalf("unexpected tag detail: %+v, %v", tag, err)
	}
}
