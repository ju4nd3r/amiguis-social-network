package engagement

import (
	"amiguis/internal/engagement/adapters/http"
	"amiguis/internal/engagement/adapters/postgres"
	"amiguis/internal/engagement/app"
	"amiguis/internal/platform/database"
	"amiguis/internal/platform/events"
	profileApp "amiguis/internal/profiles/app"
)

type Module struct {
	Service *app.EngagementService
	Handler *http.EngagementHandler
	Repo    app.EngagementRepository
}

func NewModule(db *database.DB, profileRepo profileApp.ProfileRepository, eventBus events.Bus) *Module {
	var repo app.EngagementRepository
	if db != nil && db.Pool != nil {
		repo = postgres.NewPostgresEngagementRepository(db)
	} else {
		repo = postgres.NewMemoryEngagementRepository()
	}

	service := app.NewEngagementService(repo, profileRepo, eventBus)
	handler := http.NewEngagementHandler(service)

	return &Module{
		Service: service,
		Handler: handler,
		Repo:    repo,
	}
}
