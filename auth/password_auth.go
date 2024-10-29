package auth

import (
	"app/database/db"
	"context"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// Login for password-based authentication.
type PasswordAuth struct{}

// Login for password-based authentication.
func (p *PasswordAuth) Login(ctx context.Context, user db.User, req LoginRequest) (string, error) {

	// Compare passwords
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Credential)) != nil {
		log.Printf("Err: PasswordHash: %s, req.Credential: %s", user.PasswordHash, req.Credential)
		return "", errors.New("invalid email or password")
	}

	return "", nil
}
