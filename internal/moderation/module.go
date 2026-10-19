package moderation

import (
	moderationHttp "amiguis/internal/moderation/adapters/http"
	moderationApp "amiguis/internal/moderation/app"
)

type Module struct {
	Service *moderationApp.ModerationService
	Handler *moderationHttp.ModerationHandler
	Repo    moderationApp.ModerationRepository
}

func NewModule() *Module {
	repo := moderationApp.NewMemoryModerationRepository()
	service := moderationApp.NewModerationService(repo)
	handler := moderationHttp.NewModerationHandler(service)

	return &Module{
		Service: service,
		Handler: handler,
		Repo:    repo,
	}
}
