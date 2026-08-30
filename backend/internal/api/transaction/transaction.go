package transaction

import (
	"context"

	"gorm.io/gorm"
)

type Runner interface {
	Within(ctx context.Context, fn func(*gorm.DB) error) error
}

type GormRunner struct {
	DB *gorm.DB
}

func (r GormRunner) Within(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.DB.WithContext(ctx).Transaction(fn)
}

type TestRunner struct{}

func (r TestRunner) Within(ctx context.Context, fn func(*gorm.DB) error) error {
	return fn(nil)
}
