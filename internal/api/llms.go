package api

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed llmsdocs/*
var llmsDocsFS embed.FS

func handleLLMSDocs() http.Handler {
	sub, _ := fs.Sub(llmsDocsFS, "llmsdocs")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve plain text for .txt and .md files
		path := r.URL.Path
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}

		// Map URL paths to files
		switch path {
		case "llms.txt":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			data, err := fs.ReadFile(sub, "llms.txt")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Write(data)
		default:
			// Try to serve from llms/ subdirectory
			filePath := path
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			data, err := fs.ReadFile(sub, filePath)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Write(data)
		}
	})
}
