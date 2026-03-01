package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"cloud-webdav-server/internal/config"
)

// OIDCAuth validates OAuth2 Bearer tokens via an OIDC provider.
// It also supports the authorization-code flow for browser clients.
type OIDCAuth struct {
	cfg      *config.OIDCConfig
	provider *gooidc.Provider
	verifier *gooidc.IDTokenVerifier
	oauth2   oauth2.Config
}

// NewOIDC creates an OIDCAuth. The provider is initialized lazily on first use
// to avoid failing at startup if the OIDC provider is not yet reachable.
func NewOIDC(cfg *config.OIDCConfig) *OIDCAuth {
	return &OIDCAuth{cfg: cfg}
}

func (o *OIDCAuth) init(ctx context.Context) error {
	if o.provider != nil {
		return nil
	}
	provider, err := gooidc.NewProvider(ctx, o.cfg.ProviderURL)
	if err != nil {
		return fmt.Errorf("oidc provider init: %w", err)
	}
	o.provider = provider
	o.verifier = provider.Verifier(&gooidc.Config{ClientID: o.cfg.ClientID})
	o.oauth2 = oauth2.Config{
		ClientID:     o.cfg.ClientID,
		ClientSecret: o.cfg.ClientSecret,
		RedirectURL:  o.cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       o.cfg.Scopes,
	}
	return nil
}

// Authenticate expects a Bearer token in the Authorization header.
// Browser-based flows (authorization-code) must be handled separately via
// the redirect and callback endpoints registered in RegisterHandlers.
func (o *OIDCAuth) Authenticate(ctx context.Context, r *http.Request) (string, error) {
	if err := o.init(ctx); err != nil {
		return "", err
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrInvalidCredentials
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", ErrInvalidCredentials
	}
	rawToken := parts[1]

	idToken, err := o.verifier.Verify(ctx, rawToken)
	if err != nil {
		return "", fmt.Errorf("oidc token verify: %w", err)
	}

	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return "", fmt.Errorf("oidc claims: %w", err)
	}

	claim := o.cfg.UsernameClaim
	if claim == "" {
		claim = "preferred_username"
	}
	usernameRaw, ok := claims[claim]
	if !ok {
		return "", fmt.Errorf("oidc: claim %q not found in token", claim)
	}
	username, ok := usernameRaw.(string)
	if !ok || username == "" {
		return "", fmt.Errorf("oidc: claim %q is not a string", claim)
	}

	return username, nil
}

// AuthCodeURL returns the URL to redirect the browser to for login.
func (o *OIDCAuth) AuthCodeURL(ctx context.Context, state string) (string, error) {
	if err := o.init(ctx); err != nil {
		return "", err
	}
	return o.oauth2.AuthCodeURL(state), nil
}

// Exchange handles the OIDC callback and returns the username from the token.
func (o *OIDCAuth) Exchange(ctx context.Context, code string) (string, error) {
	if err := o.init(ctx); err != nil {
		return "", err
	}
	token, err := o.oauth2.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("oidc exchange: %w", err)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return "", fmt.Errorf("oidc: no id_token in response")
	}
	idToken, err := o.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", fmt.Errorf("oidc verify: %w", err)
	}
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return "", err
	}
	claim := o.cfg.UsernameClaim
	if claim == "" {
		claim = "preferred_username"
	}
	username, _ := claims[claim].(string)
	return username, nil
}

func (o *OIDCAuth) Name() string { return "oidc" }
