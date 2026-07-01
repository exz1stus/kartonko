package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

type mockStorageObj struct {
	Data []byte
	Info FileInfo
}

type MockStorage struct {
	mu      sync.Mutex
	objects map[string]mockStorageObj

	FailUploadOn func(key string) error
	FailDeleteOn func(key string) error
}

func NewMockStorage() *MockStorage {
	return &MockStorage{objects: make(map[string]mockStorageObj)}
}

func (m *MockStorage) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	if m.FailUploadOn != nil {
		if err := m.FailUploadOn(key); err != nil {
			return err
		}
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	info := FileInfo{
		Key:          key,
		Size:         int64(len(data)),
		ContentType:  contentType,
		LastModified: time.Now(),
	}

	obj := mockStorageObj{
		Data: data,
		Info: info,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[key] = obj
	return nil
}

func (m *MockStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	obj, ok := m.objects[key]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return io.NopCloser(bytes.NewReader(obj.Data)), nil
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	if m.FailDeleteOn != nil {
		if err := m.FailDeleteOn(key); err != nil {
			return err
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.objects[key]; !ok {
		return fmt.Errorf("key not found: %s", key)
	}
	delete(m.objects, key)
	return nil
}

func (m *MockStorage) Stat(ctx context.Context, key string) (*FileInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	obj, ok := m.objects[key]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	return &obj.Info, nil
}

func (m *MockStorage) List(ctx context.Context, prefix string) ([]FileInfo, error) {
	var items []FileInfo
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, obj := range m.objects {
		if strings.HasPrefix(k, prefix) {
			items = append(items, obj.Info)
		}
	}

	return items, nil
}

func (m *MockStorage) Has(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.objects[key]
	return ok
}

func (m *MockStorage) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.objects)
}
