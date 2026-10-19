package content

import (
	"amiguis/internal/content/adapters/http"
	"amiguis/internal/content/adapters/postgres"
	"amiguis/internal/content/app"
	"amiguis/internal/content/domain"
	"amiguis/internal/platform/database"
	"amiguis/internal/platform/events"
	profileApp "amiguis/internal/profiles/app"
)

type Module struct {
	Service *app.ContentService
	Handler *http.ContentHandler
	Repo    app.PostRepository
}

func NewModule(db *database.DB, profileRepo profileApp.ProfileRepository, eventBus events.Bus, memoryPosts map[string]*domain.Post) *Module {
	var repo app.PostRepository
	if db != nil && db.Pool != nil {
		repo = postgres.NewPostgresPostRepository(db)
	} else {
		repo = postgres.NewMemoryPostRepository(memoryPosts)
	}

	service := app.NewContentService(repo, profileRepo, eventBus)
	handler := http.NewContentHandler(service)

	return &Module{
		Service: service,
		Handler: handler,
		Repo:    repo,
	}
}
