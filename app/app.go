package app

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"git.inkyquill.net/inky/writer/agent"
	"git.inkyquill.net/inky/writer/project"
	"github.com/go-chi/chi/v5"
)

// Dependencies holds services used by the application router.
type Dependencies struct {
	ProjectService *project.Service
	AgentService   *agent.Service
	StaticFS       fs.FS
}

// New builds the HTTP handler for the Writer service.
func New(deps *Dependencies) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	if deps != nil {
		r.Route("/api", func(r chi.Router) {
			project.RegisterRoutes(r, deps.ProjectService)
			agent.RegisterRoutes(r, deps.AgentService)
		})
		if deps.StaticFS != nil {
			r.NotFound(spaHandler(deps.StaticFS))
		}
	}
	return r
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func spaHandler(staticFS fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(staticFS))
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(staticFS, path); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				http.Error(w, "static file error", http.StatusInternalServerError)
				return
			}
			http.ServeFileFS(w, r, staticFS, "index.html")
			return
		}

		fileServer.ServeHTTP(w, r)
	}
}
