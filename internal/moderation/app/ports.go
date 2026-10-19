package app

import (
	"context"
	"time"

	"amiguis/internal/moderation/domain"

	"github.com/google/uuid"
)

type ModerationRepository interface {
	CreateReport(ctx context.Context, report *domain.Report) error
	GetReports(ctx context.Context, status string, limit int) ([]*domain.Report, error)
	ResolveReport(ctx context.Context, id string, action string) error
}

type ModerationService struct {
	repo ModerationRepository
}

func NewModerationService(repo ModerationRepository) *ModerationService {
	return &ModerationService{repo: repo}
}

func (s *ModerationService) CreateReport(ctx context.Context, reporterID string, targetType domain.TargetType, targetID, reason, details string) (*domain.Report, error) {
	id := uuid.New().String()
	report, err := domain.NewReport(id, reporterID, targetType, targetID, reason, details)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateReport(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

func (s *ModerationService) ListReports(ctx context.Context, status string, limit int) ([]*domain.Report, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.GetReports(ctx, status, limit)
}

func (s *ModerationService) ResolveReport(ctx context.Context, id string, action string) error {
	return s.repo.ResolveReport(ctx, id, action)
}

type MemoryModerationRepository struct {
	reports map[string]*domain.Report
}

func NewMemoryModerationRepository() *MemoryModerationRepository {
	return &MemoryModerationRepository{
		reports: make(map[string]*domain.Report),
	}
}

func (r *MemoryModerationRepository) CreateReport(ctx context.Context, report *domain.Report) error {
	r.reports[report.ID] = report
	return nil
}

func (r *MemoryModerationRepository) GetReports(ctx context.Context, status string, limit int) ([]*domain.Report, error) {
	var list []*domain.Report
	for _, rep := range r.reports {
		if status == "" || string(rep.Status) == status {
			list = append(list, rep)
		}
	}
	return list, nil
}

func (r *MemoryModerationRepository) ResolveReport(ctx context.Context, id string, action string) error {
	if rep, ok := r.reports[id]; ok {
		now := time.Now().UTC()
		rep.ResolvedAt = &now
		if action == "takedown" {
			rep.Status = domain.StatusResolved
		} else {
			rep.Status = domain.StatusDismissed
		}
	}
	return nil
}
