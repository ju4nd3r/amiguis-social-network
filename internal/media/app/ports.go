package app

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"amiguis/internal/media/domain"

	"github.com/google/uuid"
)

type MediaRepository interface {
	Save(ctx context.Context, item *domain.MediaItem) error
	FindByID(ctx context.Context, id string) (*domain.MediaItem, error)
	UpdateStatus(ctx context.Context, id string, status domain.MediaStatus, url string, thumb *string) error
}

type StorageAdapter interface {
	GeneratePresignedURL(ctx context.Context, key, contentType string) (string, error)
	GetPublicURL(key string) string
}

type UploadTicket struct {
	MediaID    string `json:"media_id"`
	StorageKey string `json:"storage_key"`
	UploadURL  string `json:"upload_url"`
}

type MediaService struct {
	repo    MediaRepository
	storage StorageAdapter
}

func NewMediaService(repo MediaRepository, storage StorageAdapter) *MediaService {
	return &MediaService{
		repo:    repo,
		storage: storage,
	}
}

func (s *MediaService) RequestPresignedUpload(ctx context.Context, filename, contentType string, sizeBytes int64) (*UploadTicket, error) {
	kind, err := domain.ValidateUploadRequest(filename, contentType, sizeBytes)
	if err != nil {
		return nil, err
	}

	mediaID := uuid.New().String()
	ext := filepath.Ext(filename)
	storageKey := fmt.Sprintf("%s/%s%s", kind, mediaID, ext)

	uploadURL, err := s.storage.GeneratePresignedURL(ctx, storageKey, contentType)
	if err != nil {
		return nil, err
	}

	publicURL := s.storage.GetPublicURL(storageKey)

	item := &domain.MediaItem{
		ID:         mediaID,
		StorageKey: storageKey,
		URL:        publicURL,
		Kind:       kind,
		Status:     domain.StatusReady, // Presigned/ready
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.repo.Save(ctx, item); err != nil {
		return nil, err
	}

	return &UploadTicket{
		MediaID:    mediaID,
		StorageKey: storageKey,
		UploadURL:  uploadURL,
	}, nil
}

func (s *MediaService) GetMediaStatus(ctx context.Context, id string) (*domain.MediaItem, error) {
	return s.repo.FindByID(ctx, id)
}
