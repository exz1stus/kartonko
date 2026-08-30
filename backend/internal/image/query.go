package image

import (
	"server/internal/user"
)

// Query represents an image search query
type Query struct {
	Prefix string
	Tags   []string
	User   *user.User
	Cursor int
	Limit  int
}

// QueryBuilder builds image queries
type QueryBuilder struct {
	query Query
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		query: Query{
			Cursor: 0,
			Limit:  0,
		},
	}
}

func (b *QueryBuilder) Prefix(prefix string) *QueryBuilder {
	b.query.Prefix = prefix
	return b
}

func (b *QueryBuilder) Tags(tags []string) *QueryBuilder {
	b.query.Tags = tags
	return b
}

func (b *QueryBuilder) User(user *user.User) *QueryBuilder {
	b.query.User = user
	return b
}

func (b *QueryBuilder) Cursor(cursor int) *QueryBuilder {
	b.query.Cursor = cursor
	return b
}

func (b *QueryBuilder) Limit(limit int) *QueryBuilder {
	b.query.Limit = limit
	return b
}

func (b *QueryBuilder) Build() *Query {
	return &b.query
}
