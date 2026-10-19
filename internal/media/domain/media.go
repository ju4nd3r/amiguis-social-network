package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidFileType = errors.New("unsupported file type. Allowed: JPEG, PNG, WebP, MP4, MOV")
	ErrFileTooLarge    = errors.New("file exceeds maximum allowed size")
)

const (
	MaxPhotoSizeBytes = 15 * 1024 * 1024  // 15 MB
	MaxVideoSizeBytes = 100 * 1024 * 1024 // 100 MB
)

type MediaKind string

const (
	MediaKindPhoto MediaKind = "photo"
	MediaKindVideo MediaKind = "video"
)

type MediaStatus string

const (
	StatusPending    MediaStatus = "pending"
	StatusProcessing MediaStatus = "processing"
	StatusReady      MediaStatus = "ready"
	StatusFailed     MediaStatus = "failed"
)

type MediaItem struct {
	ID           string      `json:"id"`
	StorageKey   string      `json:"storage_key"`
	URL          string      `json:"url"`
	ThumbnailURL *string     `json:"thumbnail_url,omitempty"`
	Kind         MediaKind   `json:"kind"`
	Width        *int        `json:"width,omitempty"`
	Height       *int        `json:"height,omitempty"`
	Duration     *float64    `json:"duration,omitempty"`
	Status       MediaStatus `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
}

func ValidateUploadRequest(filename, contentType string, sizeBytes int64) (MediaKind, error) {
	ct := strings.ToLower(contentType)
	switch {
	case strings.HasPrefix(ct, "image/jpeg"), strings.HasPrefix(ct, "image/png"), strings.HasPrefix(ct, "image/webp"):
		if sizeBytes > MaxPhotoSizeBytes {
			return "", ErrFileTooLarge
		}
		return MediaKindPhoto, nil
	case strings.HasPrefix(ct, "video/mp4"), strings.HasPrefix(ct, "video/quicktime"):
		if sizeBytes > MaxVideoSizeBytes {
			return "", ErrFileTooLarge
		}
		return MediaKindVideo, nil
	default:
		return "", ErrInvalidFileType
	}
}
