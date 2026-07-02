package app

import (
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"git.inkyquill.net/inky/writer/agent"
	"git.inkyquill.net/inky/writer/auth"
	"git.inkyquill.net/inky/writer/internal/httputil"
	"git.inkyquill.net/inky/writer/project"
	"git.inkyquill.net/inky/writer/skill"
	"github.com/go-chi/chi/v5"
)

// Dependencies holds services used by the application router.
type Dependencies struct {
	AuthService    *auth.Service
	ProjectService *project.Service
	AgentService   *agent.Service
	SkillService   *skill.Service
	StaticFS       fs.FS
}

// New builds the HTTP handler for the Open Edda service.
func New(deps *Dependencies) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	if deps != nil {
		r.Route("/api", func(r chi.Router) {
			// Public routes — no auth required.
			if deps.AuthService != nil {
				auth.RegisterRoutes(r, deps.AuthService)
			}

			// Protected routes — require valid Bearer token.
			r.Group(func(r chi.Router) {
				if deps.AuthService != nil {
					r.Use(auth.Required(deps.AuthService))
					if deps.ProjectService != nil {
						r.Use(requireProjectAccess(deps.ProjectService))
					}
				}
				project.RegisterRoutes(r, deps.ProjectService)
				agent.RegisterRoutes(r, deps.AgentService)
				skill.RegisterRoutes(r, deps.SkillService)
			})
		})

		if deps.StaticFS != nil {
			r.NotFound(spaHandler(deps.StaticFS))
		}
	}
	return r
}

func requireProjectAccess(projectService *project.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			projectID, ok := projectIDFromPath(r.URL.Path)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			authorID, ok := auth.AuthorIDFromContext(r.Context())
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing author context"})
				return
			}
			owns, err := projectService.AuthorOwnsProject(r.Context(), authorID, projectID)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "project access check failed"})
				return
			}
			if !owns {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func projectIDFromPath(path string) (string, bool) {
	const prefix = "/api/projects/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(path, prefix)
	if rest == "" {
		return "", false
	}
	projectID, _, _ := strings.Cut(rest, "/")
	if projectID == "" || projectID == "import" {
		return "", false
	}
	return projectID, true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	httputil.WriteJSON(w, status, value)
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
