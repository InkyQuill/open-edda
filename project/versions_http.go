package project

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterVersionRoutes exposes project-scoped manifests and immutable bytes.
func RegisterVersionRoutes(r chi.Router, s *VersionStore) {
	if s == nil {
		return
	}
	r.Route("/projects/{projectID}/files", func(r chi.Router) {
		// Keep legacy projects from acquiring an unrelated second source of truth.
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var mode string
				err := s.db.QueryRowContext(r.Context(), "SELECT storage_mode FROM story_projects WHERE id=? AND author_id=?", chi.URLParam(r, "projectID"), authorID(r)).Scan(&mode)
				if err != nil {
					writeVersionError(w, err)
					return
				}
				if mode != "files" {
					writeJSON(w, 409, errorResponse{Error: "this project does not use file storage"})
					return
				}
				next.ServeHTTP(w, r)
			})
		})
		r.Get("/versions", func(w http.ResponseWriter, r *http.Request) {
			page, err := s.History(r.Context(), authorID(r), chi.URLParam(r, "projectID"), r.URL.Query().Get("cursor"))
			if err != nil {
				writeVersionError(w, err)
				return
			}
			w.Header().Set("Cache-Control", "no-store")
			writeJSON(w, 200, page)
		})
		r.Post("/restore", func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				ExpectedVersion string `json:"expectedVersion"`
				OperationID     string `json:"operationId"`
				VersionID       string `json:"versionId"`
			}
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&body); err != nil {
				writeJSON(w, 400, errorResponse{Error: "invalid restore request"})
				return
			}
			v, err := s.Restore(r.Context(), PublishVersionInput{AuthorID: authorID(r), ProjectID: chi.URLParam(r, "projectID"), ExpectedVersion: body.ExpectedVersion, OperationID: body.OperationID, Message: "Restore version " + body.VersionID}, body.VersionID)
			if err != nil {
				writeVersionError(w, err)
				return
			}
			writeJSON(w, 201, v)
		})
		r.Get("/operations/{operationID}", func(w http.ResponseWriter, r *http.Request) {
			var id string
			err := s.db.QueryRowContext(r.Context(), "SELECT id FROM project_versions WHERE project_id=? AND operation_id=?", chi.URLParam(r, "projectID"), chi.URLParam(r, "operationID")).Scan(&id)
			if err != nil {
				writeVersionError(w, err)
				return
			}
			v, err := s.Version(r.Context(), authorID(r), chi.URLParam(r, "projectID"), id)
			if err != nil {
				writeVersionError(w, err)
				return
			}
			w.Header().Set("Cache-Control", "no-store")
			writeJSON(w, 200, v)
		})
		r.Get("/versions/{versionID}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "versionID")
			if id == "current" {
				id = ""
			}
			v, err := s.Version(r.Context(), authorID(r), chi.URLParam(r, "projectID"), id)
			if err != nil {
				writeVersionError(w, err)
				return
			}
			w.Header().Set("Cache-Control", "no-store")
			writeJSON(w, 200, v)
		})
		r.Post("/versions", func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				ExpectedVersion string      `json:"expectedVersion"`
				OperationID     string      `json:"operationId"`
				Message         string      `json:"message"`
				Entries         []TreeEntry `json:"entries"`
			}
			r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&body); err != nil {
				writeMalformedJSON(w, err)
				return
			}
			var extra any
			if err := decoder.Decode(&extra); err != io.EOF {
				if err == nil {
					err = errors.New("expected one JSON document")
				}
				writeMalformedJSON(w, err)
				return
			}
			v, err := s.Publish(r.Context(), PublishVersionInput{AuthorID: authorID(r), ProjectID: chi.URLParam(r, "projectID"), ExpectedVersion: body.ExpectedVersion, OperationID: body.OperationID, Message: body.Message, Entries: body.Entries})
			if err != nil {
				writeVersionError(w, err)
				return
			}
			writeJSON(w, 201, v)
		})
		r.Put("/objects/{hash}", func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength < 0 {
				writeJSON(w, 411, errorResponse{Error: "Content-Length is required"})
				return
			}
			if r.ContentLength > s.limits.MaxObjectBytes {
				writeJSON(w, 413, errorResponse{Error: "file exceeds upload limit"})
				return
			}
			err := s.UploadObject(r.Context(), authorID(r), chi.URLParam(r, "projectID"), chi.URLParam(r, "hash"), r.ContentLength, http.MaxBytesReader(w, r.Body, s.limits.MaxObjectBytes))
			if err != nil {
				writeVersionError(w, err)
				return
			}
			w.WriteHeader(204)
		})
		r.Get("/versions/{versionID}/entries/{entryID}", func(w http.ResponseWriter, r *http.Request) {
			file, err := s.OpenVersionFile(r.Context(), authorID(r), chi.URLParam(r, "projectID"), chi.URLParam(r, "versionID"), chi.URLParam(r, "entryID"))
			if err != nil {
				writeVersionError(w, err)
				return
			}
			defer file.Close()
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", "attachment")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cache-Control", "private, no-store")
			_, _ = io.Copy(w, file)
		})
	})
}

func writeVersionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		writeJSON(w, 404, errorResponse{Error: "project, version or file not found"})
	case errors.Is(err, ErrVersionConflict), errors.Is(err, ErrOperationConflict):
		writeJSON(w, 409, errorResponse{Error: err.Error()})
	case errors.Is(err, ErrInvalidTree):
		writeJSON(w, 400, errorResponse{Error: err.Error()})
	case errors.Is(err, ErrObjectIntegrity):
		writeJSON(w, 422, errorResponse{Error: "file bytes are missing or failed verification; upload again"})
	default:
		writeJSON(w, 500, errorResponse{Error: "file storage is unavailable"})
	}
}
