package auth

import (
	"app/database/db"
	"context"
)

var JwtKey = []byte("your-secret-key")

type LoginRequest struct {
	// For email-based authentication
	Email      string `json:"email"`
	Credential string `json:"credential"`
	// For web3-based authentication
	Address   string `json:"address"`
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
}

// Authenticator interface supports multiple authentication strategies.
type Authenticator interface {
	Login(ctx context.Context, user db.User, req LoginRequest) (string, error) // Returns token or error
}

func NewAuthenticator(authType string) Authenticator {
	switch authType {
	case "Password":
		return &PasswordAuth{}
	case "Web3":
		return &Web3Auth{}
	default:
		return nil
	}
}
