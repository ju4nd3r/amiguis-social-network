package notifications

import (
	notificationsHttp "amiguis/internal/notifications/adapters/http"
	"amiguis/internal/notifications/adapters/postgres"
	notificationsApp "amiguis/internal/notifications/app"
	"amiguis/internal/platform/database"
	"amiguis/internal/platform/events"
	profileApp "amiguis/internal/profiles/app"
)

type Module struct {
	Service *notificationsApp.NotificationService
	Handler *notificationsHttp.NotificationHandler
	Repo    notificationsApp.NotificationRepository
	SSEHub  *notificationsApp.SSEHub
}

func NewModule(db *database.DB, profileRepo profileApp.ProfileRepository, bus events.Bus) *Module {
	var repo notificationsApp.NotificationRepository
	if db != nil && db.Pool != nil {
		repo = postgres.NewPostgresNotificationRepository(db)
	} else {
		repo = postgres.NewMemoryNotificationRepository()
	}

	sseHub := notificationsApp.NewSSEHub()
	service := notificationsApp.NewNotificationService(repo, profileRepo, sseHub)
	service.RegisterEventHandler(bus)
	handler := notificationsHttp.NewNotificationHandler(service, sseHub)

	return &Module{
		Service: service,
		Handler: handler,
		Repo:    repo,
		SSEHub:  sseHub,
	}
}
