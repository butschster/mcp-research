package api

import (
	"net/http"

	"github.com/dovod-app/app/internal/docs"
)

func handleLLMSDocs() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}

		switch path {
		case "llms.txt":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		default:
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		}

		// Map URL path to file name
		fileName := path
		if len(fileName) > len("llms/") && fileName[:5] == "llms/" {
			fileName = fileName[5:]
		}

		data, err := docs.ReadFile(fileName)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write(data)
	})
}
