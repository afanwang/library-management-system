package auth

import (
	"app/database/db"
	"context"
	"encoding/hex"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang-jwt/jwt"
)

type Web3Auth struct{}

// Verify if signature is valid
func verifySignature(address, nonce, signature string) bool {
	data := []byte(nonce)
	sig, _ := hex.DecodeString(signature[2:]) // Remove "0x"

	// Recover public key from signature
	pubKey, err := crypto.SigToPub(crypto.Keccak256Hash(data).Bytes(), sig)
	if err != nil {
		return false
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey).Hex()
	return recoveredAddr == address
}

// Login for web3-based authentication.
func (w *Web3Auth) Login(ctx context.Context, user db.User, req LoginRequest) (string, error) {
	if !verifySignature(req.Address, user.Nonce, req.Signature) {
		return "", errors.New("invalid signature")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"address": user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		return "", err
	}

	// Generate and return a mock token
	return tokenString, nil
}
