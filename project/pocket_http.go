package project

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func registerPocketRoutes(r chi.Router, s *VersionStore) {
	r.Post("/book", func(w http.ResponseWriter, r *http.Request) {
		var input PocketBookAction
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&input); err != nil {
			writeJSON(w, 400, errorResponse{Error: "invalid book action"})
			return
		}
		saved, err := s.ChangePocketBook(r.Context(), authorID(r), chi.URLParam(r, "projectID"), input)
		if err != nil {
			writeVersionError(w, err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, saved)
	})
	r.Get("/versions/{versionID}/review/{entryID}", func(w http.ResponseWriter, r *http.Request) {
		review, err := s.ReadPocketReview(r.Context(), authorID(r), chi.URLParam(r, "projectID"), chi.URLParam(r, "versionID"), chi.URLParam(r, "entryID"))
		if err != nil {
			writeVersionError(w, err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, review)
	})
	r.Post("/review", func(w http.ResponseWriter, r *http.Request) {
		var input PocketAction
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&input); err != nil {
			writeJSON(w, 400, errorResponse{Error: "invalid review action"})
			return
		}
		saved, err := s.ChangePocketReview(r.Context(), authorID(r), chi.URLParam(r, "projectID"), input)
		if err != nil {
			writeVersionError(w, err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, saved)
	})
}
