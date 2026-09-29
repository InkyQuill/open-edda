package auth

import (
	"database/sql"
	"errors"
	"net/http"

	"git.inkyquill.net/inky/writer/internal/httputil"
	"github.com/go-chi/chi/v5"
)

// Theme IDs mirror @inkyquill/galley-themes 0.16.0. Update with the package catalog.
var profileThemeIDs = map[string]bool{
	"galley-light": true, "galley-dark": true, "thoth-light": true, "thoth-dark": true,
	"gruvbox-light": true, "gruvbox-dark": true, "catppuccin-latte": true, "catppuccin-mocha": true,
	"tokyo-night-day": true, "tokyo-night": true, "nord-light": true, "nord-dark": true,
	"darcula": true, "solarized-light": true, "solarized-dark": true,
}

type AuthorPreferences struct {
	ThemeID string `json:"themeId"`
}

// RegisterProfileRoutes protects profile preferences independently of project access.
func RegisterProfileRoutes(r chi.Router, service *Service) {
	if service == nil {
		return
	}
	r.Group(func(r chi.Router) {
		r.Use(Required(service))
		r.Get("/auth/preferences", service.getPreferences)
		r.Put("/auth/preferences", service.putPreferences)
	})
}

func (s *Service) getPreferences(w http.ResponseWriter, r *http.Request) {
	preferences := AuthorPreferences{ThemeID: "thoth-light"}
	err := s.db.QueryRowContext(r.Context(), "SELECT theme_id FROM author_preferences WHERE author_id = ?", MustAuthorIDFromContext(r.Context())).Scan(&preferences.ThemeID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not load preferences"})
		return
	}
	writeJSON(w, http.StatusOK, preferences)
}

func (s *Service) putPreferences(w http.ResponseWriter, r *http.Request) {
	var preferences AuthorPreferences
	if err := httputil.DecodeJSON(w, r, &preferences, httputil.DefaultJSONBodyLimit); err != nil {
		status := http.StatusBadRequest
		if httputil.IsRequestTooLarge(err) {
			status = http.StatusRequestEntityTooLarge
		}
		writeJSON(w, status, map[string]string{"error": "invalid preferences"})
		return
	}
	if !profileThemeIDs[preferences.ThemeID] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown theme"})
		return
	}
	_, err := s.db.ExecContext(r.Context(), "INSERT INTO author_preferences (author_id, theme_id) VALUES (?, ?) ON CONFLICT(author_id) DO UPDATE SET theme_id = excluded.theme_id", MustAuthorIDFromContext(r.Context()), preferences.ThemeID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not save preferences"})
		return
	}
	writeJSON(w, http.StatusOK, preferences)
}
