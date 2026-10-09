package board

import (
	"context"
	"errors"
	apperrors "server/internal/errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestListFiltersAndCountsImages(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"CREATE TABLE boards (id integer primary key, name text, description text, user_id integer, deleted_at datetime)",
		"CREATE TABLE board_items (id integer primary key, board_id integer, image_id integer, created_at datetime)",
		"CREATE TABLE image_metadata (id integer primary key, filename text, deleted_at datetime)",
		"CREATE TABLE tags (id integer primary key, name text, deleted_at datetime)",
		"CREATE TABLE image_tags (image_metadata_id integer, tag_id integer)",
		"INSERT INTO boards (id, name, user_id) VALUES (1, 'Art', 5), (2, 'Articles', 6), (3, 'Nature', 5), (4, 'Zebra', 6)",
		"INSERT INTO board_items (id, board_id, image_id) VALUES (1, 1, 10), (2, 1, 11), (3, 2, 12), (4, 1, 13)",
		"INSERT INTO image_metadata (id, filename) VALUES (10, 'a.png'), (11, 'b.png'), (12, 'c.png')",
		"INSERT INTO image_metadata (id, filename, deleted_at) VALUES (13, 'deleted.png', CURRENT_TIMESTAMP)",
		"INSERT INTO tags (id, name) VALUES (1, 'red'), (2, 'art')",
		"INSERT INTO image_tags (image_metadata_id, tag_id) VALUES (10, 1), (10, 2), (11, 2)",
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}

	repo := NewBoardRepository(db)
	boards, err := repo.List(context.Background(), "art", 0, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(boards) != 2 || boards[0].ID != 1 || boards[0].ImageCount != 2 || boards[1].ID != 2 || boards[1].ImageCount != 1 {
		t.Fatalf("unexpected search results: %+v", boards)
	}

	owned, err := repo.List(context.Background(), "", 5, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(owned) != 2 || owned[0].ID != 1 || owned[0].ImageCount != 2 || owned[1].ID != 3 || owned[1].ImageCount != 0 {
		t.Fatalf("unexpected owner results: %+v", owned)
	}

	page, err := repo.List(context.Background(), "", 0, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(page) != 1 || page[0].ID != 2 || page[0].ImageCount != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}
	emptyPage, err := repo.List(context.Background(), "", 0, 2, 2)
	if err != nil || len(emptyPage) != 2 || emptyPage[0].ID != 3 || emptyPage[0].ImageCount != 0 || emptyPage[1].ID != 4 || emptyPage[1].ImageCount != 0 {
		t.Fatalf("unexpected empty-board page: %+v, %v", emptyPage, err)
	}

	board, err := repo.Get(context.Background(), 1)
	if err != nil || board.ImageCount != 2 {
		t.Fatalf("unexpected board detail: %+v, %v", board, err)
	}

	items, err := repo.ListItems(context.Background(), 1, "", nil, 0, 10)
	if err != nil || len(items) != 2 || items[0].Image.ID != 11 || items[1].Image.ID != 10 {
		t.Fatalf("unexpected board images: %+v, %v", items, err)
	}
	filtered, err := repo.ListItems(context.Background(), 1, "a", []string{"red", "art"}, 0, 10)
	if err != nil || len(filtered) != 1 || filtered[0].Image.ID != 10 {
		t.Fatalf("unexpected filtered board images: %+v, %v", filtered, err)
	}
	filtered, err = repo.ListItems(context.Background(), 1, "", []string{"art"}, 1, 1)
	if err != nil || len(filtered) != 1 || filtered[0].Image.ID != 10 {
		t.Fatalf("unexpected filtered board page: %+v, %v", filtered, err)
	}
	if _, err := repo.GetItem(context.Background(), 1, 13); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("soft-deleted image should not be returned: %v", err)
	}
	imageIDs, err := repo.ImageIDs(context.Background(), 1)
	if err != nil || len(imageIDs) != 2 || imageIDs[0] != 10 || imageIDs[1] != 11 {
		t.Fatalf("unexpected image membership: %v, %v", imageIDs, err)
	}
	boardIDs, err := repo.BoardIDsForImage(context.Background(), 10)
	if err != nil || len(boardIDs) != 1 || boardIDs[0] != 1 {
		t.Fatalf("unexpected board membership: %v, %v", boardIDs, err)
	}
	emptyIDs, err := repo.ImageIDs(context.Background(), 3)
	if err != nil || emptyIDs == nil || len(emptyIDs) != 0 {
		t.Fatalf("empty memberships should be an empty list: %v, %v", emptyIDs, err)
	}
}

func TestAddItemRejectsDuplicateImage(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TABLE board_items (id integer primary key, board_id integer, image_id integer, created_at datetime, UNIQUE(board_id, image_id))").Error; err != nil {
		t.Fatal(err)
	}
	repo := NewBoardRepository(db)
	if err := repo.AddItem(context.Background(), 1, &BoardItem{ImageID: 10}); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddItem(context.Background(), 1, &BoardItem{ImageID: 10}); !errors.Is(err, apperrors.ErrAlreadyExists) {
		t.Fatalf("duplicate image should be rejected: %v", err)
	}
}
