package project

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const placeholderAuthorID = "author-1"

type httpHandler struct {
	service *Service
}

type createProjectRequest struct {
	Title    string `json:"title"`
	Language string `json:"language"`
}

type createContentRequest struct {
	Kind         ContentKind `json:"kind"`
	Title        string      `json:"title"`
	BodyMarkdown string      `json:"bodyMarkdown"`
	MetadataJSON string      `json:"metadataJson"`
	SortOrder    int64       `json:"sortOrder"`
	Reason       string      `json:"reason"`
}

type updateContentRequest struct {
	ExpectedRevision int64  `json:"expectedRevision"`
	BodyMarkdown     string `json:"bodyMarkdown"`
	MetadataJSON     string `json:"metadataJson"`
	Reason           string `json:"reason"`
}

// RegisterRoutes mounts project core routes on an /api router.
func RegisterRoutes(r chi.Router, service *Service) {
	if service == nil {
		return
	}

	h := httpHandler{service: service}
	r.Get("/projects", h.listProjects)
	r.Post("/projects", h.createProject)
	r.Get("/projects/{projectID}/content", h.listContent)
	r.Post("/projects/{projectID}/content", h.createContent)
	r.Get("/projects/{projectID}/content/{contentID}", h.getContent)
	r.Put("/projects/{projectID}/content/{contentID}", h.updateContent)
	r.Get("/projects/{projectID}/content/{contentID}/revisions", h.listRevisions)
}

func (h httpHandler) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.ListProjects(r.Context(), placeholderAuthorID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, projects)
}

func (h httpHandler) createProject(w http.ResponseWriter, r *http.Request) {
	var input createProjectRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
		return
	}

	project, err := h.service.CreateProject(r.Context(), CreateProjectInput{
		AuthorID: placeholderAuthorID,
		Title:    input.Title,
		Language: input.Language,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, project)
}

func (h httpHandler) listContent(w http.ResponseWriter, r *http.Request) {
	kind := ContentKind(r.URL.Query().Get("kind"))
	if !validContentKind(kind) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid content kind"})
		return
	}

	items, err := h.service.ListContent(r.Context(), chi.URLParam(r, "projectID"), kind)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (h httpHandler) createContent(w http.ResponseWriter, r *http.Request) {
	var input createContentRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
		return
	}
	if !validContentKind(input.Kind) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid content kind"})
		return
	}

	item, err := h.service.CreateContent(r.Context(), CreateContentInput{
		ProjectID:    chi.URLParam(r, "projectID"),
		Kind:         input.Kind,
		Title:        input.Title,
		BodyMarkdown: input.BodyMarkdown,
		MetadataJSON: input.MetadataJSON,
		SortOrder:    input.SortOrder,
		Reason:       input.Reason,
		CreatedBy:    "author",
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h httpHandler) getContent(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetContent(r.Context(), chi.URLParam(r, "projectID"), chi.URLParam(r, "contentID"))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h httpHandler) updateContent(w http.ResponseWriter, r *http.Request) {
	var input updateContentRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
		return
	}

	item, err := h.service.UpdateContent(r.Context(), UpdateContentInput{
		ProjectID:        chi.URLParam(r, "projectID"),
		ContentID:        chi.URLParam(r, "contentID"),
		ExpectedRevision: input.ExpectedRevision,
		BodyMarkdown:     input.BodyMarkdown,
		MetadataJSON:     input.MetadataJSON,
		Reason:           input.Reason,
		CreatedBy:        "author",
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h httpHandler) listRevisions(w http.ResponseWriter, r *http.Request) {
	revisions, err := h.service.ListRevisions(r.Context(), chi.URLParam(r, "projectID"), chi.URLParam(r, "contentID"))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, revisions)
}

func decodeJSON(r *http.Request, value any) error {
	return json.NewDecoder(r.Body).Decode(value)
}

func validContentKind(kind ContentKind) bool {
	switch kind {
	case KindChapter, KindStoryBibleEntry, KindWritingBrief, KindProjectNote:
		return true
	default:
		return false
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrConflict):
		writeJSON(w, http.StatusConflict, errorResponse{Error: "conflict"})
	case errors.Is(err, sql.ErrNoRows):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
