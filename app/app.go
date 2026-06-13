package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Dependencies holds services used by the application router.
type Dependencies struct{}

// New builds the HTTP handler for the Writer service.
func New(deps *Dependencies) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return r
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
