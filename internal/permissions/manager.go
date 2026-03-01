package permissions

import (
	"net/http"
	"sort"
	"strings"

	"cloud-webdav-server/internal/config"
)

// Manager handles path-based permission checks.
type Manager struct {
	permissions []config.FolderPermission
	roMethods   map[string]bool
	rwMethods   map[string]bool
}

// New creates a permission Manager from config.
func New(cfg *config.Config) *Manager {
	// Sort permissions by path length descending for longest-prefix match.
	perms := make([]config.FolderPermission, len(cfg.FolderPermissions))
	copy(perms, cfg.FolderPermissions)
	sort.Slice(perms, func(i, j int) bool {
		return len(perms[i].Path) > len(perms[j].Path)
	})

	roMethods := make(map[string]bool, len(cfg.ROMethods))
	for _, m := range cfg.ROMethods {
		roMethods[m] = true
	}
	rwMethods := make(map[string]bool, len(cfg.RWMethods))
	for _, m := range cfg.RWMethods {
		rwMethods[m] = true
	}

	return &Manager{
		permissions: perms,
		roMethods:   roMethods,
		rwMethods:   rwMethods,
	}
}

// CheckResult is the outcome of a permission check.
type CheckResult int

const (
	Allow      CheckResult = iota // proceed
	DenyUnauth                    // 401 – not authenticated
	DenyForbid                    // 403 – authenticated but not allowed
)

// Check evaluates whether the given user (empty = unauthenticated) may
// perform method on path. It returns the result and the matched permission.
func (m *Manager) Check(path, method, username string) CheckResult {
	perm := m.findPermission(path)
	if perm == nil {
		// No matching rule → 404-style deny without WWW-Authenticate.
		// Never send DenyUnauth here: unmatched paths (e.g. /favicon.ico)
		// must not trigger a browser auth dialog.
		return DenyForbid
	}

	isWrite := m.isWriteMethod(method)

	// Check method is globally known.
	if !m.roMethods[method] && !m.rwMethods[method] {
		return DenyForbid
	}

	// Public folder: no auth needed, but respect ro/rw.
	if m.isPublic(perm) {
		if isWrite && perm.Mode == "ro" {
			return DenyForbid
		}
		return Allow
	}

	// From here on, authentication is required.
	if username == "" {
		return DenyUnauth
	}

	// Check exclusions first.
	for _, ex := range perm.Excluded {
		if strings.EqualFold(ex, username) {
			return DenyForbid
		}
	}

	// Check if user is allowed.
	allowed := false
	for _, u := range perm.Users {
		if u == "*" || strings.EqualFold(u, username) {
			allowed = true
			break
		}
	}
	if !allowed {
		return DenyForbid
	}

	// Check method vs mode.
	if isWrite && perm.Mode == "ro" {
		return DenyForbid
	}

	return Allow
}

// RequiresAuth returns true when the path is not a public folder.
func (m *Manager) RequiresAuth(path string) bool {
	perm := m.findPermission(path)
	if perm == nil {
		return true
	}
	return !m.isPublic(perm)
}

// findPermission returns the most-specific permission matching path.
func (m *Manager) findPermission(path string) *config.FolderPermission {
	// Normalize path.
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	for i := range m.permissions {
		p := &m.permissions[i]
		prefix := p.Path
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		if path == p.Path || strings.HasPrefix(path, prefix) {
			return p
		}
	}
	return nil
}

func (m *Manager) isPublic(p *config.FolderPermission) bool {
	for _, u := range p.Users {
		if u == "public" {
			return true
		}
	}
	return false
}

func (m *Manager) isWriteMethod(method string) bool {
	return !m.roMethods[method]
}

// Authenticate header realm for a path.
func (m *Manager) Realm(path string) string {
	perm := m.findPermission(path)
	if perm == nil {
		return "WebDAV"
	}
	return "WebDAV " + perm.Path
}

// ParseBasicAuth extracts the username and password from the Authorization header.
func ParseBasicAuth(r *http.Request) (username, password string, ok bool) {
	return r.BasicAuth()
}
