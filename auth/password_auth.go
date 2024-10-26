package auth

import (
	"app/database/db"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Login for password-based authentication.
type PasswordAuth struct{}

// Login for password-based authentication.
func (p *PasswordAuth) Login(ctx context.Context, user db.User, req LoginRequest) (string, error) {
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Credential)) != nil {
		return "", errors.New("invalid email or password")
	}

	return "", nil
}
