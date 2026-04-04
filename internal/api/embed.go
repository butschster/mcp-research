package api

import (
	"embed"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed all:static
var staticFS embed.FS

func init() {
	// Ensure common MIME types are registered
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".html", "text/html")
	mime.AddExtensionType(".json", "application/json")
	mime.AddExtensionType(".svg", "image/svg+xml")
	mime.AddExtensionType(".png", "image/png")
	mime.AddExtensionType(".ico", "image/x-icon")
	mime.AddExtensionType(".woff", "font/woff")
	mime.AddExtensionType(".woff2", "font/woff2")
	mime.AddExtensionType(".map", "application/json")
}

func staticHandler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("embedded static fs: " + err.Error())
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		f, err := sub.Open(path)
		if err != nil {
			// Asset file not found — return 404
			if hasFileExtension(r.URL.Path) {
				http.NotFound(w, r)
				return
			}
			// SPA fallback for client-side routes
			path = "index.html"
			f, err = sub.Open(path)
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}
		defer f.Close()

		// Set correct Content-Type from extension
		ext := filepath.Ext(path)
		if ct := mime.TypeByExtension(ext); ct != "" {
			w.Header().Set("Content-Type", ct)
		}

		// Set cache headers for hashed assets
		if strings.HasPrefix(r.URL.Path, "/_nuxt/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}

		io.Copy(w, f.(io.Reader))
	})
}

func hasFileExtension(path string) bool {
	return strings.Contains(filepath.Base(path), ".")
}
