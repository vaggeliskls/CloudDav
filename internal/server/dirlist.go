package server

import (
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"path"
	"sort"
	"strings"

	"golang.org/x/net/webdav"
)

// dirListHandler wraps a webdav.Handler and intercepts GET/HEAD requests on
// directories, serving an HTML directory listing instead of returning 405.
type dirListHandler struct {
	webdav http.Handler
	fs     webdav.FileSystem
}

func (h *dirListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		h.webdav.ServeHTTP(w, r)
		return
	}

	// Ensure directory paths end with "/".
	urlPath := r.URL.Path
	if !strings.HasSuffix(urlPath, "/") {
		fi, err := h.fs.Stat(r.Context(), urlPath)
		if err == nil && fi.IsDir() {
			http.Redirect(w, r, urlPath+"/", http.StatusMovedPermanently)
			return
		}
	}

	fi, err := h.fs.Stat(r.Context(), urlPath)
	if err != nil || !fi.IsDir() {
		// Not a directory — let the webdav handler deal with it (file or 404).
		h.webdav.ServeHTTP(w, r)
		return
	}

	// Open the directory and list its children.
	f, err := h.fs.OpenFile(r.Context(), urlPath, 0, 0)
	if err != nil {
		slog.Error("openfile failed", "path", urlPath, "err", err)
		http.Error(w, "cannot open directory", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	entries, err := f.Readdir(-1)
	if err != nil {
		slog.Error("readdir failed", "path", urlPath, "err", err)
		http.Error(w, "cannot read directory", http.StatusInternalServerError)
		return
	}

	// Sort: directories first, then files, both alphabetically.
	sort.Slice(entries, func(i, j int) bool {
		di, dj := entries[i].IsDir(), entries[j].IsDir()
		if di != dj {
			return di
		}
		return entries[i].Name() < entries[j].Name()
	})

	if r.Method == http.MethodHead {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	title := html.EscapeString(urlPath)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Index of %s</title>
  <style>
    body { font-family: monospace; margin: 2rem; }
    h1   { border-bottom: 1px solid #ccc; padding-bottom: .4rem; }
    table { border-collapse: collapse; width: 100%%; }
    th, td { text-align: left; padding: .3rem .8rem; }
    th { border-bottom: 2px solid #ccc; }
    tr:hover { background: #f5f5f5; }
    a    { text-decoration: none; color: #0366d6; }
    a:hover { text-decoration: underline; }
    .size { text-align: right; }
    .date { color: #666; }
  </style>
</head>
<body>
<h1>Index of %s</h1>
<table>
  <tr><th>Name</th><th class="size">Size</th><th class="date">Modified</th></tr>
`, title, title)

	// Parent directory link (except at root).
	if urlPath != "/" {
		parent := path.Dir(strings.TrimSuffix(urlPath, "/"))
		if !strings.HasSuffix(parent, "/") {
			parent += "/"
		}
		fmt.Fprintf(w, `  <tr><td><a href="%s">../</a></td><td class="size">-</td><td class="date">-</td></tr>
`, html.EscapeString(parent))
	}

	for _, e := range entries {
		name := html.EscapeString(e.Name())
		href := html.EscapeString(e.Name())
		sizeStr := "-"
		if !e.IsDir() {
			sizeStr = formatSize(e.Size())
		} else {
			href += "/"
			name += "/"
		}
		modTime := e.ModTime().Format("2006-01-02 15:04")
		fmt.Fprintf(w, "  <tr><td><a href=\"%s\">%s</a></td><td class=\"size\">%s</td><td class=\"date\">%s</td></tr>\n",
			href, name, sizeStr, modTime)
	}

	fmt.Fprint(w, "</table>\n</body>\n</html>\n")
}

func formatSize(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
