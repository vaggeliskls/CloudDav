package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---- health middleware -------------------------------------------------------

func TestHealthMiddleware_ReturnsOK(t *testing.T) {
	h := healthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called for /_health")
	}))

	req := httptest.NewRequest(http.MethodGet, "/_health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "OK" {
		t.Errorf("expected body 'OK', got %q", body)
	}
}

func TestHealthMiddleware_PassesOtherPaths(t *testing.T) {
	called := false
	h := healthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/files/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !called {
		t.Error("inner handler should be called for non-health path")
	}
}

// ---- download middleware -----------------------------------------------------

func TestDownloadMiddleware_SetsAttachmentHeader(t *testing.T) {
	h := downloadMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No Content-Type set → should trigger attachment header.
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/files/report.pdf", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	cd := rec.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "attachment") {
		t.Errorf("expected Content-Disposition: attachment, got %q", cd)
	}
	if !strings.Contains(cd, "report.pdf") {
		t.Errorf("expected filename in Content-Disposition, got %q", cd)
	}
}

func TestDownloadMiddleware_SkipsHTMLResponse(t *testing.T) {
	h := downloadMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/files/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if cd := rec.Header().Get("Content-Disposition"); cd != "" {
		t.Errorf("expected no Content-Disposition for HTML, got %q", cd)
	}
}

func TestDownloadMiddleware_SkipsNonGET(t *testing.T) {
	h := downloadMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, method := range []string{http.MethodPut, http.MethodDelete, http.MethodOptions} {
		req := httptest.NewRequest(method, "/files/report.pdf", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if cd := rec.Header().Get("Content-Disposition"); cd != "" {
			t.Errorf("%s: expected no Content-Disposition, got %q", method, cd)
		}
	}
}

// ---- security headers middleware --------------------------------------------

func TestSecurityHeadersMiddleware(t *testing.T) {
	h := securityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if v := rec.Header().Get("X-Content-Type-Options"); v != "nosniff" {
		t.Errorf("expected X-Content-Type-Options: nosniff, got %q", v)
	}
	if v := rec.Header().Get("X-Frame-Options"); v != "DENY" {
		t.Errorf("expected X-Frame-Options: DENY, got %q", v)
	}
}

// ---- browser block middleware -----------------------------------------------

func TestBrowserBlockMiddleware_BlocksMozilla(t *testing.T) {
	h := browserBlockMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called for browser requests")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestBrowserBlockMiddleware_AllowsNonBrowser(t *testing.T) {
	called := false
	h := browserBlockMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "WebDAVClient/1.0")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !called {
		t.Error("inner handler should be called for non-browser client")
	}
}

// ---- CORS middleware --------------------------------------------------------

func TestCORSMiddleware_SetsHeaders(t *testing.T) {
	h := corsMiddleware("*", "GET,PUT", "Authorization")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "*" {
		t.Errorf("expected CORS origin *, got %q", v)
	}
}

func TestCORSMiddleware_PreflightReturns204(t *testing.T) {
	h := corsMiddleware("*", "GET,PUT", "Authorization")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("inner handler should not be called for OPTIONS preflight")
		}),
	)

	req := httptest.NewRequest(http.MethodOptions, "/files/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}
