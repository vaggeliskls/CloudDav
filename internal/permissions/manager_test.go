package permissions

import (
	"testing"

	"cloud-webdav-server/internal/config"
)

func newManager(perms []config.FolderPermission) *Manager {
	cfg := &config.Config{
		FolderPermissions: perms,
		ROMethods:         []string{"GET", "HEAD", "OPTIONS", "PROPFIND"},
		RWMethods:         []string{"GET", "HEAD", "OPTIONS", "PROPFIND", "PUT", "DELETE", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK", "PROPPATCH"},
	}
	return New(cfg)
}

func TestCheck_PublicFolder(t *testing.T) {
	m := newManager([]config.FolderPermission{
		{Path: "/public", Users: []string{"public"}, Mode: "rw"},
	})

	if got := m.Check("/public/", "GET", ""); got != Allow {
		t.Errorf("expected Allow, got %v", got)
	}
	if got := m.Check("/public/file.txt", "GET", ""); got != Allow {
		t.Errorf("expected Allow for unauthenticated GET, got %v", got)
	}
	if got := m.Check("/public/file.txt", "PUT", ""); got != Allow {
		t.Errorf("expected Allow for unauthenticated PUT on rw, got %v", got)
	}
}

func TestCheck_PublicFolderReadOnly(t *testing.T) {
	m := newManager([]config.FolderPermission{
		{Path: "/public", Users: []string{"public"}, Mode: "ro"},
	})

	if got := m.Check("/public/file.txt", "GET", ""); got != Allow {
		t.Errorf("expected Allow for GET on ro public, got %v", got)
	}
	if got := m.Check("/public/file.txt", "PUT", ""); got != DenyForbid {
		t.Errorf("expected DenyForbid for PUT on ro public, got %v", got)
	}
}

func TestCheck_AuthRequired_NoCredentials(t *testing.T) {
	m := newManager([]config.FolderPermission{
		{Path: "/private", Users: []string{"alice"}, Mode: "rw"},
	})

	if got := m.Check("/private/file.txt", "GET", ""); got != DenyUnauth {
		t.Errorf("expected DenyUnauth for unauthenticated request, got %v", got)
	}
}

func TestCheck_AuthRequired_WrongUser(t *testing.T) {
	m := newManager([]config.FolderPermission{
		{Path: "/private", Users: []string{"alice"}, Mode: "rw"},
	})

	if got := m.Check("/private/file.txt", "GET", "bob"); got != DenyForbid {
		t.Errorf("expected DenyForbid for wrong user, got %v", got)
	}
}

func TestCheck_AuthRequired_CorrectUser(t *testing.T) {
	m := newManager([]config.FolderPermission{
		{Path: "/private", Users: []string{"alice"}, Mode: "rw"},
	})

	if got := m.Check("/private/file.txt", "GET", "alice"); got != Allow {
		t.Errorf("expected Allow for correct user, got %v", got)
	}
}

func TestCheck_Wildcard_User(t *testing.T) {
	m := newManager([]config.FolderPermission{
		{Path: "/shared", Users: []string{"*"}, Mode: "rw"},
	})

	if got := m.Check("/shared/file.txt", "GET", "anyone"); got != Allow {
		t.Errorf("expected Allow for wildcard user, got %v", got)
	}
}

func TestCheck_ExcludedUser(t *testing.T) {
	m := newManager([]config.FolderPermission{
		{Path: "/shared", Users: []string{"*"}, Excluded: []string{"banned"}, Mode: "rw"},
	})

	if got := m.Check("/shared/file.txt", "GET", "banned"); got != DenyForbid {
		t.Errorf("expected DenyForbid for excluded user, got %v", got)
	}
	if got := m.Check("/shared/file.txt", "GET", "alice"); got != Allow {
		t.Errorf("expected Allow for non-excluded user, got %v", got)
	}
}

func TestCheck_NoMatchingPath(t *testing.T) {
	m := newManager([]config.FolderPermission{
		{Path: "/public", Users: []string{"public"}, Mode: "rw"},
	})

	if got := m.Check("/favicon.ico", "GET", ""); got != DenyForbid {
		t.Errorf("expected DenyForbid for unmatched path, got %v", got)
	}
}

func TestCheck_LongestPrefixMatch(t *testing.T) {
	m := newManager([]config.FolderPermission{
		{Path: "/", Users: []string{"public"}, Mode: "ro"},
		{Path: "/private", Users: []string{"alice"}, Mode: "rw"},
	})

	// /private should match the more specific rule, not root.
	if got := m.Check("/private/file.txt", "GET", ""); got != DenyUnauth {
		t.Errorf("expected DenyUnauth (specific rule wins), got %v", got)
	}
	// Root should match the root rule.
	if got := m.Check("/readme.txt", "GET", ""); got != Allow {
		t.Errorf("expected Allow for root public rule, got %v", got)
	}
}

func TestCheck_ReadOnlyUser_CannotWrite(t *testing.T) {
	m := newManager([]config.FolderPermission{
		{Path: "/files", Users: []string{"alice"}, Mode: "ro"},
	})

	if got := m.Check("/files/doc.txt", "PUT", "alice"); got != DenyForbid {
		t.Errorf("expected DenyForbid for PUT on ro folder, got %v", got)
	}
	if got := m.Check("/files/doc.txt", "GET", "alice"); got != Allow {
		t.Errorf("expected Allow for GET on ro folder, got %v", got)
	}
}

func TestCheck_UnknownMethod(t *testing.T) {
	m := newManager([]config.FolderPermission{
		{Path: "/files", Users: []string{"public"}, Mode: "rw"},
	})

	if got := m.Check("/files/", "PATCH", ""); got != DenyForbid {
		t.Errorf("expected DenyForbid for unknown method, got %v", got)
	}
}

func TestCheck_CaseInsensitiveUsername(t *testing.T) {
	m := newManager([]config.FolderPermission{
		{Path: "/files", Users: []string{"Alice"}, Mode: "rw"},
	})

	if got := m.Check("/files/doc.txt", "GET", "alice"); got != Allow {
		t.Errorf("expected Allow for case-insensitive username match, got %v", got)
	}
}

func TestRequiresAuth(t *testing.T) {
	m := newManager([]config.FolderPermission{
		{Path: "/public", Users: []string{"public"}, Mode: "ro"},
		{Path: "/private", Users: []string{"alice"}, Mode: "rw"},
	})

	if m.RequiresAuth("/public/file") {
		t.Error("expected /public to not require auth")
	}
	if !m.RequiresAuth("/private/file") {
		t.Error("expected /private to require auth")
	}
}
