package identity

import (
	"amiguis/internal/identity/adapters/http"
	"amiguis/internal/identity/adapters/postgres"
	"amiguis/internal/identity/app"
	"amiguis/internal/platform/auth"
	"amiguis/internal/platform/database"
)

type Module struct {
	Service *app.IdentityService
	Handler *http.IdentityHandler
	Repo    app.UserRepository
}

type authServiceWrapper struct {
	secret string
}

func (w *authServiceWrapper) GenerateTokenPair(userID, username string) (*auth.TokenPair, error) {
	return auth.GenerateTokenPair(w.secret, userID, username)
}

func (w *authServiceWrapper) ValidateToken(tokenString string) (*auth.Claims, error) {
	return auth.ValidateToken(w.secret, tokenString)
}

type hasherWrapper struct{}

func (h *hasherWrapper) HashPassword(password string) (string, error) {
	return auth.HashPassword(password)
}

func (h *hasherWrapper) VerifyPassword(password, hash string) (bool, error) {
	return auth.VerifyPassword(password, hash)
}

func NewModule(db *database.DB, jwtSecret string) *Module {
	var repo app.UserRepository
	if db != nil && db.Pool != nil {
		repo = postgres.NewPostgresUserRepository(db)
	} else {
		repo = postgres.NewMemoryUserRepository()
	}

	tokenService := &authServiceWrapper{secret: jwtSecret}
	hasher := &hasherWrapper{}

	service := app.NewIdentityService(repo, hasher, tokenService)
	handler := http.NewIdentityHandler(service)

	return &Module{
		Service: service,
		Handler: handler,
		Repo:    repo,
	}
}
