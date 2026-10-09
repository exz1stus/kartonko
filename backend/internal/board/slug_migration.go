package board

import (
	"strconv"

	"gorm.io/gorm"
)

// EnsureSlugs gives existing boards stable URLs after the slug column is added.
func EnsureSlugs(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var boards []Board
		if err := tx.Unscoped().Select("id, name, slug").Order("id ASC").Find(&boards).Error; err != nil {
			return err
		}
		used := make(map[string]bool, len(boards))
		for _, board := range boards {
			if board.Slug != "" {
				used[board.Slug] = true
			}
		}
		for _, board := range boards {
			if board.Slug != "" {
				continue
			}
			base := slugForName(board.Name)
			slug := base
			for suffix := 2; used[slug]; suffix++ {
				slug = base + "-" + strconv.Itoa(suffix)
			}
			if err := tx.Unscoped().Model(&Board{}).Where("id = ?", board.ID).UpdateColumn("slug", slug).Error; err != nil {
				return err
			}
			used[slug] = true
		}
		return nil
	})
}
