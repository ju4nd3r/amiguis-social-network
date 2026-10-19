package socialgraph

import (
	"amiguis/internal/platform/database"
	"amiguis/internal/platform/events"
	profileApp "amiguis/internal/profiles/app"
	"amiguis/internal/socialgraph/adapters/http"
	"amiguis/internal/socialgraph/adapters/postgres"
	"amiguis/internal/socialgraph/app"
)

type Module struct {
	Service *app.SocialGraphService
	Handler *http.SocialGraphHandler
	Repo    app.SocialGraphRepository
}

func NewModule(db *database.DB, profileRepo profileApp.ProfileRepository, eventBus events.Bus) *Module {
	var repo app.SocialGraphRepository
	if db != nil && db.Pool != nil {
		repo = postgres.NewPostgresSocialGraphRepository(db, profileRepo)
	} else {
		repo = postgres.NewMemorySocialGraphRepository(profileRepo)
	}

	service := app.NewSocialGraphService(repo, eventBus)
	handler := http.NewSocialGraphHandler(service)

	return &Module{
		Service: service,
		Handler: handler,
		Repo:    repo,
	}
}
