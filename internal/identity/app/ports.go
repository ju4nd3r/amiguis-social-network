package app

import (
	"context"
	"errors"
	"time"

	"amiguis/internal/identity/domain"
	"amiguis/internal/platform/auth"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query, city string, limit int) ([]*domain.User, error)
}

type TokenService interface {
	GenerateTokenPair(userID, username string) (*auth.TokenPair, error)
	ValidateToken(tokenString string) (*auth.Claims, error)
}

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) (bool, error)
}

type RegisterInput struct {
	Email       string
	Username    string
	Password    string
	DisplayName string
	City        string
}

type AuthOutput struct {
	Tokens *auth.TokenPair
	User   *domain.User
}

type IdentityService struct {
	repo   UserRepository
	hasher PasswordHasher
	tokens TokenService
}

func NewIdentityService(repo UserRepository, hasher PasswordHasher, tokens TokenService) *IdentityService {
	return &IdentityService{
		repo:   repo,
		hasher: hasher,
		tokens: tokens,
	}
}

func (s *IdentityService) Register(ctx context.Context, in RegisterInput) (*AuthOutput, error) {
	if len(in.Password) < 8 {
		return nil, domain.ErrPasswordTooShort
	}

	if existing, _ := s.repo.FindByEmail(ctx, in.Email); existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}
	if existing, _ := s.repo.FindByUsername(ctx, in.Username); existing != nil {
		return nil, domain.ErrUsernameAlreadyExists
	}

	pwdHash, err := s.hasher.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	userID := uuid.New().String()
	user, err := domain.NewUser(userID, in.Email, in.Username, pwdHash, in.DisplayName, in.City)
	if err != nil {
		return nil, err
	}

	// Auto-verify for initial experience or allow verify email flow
	now := time.Now().UTC()
	user.VerifiedAt = &now

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	tokens, err := s.tokens.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &AuthOutput{
		Tokens: tokens,
		User:   user,
	}, nil
}

func (s *IdentityService) Login(ctx context.Context, login, password string) (*AuthOutput, error) {
	var user *domain.User
	var err error

	user, err = s.repo.FindByEmail(ctx, login)
	if err != nil || user == nil {
		user, err = s.repo.FindByUsername(ctx, login)
	}

	if err != nil || user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	valid, err := s.hasher.VerifyPassword(password, user.PasswordHash)
	if err != nil || !valid {
		return nil, domain.ErrInvalidCredentials
	}

	tokens, err := s.tokens.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &AuthOutput{
		Tokens: tokens,
		User:   user,
	}, nil
}

func (s *IdentityService) Refresh(ctx context.Context, refreshToken string) (*AuthOutput, error) {
	claims, err := s.tokens.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil || user == nil {
		return nil, domain.ErrUserNotFound
	}

	tokens, err := s.tokens.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &AuthOutput{
		Tokens: tokens,
		User:   user,
	}, nil
}

func (s *IdentityService) GetCurrentUser(ctx context.Context, userID string) (*domain.User, error) {
	return s.repo.FindByID(ctx, userID)
}

func (s *IdentityService) DeleteAccount(ctx context.Context, userID, confirmPassword string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return domain.ErrUserNotFound
	}

	valid, err := s.hasher.VerifyPassword(confirmPassword, user.PasswordHash)
	if err != nil || !valid {
		return domain.ErrInvalidCredentials
	}

	return s.repo.Delete(ctx, userID)
}
