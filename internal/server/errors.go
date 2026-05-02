package server

import (
	_ "embed"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
)

//go:embed templates/error.html
var errorTemplateSrc string

var errorTemplate = template.Must(template.New("error").Parse(errorTemplateSrc))

type errorData struct {
	Status  int
	Title   string
	Message string
}

// renderError writes a status response. Browser clients (Accept: text/html)
// receive a styled HTML page; everyone else gets a plain-text body so WebDAV
// clients keep working as before.
func renderError(w http.ResponseWriter, r *http.Request, status int, message string) {
	if !acceptsHTML(r) {
		http.Error(w, message, status)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	data := errorData{
		Status:  status,
		Title:   http.StatusText(status),
		Message: message,
	}
	if err := errorTemplate.Execute(w, data); err != nil {
		slog.Error("error template execute failed", "err", err)
	}
}

func acceptsHTML(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}
