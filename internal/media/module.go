package media

import (
	mediaHttp "amiguis/internal/media/adapters/http"
	"amiguis/internal/media/adapters/storage"
	mediaApp "amiguis/internal/media/app"
)

type Module struct {
	Service *mediaApp.MediaService
	Handler *mediaHttp.MediaHandler
	Repo    mediaApp.MediaRepository
	Storage *storage.LocalStorageAdapter
}

func NewModule(storageDir, storageURL string) *Module {
	store := storage.NewLocalStorageAdapter(storageDir, storageURL)
	repo := storage.NewMemoryMediaRepository()
	service := mediaApp.NewMediaService(repo, store)
	handler := mediaHttp.NewMediaHandler(service, store)

	return &Module{
		Service: service,
		Handler: handler,
		Repo:    repo,
		Storage: store,
	}
}
