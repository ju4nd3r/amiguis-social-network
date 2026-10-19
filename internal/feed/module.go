package feed

import (
	contentDomain "amiguis/internal/content/domain"
	feedHttp "amiguis/internal/feed/adapters/http"
	"amiguis/internal/feed/adapters/postgres"
	feedApp "amiguis/internal/feed/app"
	"amiguis/internal/platform/database"
	profileApp "amiguis/internal/profiles/app"
	socialgraphApp "amiguis/internal/socialgraph/app"
)

type Module struct {
	Service *feedApp.FeedService
	Handler *feedHttp.FeedHandler
	Repo    feedApp.FeedRepository
}

func NewModule(db *database.DB, sgService *socialgraphApp.SocialGraphService, profileRepo profileApp.ProfileRepository, memoryPosts map[string]*contentDomain.Post) *Module {
	var repo feedApp.FeedRepository
	if db != nil && db.Pool != nil {
		repo = postgres.NewPostgresFeedRepository(db)
	} else {
		repo = postgres.NewMemoryFeedRepository(memoryPosts)
	}

	service := feedApp.NewFeedService(repo, sgService, profileRepo)
	handler := feedHttp.NewFeedHandler(service)

	return &Module{
		Service: service,
		Handler: handler,
		Repo:    repo,
	}
}
