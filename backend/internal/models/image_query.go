package models

type ImageQuery struct {
	Prefix string
	Tags   []string
	User   *User
	Cursor int
	Limit  int
}

type ImageQueryBuilder struct {
	query ImageQuery
}

func NewImageQueryBuilder() *ImageQueryBuilder {
	return &ImageQueryBuilder{
		query: ImageQuery{
			Cursor: 0,
			Limit:  0,
		},
	}
}

func (b *ImageQueryBuilder) Prefix(prefix string) *ImageQueryBuilder {
	b.query.Prefix = prefix
	return b
}

func (b *ImageQueryBuilder) Tags(tags []string) *ImageQueryBuilder {
	b.query.Tags = tags
	return b
}

func (b *ImageQueryBuilder) User(user *User) *ImageQueryBuilder {
	b.query.User = user
	return b
}

func (b *ImageQueryBuilder) Cursor(cursor int) *ImageQueryBuilder {
	b.query.Cursor = cursor
	return b
}

func (b *ImageQueryBuilder) Limit(limit int) *ImageQueryBuilder {
	b.query.Limit = limit
	return b
}

func (b *ImageQueryBuilder) Build() *ImageQuery {
	return &b.query
}
