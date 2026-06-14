package skill

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"git.inkyquill.net/inky/writer/project"
	"github.com/go-chi/chi/v5"
)

type httpHandler struct {
	service *Service
}

type selectSessionSkillsRequest struct {
	SkillIDs []string `json:"skillIds"`
}

// RegisterRoutes mounts skill core routes on an /api router.
func RegisterRoutes(r chi.Router, service *Service) {
	if service == nil {
		return
	}
	h := httpHandler{service: service}
	r.Get("/projects/{projectID}/skills", h.listSkills)
	r.Post("/projects/{projectID}/skills/import", h.importSkill)
	r.Post("/projects/{projectID}/skills/import-local", h.importLocalSkill)
	r.Get("/projects/{projectID}/skills/{skillID}", h.getSkill)
	r.Get("/projects/{projectID}/agent/sessions/{sessionID}/skills", h.listSessionSkills)
	r.Put("/projects/{projectID}/agent/sessions/{sessionID}/skills", h.selectSessionSkills)
}

func (h httpHandler) listSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := h.service.List(r.Context(), chi.URLParam(r, "projectID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skills)
}

func (h httpHandler) importSkill(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxSkillArchiveBytes))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid skill archive"})
		return
	}

	imported, err := ParseSkillArchive(bytes.NewReader(body), int64(len(body)), "uploaded skill archive")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid skill archive"})
		return
	}

	skill, err := h.service.Install(r.Context(), InstallInput{
		ProjectID:   chi.URLParam(r, "projectID"),
		SourceType:  SourceTypeUpload,
		SourceLabel: imported.SourceLabel,
		Imported:    imported,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, skill)
}

func (h httpHandler) importLocalSkill(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	var req struct {
		Directory string `json:"directory"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
		return
	}
	if req.Directory == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "directory is required"})
		return
	}
	directory, err := resolveLocalSkillDirectory(req.Directory)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "directory not accessible"})
		return
	}
	imported, err := ParseSkillDirectory(directory)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid skill directory"})
		return
	}
	skill, err := h.service.Install(r.Context(), InstallInput{
		ProjectID:   projectID,
		SourceType:  SourceTypeLocalDirectory,
		SourceLabel: "local directory import",
		Imported:    imported,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, skill)
}

func resolveLocalSkillDirectory(requested string) (string, error) {
	base := os.Getenv("WRITER_SKILL_IMPORT_ROOT")
	if base == "" {
		return "", fmt.Errorf("local skill imports are disabled")
	}
	baseAbs, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(requested) {
		requested = filepath.Join(baseAbs, requested)
	}
	requestedAbs, err := filepath.Abs(requested)
	if err != nil {
		return "", err
	}
	baseResolved, err := filepath.EvalSymlinks(baseAbs)
	if err != nil {
		return "", err
	}
	requestedResolved, err := filepath.EvalSymlinks(requestedAbs)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(baseResolved, requestedResolved)
	if err != nil {
		return "", err
	}
	if rel == "." || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".." && !filepath.IsAbs(rel)) {
		info, err := os.Stat(requestedResolved)
		if err != nil {
			return "", err
		}
		if !info.IsDir() {
			return "", fmt.Errorf("path is not a directory")
		}
		return requestedResolved, nil
	}
	return "", fmt.Errorf("directory outside allowed root")
}

func (h httpHandler) getSkill(w http.ResponseWriter, r *http.Request) {
	skill, err := h.service.Get(r.Context(), chi.URLParam(r, "projectID"), chi.URLParam(r, "skillID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skill)
}

func (h httpHandler) listSessionSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := h.service.ListSessionSkills(r.Context(), chi.URLParam(r, "projectID"), chi.URLParam(r, "sessionID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skills)
}

func (h httpHandler) selectSessionSkills(w http.ResponseWriter, r *http.Request) {
	var input selectSessionSkillsRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
		return
	}

	skills, err := h.service.SelectSessionSkills(r.Context(), SelectSessionSkillsInput{
		ProjectID: chi.URLParam(r, "projectID"),
		SessionID: chi.URLParam(r, "sessionID"),
		SkillIDs:  input.SkillIDs,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skills)
}

func decodeJSON(r *http.Request, value any) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("trailing JSON data")
	}
	return nil
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid input"})
	case errors.Is(err, sql.ErrNoRows):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
	case errors.Is(err, project.ErrConflict):
		writeJSON(w, http.StatusConflict, errorResponse{Error: "conflict"})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
