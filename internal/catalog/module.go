package catalog

import (
	"amiguis/internal/catalog/adapters/http"
	"amiguis/internal/catalog/app"
	contentApp "amiguis/internal/content/app"
)

type Module struct {
	Service *app.CatalogService
	Handler *http.CatalogHandler
}

func NewModule(contentService *contentApp.ContentService) *Module {
	service := app.NewCatalogService(contentService)
	handler := http.NewCatalogHandler(service)
	return &Module{
		Service: service,
		Handler: handler,
	}
}
