package auth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// BasicAuth validates HTTP Basic credentials against a static user map.
// Passwords in the map are stored as bcrypt hashes.
type BasicAuth struct {
	// hashed stores username -> bcrypt-hashed password
	hashed map[string]string
}

// NewBasic creates a BasicAuth from a map of username -> plaintext password.
// Passwords are hashed with bcrypt at construction time.
func NewBasic(users map[string]string) *BasicAuth {
	hashed := make(map[string]string, len(users))
	for u, p := range users {
		h, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
		if err != nil {
			panic(fmt.Sprintf("auth/basic: bcrypt hash for user %q: %v", u, err))
		}
		hashed[u] = string(h)
	}
	return &BasicAuth{hashed: hashed}
}

func (b *BasicAuth) Authenticate(_ context.Context, r *http.Request) (string, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return "", ErrInvalidCredentials
	}
	hash, exists := b.hashed[username]
	if !exists {
		return "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return username, nil
}

func (b *BasicAuth) Name() string { return "basic" }
