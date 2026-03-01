package auth

import (
	"context"
	"errors"
	"net/http"

	"cloud-webdav-server/internal/config"
)

// ErrInvalidCredentials is returned when credentials are wrong.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Authenticator validates credentials and returns the canonical username.
type Authenticator interface {
	// Authenticate validates the request and returns the username, or an error.
	Authenticate(ctx context.Context, r *http.Request) (username string, err error)
	// Name returns a human-readable name (for logging).
	Name() string
}

// Chain tries multiple authenticators in order, returning the first success.
// If all fail, the last error is returned.
type Chain []Authenticator

func (c Chain) Authenticate(ctx context.Context, r *http.Request) (string, error) {
	var lastErr error
	for _, a := range c {
		username, err := a.Authenticate(ctx, r)
		if err == nil {
			return username, nil
		}
		lastErr = err
	}
	return "", lastErr
}

func (c Chain) Name() string { return "chain" }

// New builds the authenticator chain based on config.
func New(cfg *config.Config) Authenticator {
	var chain Chain

	if cfg.OIDCEnabled {
		chain = append(chain, NewOIDC(&cfg.OIDC))
	}
	if cfg.LDAPEnabled {
		chain = append(chain, NewLDAP(&cfg.LDAP))
	}
	if cfg.BasicAuthEnabled && len(cfg.BasicUsers) > 0 {
		chain = append(chain, NewBasic(cfg.BasicUsers))
	}

	if len(chain) == 0 {
		// No auth configured: allow everyone (guest mode).
		return &NoAuth{}
	}
	if len(chain) == 1 {
		return chain[0]
	}
	return chain
}

// NoAuth authenticator allows all requests with an empty username.
type NoAuth struct{}

func (n *NoAuth) Authenticate(_ context.Context, _ *http.Request) (string, error) {
	return "", nil
}
func (n *NoAuth) Name() string { return "none" }
