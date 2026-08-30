package repository_integration_tests

import "gorm.io/gorm"

type RepoTestContext struct {
	Db *gorm.DB
}
