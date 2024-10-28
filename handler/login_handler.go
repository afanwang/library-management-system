package handler

import (
	"app/auth"
	"app/database/adaptor"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

const (
	RoleKey = "roleKey"
)

// Adapt function to convert httprouter.Handle to http.HandlerFunc
func Adapt(handler httprouter.Handle) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		params := httprouter.ParamsFromContext(r.Context())
		handler(w, r, params)
	}
}

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenStr := cookie.Value
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return auth.JwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Store role in context for later use
		ctx := context.WithValue(r.Context(), RoleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoginHandler handles the login request.
type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // admin or user
}

func LoginHandler(dbc *adaptor.PostgresClient, a auth.Authenticator, log *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		handleLogin(dbc, a, w, r)
	}
}

func handleLogin(db *adaptor.PostgresClient, a auth.Authenticator, w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	_, err = a.Login(r.Context(), user, req)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create JWT token with user role and email
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email: user.Email,
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	cookieToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := cookieToken.SignedString(auth.JwtKey)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "cookieToken",
		Value:   tokenString,
		Expires: expirationTime,
	})

	w.WriteHeader(http.StatusOK)
}

// RegisterHandler handles user registration.
func RegisterHandler(dbc *adaptor.PostgresClient, log *log.Logger) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		handleRegister(dbc, w, r)
	}
}

func handleRegister(db *adaptor.PostgresClient, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Basic validation
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Email) == "" || len(req.Role) == 0 {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Insert user into the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert user into the database
	err = db.CreateUser(ctx, req.Name, req.Email, req.Role, string(hashedPassword))
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Respond with the created user details
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
