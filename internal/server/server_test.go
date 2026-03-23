package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloud-webdav-server/internal/config"
)

// newTestServer builds a full server with local storage in a temp directory.
func newTestServer(t *testing.T, folderPerms string, opts ...func(*config.Config)) *httptest.Server {
	t.Helper()
	dir := t.TempDir()

	cfg := &config.Config{
		StorageType:   config.StorageLocal,
		LocalDataPath: dir,
		ROMethods:     []string{"GET", "HEAD", "OPTIONS", "PROPFIND"},
		RWMethods:     []string{"GET", "HEAD", "OPTIONS", "PROPFIND", "PUT", "DELETE", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK", "PROPPATCH"},
		BasicUsers:    map[string]string{"alice": "alice123"},
		BasicAuthEnabled: true,
		AutoCreateFolders: true,
	}

	perms, err := config.ParseFolderPermissions(folderPerms)
	if err != nil {
		t.Fatalf("parse permissions: %v", err)
	}
	cfg.FolderPermissions = perms

	for _, o := range opts {
		o(cfg)
	}

	srv, err := New(cfg)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	return httptest.NewServer(srv.httpSrv.Handler)
}

func TestIntegration_PutAndGet(t *testing.T) {
	ts := newTestServer(t, "/files:alice:rw")
	defer ts.Close()

	// PUT a file.
	body := "hello webdav"
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/files/hello.txt", strings.NewReader(body))
	req.SetBasicAuth("alice", "alice123")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		t.Fatalf("PUT expected 201/204, got %d", resp.StatusCode)
	}

	// GET the file back.
	req, _ = http.NewRequest(http.MethodGet, ts.URL+"/files/hello.txt", nil)
	req.SetBasicAuth("alice", "alice123")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET expected 200, got %d", resp.StatusCode)
	}
	got, _ := io.ReadAll(resp.Body)
	if string(got) != body {
		t.Errorf("expected body %q, got %q", body, string(got))
	}
}

func TestIntegration_UnauthorizedReturns401(t *testing.T) {
	ts := newTestServer(t, "/files:alice:rw")
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/files/", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestIntegration_ForbiddenForWrongUser(t *testing.T) {
	// bob is a valid user but not listed in /files permissions → 403.
	ts := newTestServer(t, "/files:alice:rw", func(cfg *config.Config) {
		cfg.BasicUsers["bob"] = "bob123"
	})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/files/", nil)
	req.SetBasicAuth("bob", "bob123")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

func TestIntegration_PublicFolderNoAuth(t *testing.T) {
	ts := newTestServer(t, "/public:public:ro")
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/public/")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for public folder, got %d", resp.StatusCode)
	}
}

func TestIntegration_ReadOnlyFolderDeniesWrite(t *testing.T) {
	ts := newTestServer(t, "/files:alice:ro")
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/files/test.txt", strings.NewReader("data"))
	req.SetBasicAuth("alice", "alice123")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for PUT on ro folder, got %d", resp.StatusCode)
	}
}

func TestIntegration_Mkcol(t *testing.T) {
	ts := newTestServer(t, "/files:alice:rw")
	defer ts.Close()

	req, _ := http.NewRequest("MKCOL", ts.URL+"/files/newdir", nil)
	req.SetBasicAuth("alice", "alice123")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("MKCOL failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestIntegration_HealthCheck(t *testing.T) {
	ts := newTestServer(t, "/files:alice:rw")
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/_health")
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "OK" {
		t.Errorf("expected body OK, got %q", string(body))
	}
}
