package domain

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrEmailAlreadyExists    = errors.New("email is already registered")
	ErrUsernameAlreadyExists = errors.New("username is already taken")
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrInvalidEmail          = errors.New("invalid email address format")
	ErrInvalidUsername       = errors.New("username must be 3-30 characters and alphanumeric or underscore")
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters long")
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

type User struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	DisplayName  string
	City         string
	VerifiedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(id, email, username, passwordHash, displayName, city string) (*User, error) {
	cleanEmail := strings.TrimSpace(strings.ToLower(email))
	if _, err := mail.ParseAddress(cleanEmail); err != nil {
		return nil, ErrInvalidEmail
	}

	cleanUsername := strings.TrimSpace(strings.ToLower(username))
	if !usernameRegex.MatchString(cleanUsername) {
		return nil, ErrInvalidUsername
	}

	cleanDisplayName := strings.TrimSpace(displayName)
	if cleanDisplayName == "" {
		cleanDisplayName = cleanUsername
	}

	now := time.Now().UTC()
	return &User{
		ID:           id,
		Email:        cleanEmail,
		Username:     cleanUsername,
		PasswordHash: passwordHash,
		DisplayName:  cleanDisplayName,
		City:         strings.TrimSpace(city),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (u *User) IsVerified() bool {
	return u.VerifiedAt != nil
}

func (u *User) Verify() {
	now := time.Now().UTC()
	u.VerifiedAt = &now
	u.UpdatedAt = now
}
