package models

type ImageQuery struct {
	Prefix string
	Tags   []string
	User   *User
}

type ImageQueryBuilder struct {
	query ImageQuery
}

func NewImageQueryBuilder() *ImageQueryBuilder {
	return &ImageQueryBuilder{}
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

func (b *ImageQueryBuilder) Build() *ImageQuery {
	return &b.query
}
