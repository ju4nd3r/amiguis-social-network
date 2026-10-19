package profiles

import (
	"amiguis/internal/identity/app"
	"amiguis/internal/platform/database"
	"amiguis/internal/profiles/adapters/http"
	"amiguis/internal/profiles/adapters/postgres"
	profileApp "amiguis/internal/profiles/app"
)

type Module struct {
	Service *profileApp.ProfileService
	Handler *http.ProfileHandler
	Repo    profileApp.ProfileRepository
}

func NewModule(db *database.DB, identityRepo app.UserRepository) *Module {
	var repo profileApp.ProfileRepository
	if db != nil && db.Pool != nil {
		repo = postgres.NewPostgresProfileRepository(db)
	} else {
		repo = postgres.NewMemoryProfileRepository(identityRepo)
	}

	service := profileApp.NewProfileService(repo)
	handler := http.NewProfileHandler(service)

	return &Module{
		Service: service,
		Handler: handler,
		Repo:    repo,
	}
}
