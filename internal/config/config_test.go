package config

import (
	"testing"
)

func TestParseFolderPermissions_Valid(t *testing.T) {
	perms, err := parseFolderPermissions("/public:public:ro,/files:alice bob:rw,/admin:admin:rw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(perms) != 3 {
		t.Fatalf("expected 3 perms, got %d", len(perms))
	}

	if perms[0].Path != "/public" || perms[0].Mode != "ro" || len(perms[0].Users) != 1 || perms[0].Users[0] != "public" {
		t.Errorf("unexpected first perm: %+v", perms[0])
	}
	if perms[1].Path != "/files" || perms[1].Mode != "rw" || len(perms[1].Users) != 2 {
		t.Errorf("unexpected second perm: %+v", perms[1])
	}
}

func TestParseFolderPermissions_Excluded(t *testing.T) {
	perms, err := parseFolderPermissions("/shared:* !banned:rw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(perms[0].Users) != 1 || perms[0].Users[0] != "*" {
		t.Errorf("expected wildcard user, got %v", perms[0].Users)
	}
	if len(perms[0].Excluded) != 1 || perms[0].Excluded[0] != "banned" {
		t.Errorf("expected excluded user 'banned', got %v", perms[0].Excluded)
	}
}

func TestParseFolderPermissions_InvalidMode(t *testing.T) {
	_, err := parseFolderPermissions("/files:alice:readwrite")
	if err == nil {
		t.Error("expected error for invalid mode")
	}
}

func TestParseFolderPermissions_MissingFields(t *testing.T) {
	_, err := parseFolderPermissions("/files:alice")
	if err == nil {
		t.Error("expected error for missing mode field")
	}
}

func TestParseFolderPermissions_EmptyUsers(t *testing.T) {
	_, err := parseFolderPermissions("/files::rw")
	if err == nil {
		t.Error("expected error for empty users")
	}
}

func TestParseFolderPermissions_SkipsBlankEntries(t *testing.T) {
	perms, err := parseFolderPermissions("/public:public:ro,  ,/files:alice:rw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(perms) != 2 {
		t.Errorf("expected 2 perms (blank entry skipped), got %d", len(perms))
	}
}

func TestParseFolderPermissions_ModeNormalized(t *testing.T) {
	perms, err := parseFolderPermissions("/files:alice:RW")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if perms[0].Mode != "rw" {
		t.Errorf("expected mode to be lowercased, got %q", perms[0].Mode)
	}
}
