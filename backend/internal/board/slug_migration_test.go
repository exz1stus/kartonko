package board

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestEnsureSlugsKeepsDuplicateNamesDistinct(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"CREATE TABLE boards (id integer primary key, created_at datetime, updated_at datetime, deleted_at datetime, name text, slug text UNIQUE, description text, user_id integer)",
		"CREATE TABLE board_items (id integer primary key, board_id integer, image_id integer)",
		"CREATE TABLE image_metadata (id integer primary key, deleted_at datetime)",
		"INSERT INTO boards (id, name) VALUES (1, 'My Photos'), (2, 'My Photos'), (3, 'Żółw & Kot')",
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := EnsureSlugs(db); err != nil {
		t.Fatal(err)
	}
	repo := NewBoardRepository(db)
	for slug, id := range map[string]uint{"my-photos": 1, "my-photos-2": 2, "żółw-kot": 3} {
		board, err := repo.GetBySlug(context.Background(), slug)
		if err != nil || board.ID != id {
			t.Fatalf("slug %q resolved to %+v: %v", slug, board, err)
		}
	}
	if err := EnsureSlugs(db); err != nil {
		t.Fatalf("backfill should be repeatable: %v", err)
	}
	created, err := NewBoardService(repo, nil).Create(context.Background(), 5, &BoardCreateRequest{Name: "My Photos"})
	if err != nil || created.Slug != "my-photos-3" {
		t.Fatalf("new boards should get a free name-based slug: %+v, %v", created, err)
	}
}
