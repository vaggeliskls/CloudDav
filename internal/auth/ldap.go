package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	ldap "github.com/go-ldap/ldap/v3"

	"cloud-webdav-server/internal/config"
)

// LDAPAuth validates HTTP Basic credentials against an LDAP/AD directory.
type LDAPAuth struct {
	cfg *config.LDAPConfig
}

func NewLDAP(cfg *config.LDAPConfig) *LDAPAuth {
	return &LDAPAuth{cfg: cfg}
}

func (l *LDAPAuth) Authenticate(_ context.Context, r *http.Request) (string, error) {
	username, password, ok := r.BasicAuth()
	if !ok || username == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	conn, err := l.connect()
	if err != nil {
		return "", fmt.Errorf("ldap connect: %w", err)
	}
	defer conn.Close()

	// Bind with service account to search for the user.
	if l.cfg.BindDN != "" {
		if err := conn.Bind(l.cfg.BindDN, l.cfg.BindPassword); err != nil {
			return "", fmt.Errorf("ldap service bind: %w", err)
		}
	}

	// Search for the user DN.
	filter := fmt.Sprintf("(%s=%s)", ldap.EscapeFilter(l.cfg.Attribute), ldap.EscapeFilter(username))
	req := ldap.NewSearchRequest(
		l.cfg.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1, // size limit
		0, // time limit
		false,
		filter,
		[]string{"dn"},
		nil,
	)
	sr, err := conn.Search(req)
	if err != nil {
		return "", fmt.Errorf("ldap search: %w", err)
	}
	if len(sr.Entries) == 0 {
		return "", ErrInvalidCredentials
	}

	userDN := sr.Entries[0].DN

	// Attempt to bind as the user to verify the password.
	if err := conn.Bind(userDN, password); err != nil {
		return "", ErrInvalidCredentials
	}

	return username, nil
}

func (l *LDAPAuth) connect() (*ldap.Conn, error) {
	var conn *ldap.Conn
	var err error

	if l.cfg.StartTLS {
		conn, err = ldap.Dial("tcp", l.cfg.URL)
		if err != nil {
			return nil, err
		}
		if err := conn.StartTLS(&tls.Config{InsecureSkipVerify: false}); err != nil {
			conn.Close()
			return nil, err
		}
	} else {
		conn, err = ldap.DialURL(l.cfg.URL)
		if err != nil {
			return nil, err
		}
	}
	return conn, nil
}

func (l *LDAPAuth) Name() string { return "ldap" }
