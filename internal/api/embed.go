package api

import (
	"bytes"
	"compress/gzip"
	"embed"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
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
		if err == nil {
			// If the path is a directory, try serving its index.html
			if stat, sErr := f.Stat(); sErr == nil && stat.IsDir() {
				f.Close()
				idxPath := path + "/index.html"
				f2, err2 := sub.Open(idxPath)
				if err2 == nil {
					f = f2
					path = idxPath
				} else {
					// No index.html in dir — SPA fallback
					f, _ = sub.Open("index.html")
					path = "index.html"
				}
			}
		} else {
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

		if body, ok := gzipped(sub, path, r); ok {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")
			w.Write(body)
			return
		}
		w.Header().Set("Vary", "Accept-Encoding")
		io.Copy(w, f.(io.Reader))
	})
}

// The embedded assets were served raw, so the API reference page cost 3.9 MB on
// a cold visit where its largest chunk alone is 2.3 MB uncompressed and 670 KB
// gzipped. Nothing in front compresses either: the nginx template ships no
// `gzip` directive, and an instance run without a proxy has nothing at all.
//
// The files never change for the life of the process — they are compiled into
// it — so each is compressed once, on the first request that can use it, and
// kept.
var (
	gzipOnce  sync.Mutex
	gzipCache = map[string][]byte{}
)

// compressible lists what is worth the CPU. Images, fonts and archives are
// already compressed, and gzipping them costs time to make them slightly
// larger.
func compressible(ext string) bool {
	switch ext {
	case ".js", ".css", ".html", ".json", ".svg", ".txt", ".map", ".xml", ".webmanifest":
		return true
	}
	return false
}

func gzipped(sub fs.FS, path string, r *http.Request) ([]byte, bool) {
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		return nil, false
	}
	if !compressible(filepath.Ext(path)) {
		return nil, false
	}

	gzipOnce.Lock()
	defer gzipOnce.Unlock()
	if body, ok := gzipCache[path]; ok {
		return body, body != nil
	}

	raw, err := fs.ReadFile(sub, path)
	if err != nil {
		gzipCache[path] = nil
		return nil, false
	}
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		gzipCache[path] = nil
		return nil, false
	}
	if _, err := zw.Write(raw); err != nil {
		zw.Close()
		gzipCache[path] = nil
		return nil, false
	}
	if err := zw.Close(); err != nil {
		gzipCache[path] = nil
		return nil, false
	}
	// A file that does not shrink is served as it is; the header would only
	// cost the client a decompression pass for nothing.
	if buf.Len() >= len(raw) {
		gzipCache[path] = nil
		return nil, false
	}
	body := buf.Bytes()
	gzipCache[path] = body
	return body, true
}

func hasFileExtension(path string) bool {
	return strings.Contains(filepath.Base(path), ".")
}
