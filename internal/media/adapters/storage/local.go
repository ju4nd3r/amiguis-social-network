package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"amiguis/internal/media/domain"
)

type LocalStorageAdapter struct {
	baseDir string
	baseURL string
}

func NewLocalStorageAdapter(baseDir, baseURL string) *LocalStorageAdapter {
	_ = os.MkdirAll(baseDir, 0755)
	return &LocalStorageAdapter{
		baseDir: baseDir,
		baseURL: baseURL,
	}
}

func (a *LocalStorageAdapter) GeneratePresignedURL(ctx context.Context, key, contentType string) (string, error) {
	// For local dev, returns a direct upload endpoint
	return fmt.Sprintf("%s/upload/%s", a.baseURL, key), nil
}

func (a *LocalStorageAdapter) GetPublicURL(key string) string {
	return fmt.Sprintf("%s/%s", a.baseURL, key)
}

func (a *LocalStorageAdapter) SaveFile(key string, data []byte) error {
	fullPath := filepath.Join(a.baseDir, key)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, data, 0644)
}

type MemoryMediaRepository struct {
	mu    sync.RWMutex
	items map[string]*domain.MediaItem
}

func NewMemoryMediaRepository() *MemoryMediaRepository {
	return &MemoryMediaRepository{
		items: make(map[string]*domain.MediaItem),
	}
}

func (r *MemoryMediaRepository) Save(ctx context.Context, item *domain.MediaItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = item
	return nil
}

func (r *MemoryMediaRepository) FindByID(ctx context.Context, id string) (*domain.MediaItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok {
		return nil, fmt.Errorf("media not found: %s", id)
	}
	return item, nil
}

func (r *MemoryMediaRepository) UpdateStatus(ctx context.Context, id string, status domain.MediaStatus, url string, thumb *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if item, ok := r.items[id]; ok {
		item.Status = status
		item.URL = url
		item.ThumbnailURL = thumb
	}
	return nil
}
