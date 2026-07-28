package storage

type TestStorage interface {
	Storage
	Count(prefix string) (int, error)
}
