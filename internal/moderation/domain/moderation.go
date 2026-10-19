package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidTargetType = errors.New("target type must be 'post', 'user', or 'comment'")
	ErrEmptyReason       = errors.New("report reason cannot be empty")
)

type TargetType string

const (
	TargetTypePost    TargetType = "post"
	TargetTypeUser    TargetType = "user"
	TargetTypeComment TargetType = "comment"
)

type ReportStatus string

const (
	StatusPending   ReportStatus = "pending"
	StatusReviewed  ReportStatus = "reviewed"
	StatusResolved  ReportStatus = "resolved"
	StatusDismissed ReportStatus = "dismissed"
)

type Report struct {
	ID         string       `json:"id"`
	ReporterID string       `json:"reporter_id"`
	TargetType TargetType   `json:"target_type"`
	TargetID   string       `json:"target_id"`
	Reason     string       `json:"reason"`
	Details    string       `json:"details,omitempty"`
	Status     ReportStatus `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
	ResolvedAt *time.Time   `json:"resolved_at,omitempty"`
}

func NewReport(id, reporterID string, targetType TargetType, targetID, reason, details string) (*Report, error) {
	if targetType != TargetTypePost && targetType != TargetTypeUser && targetType != TargetTypeComment {
		return nil, ErrInvalidTargetType
	}
	cleanReason := strings.TrimSpace(reason)
	if cleanReason == "" {
		return nil, ErrEmptyReason
	}

	return &Report{
		ID:         id,
		ReporterID: reporterID,
		TargetType: targetType,
		TargetID:   targetID,
		Reason:     cleanReason,
		Details:    strings.TrimSpace(details),
		Status:     StatusPending,
		CreatedAt:  time.Now().UTC(),
	}, nil
}
