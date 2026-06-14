package project

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"git.inkyquill.net/inky/writer/auth"
	"git.inkyquill.net/inky/writer/markdownio"
	"github.com/go-chi/chi/v5"
)

func authorID(r *http.Request) string {
	if id, ok := auth.AuthorIDFromContext(r.Context()); ok {
		return id
	}
	return "author-1"
}

const maxElysiumZipBytes = 10 << 20
const maxElysiumFiles = 512
const maxElysiumUncompressedBytes = 50 << 20

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
	r.Post("/projects/import/elysium", h.importElysium)
	r.Get("/projects/{projectID}/export/elysium", h.exportElysium)
	r.Get("/projects/{projectID}/content", h.listContent)
	r.Post("/projects/{projectID}/content", h.createContent)
	r.Get("/projects/{projectID}/content/{contentID}", h.getContent)
	r.Put("/projects/{projectID}/content/{contentID}", h.updateContent)
	r.Get("/projects/{projectID}/content/{contentID}/revisions", h.listRevisions)
}

func (h httpHandler) listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.ListProjects(r.Context(), authorID(r))
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
		AuthorID: authorID(r),
		Title:    input.Title,
		Language: input.Language,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, project)
}

func (h httpHandler) importElysium(w http.ResponseWriter, r *http.Request) {
	root, err := unzipRequestBody(w, r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid Elysium zip"})
		return
	}
	defer os.RemoveAll(root)

	items, err := markdownio.ImportElysiumLayout(root)
	if err != nil {
		writeError(w, err)
		return
	}

	project, err := h.service.ImportElysiumProject(r.Context(), authorID(r), "Imported Elysium", "en", items)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, project)
}

func (h httpHandler) exportElysium(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	content, err := h.service.ListProjectContent(r.Context(), projectID)
	if err != nil {
		writeError(w, err)
		return
	}

	items := make([]markdownio.ImportedItem, 0, len(content))
	for _, item := range content {
		sections, err := h.service.ListEntrySections(r.Context(), projectID, item.ID)
		if err != nil {
			writeError(w, err)
			return
		}
		relations, err := h.service.ListEntryRelations(r.Context(), projectID, item.ID)
		if err != nil {
			writeError(w, err)
			return
		}
		items = append(items, importedItemFromContent(item, sections, relations))
	}

	root, err := os.MkdirTemp("", "writer-elysium-export-*")
	if err != nil {
		writeError(w, err)
		return
	}
	defer os.RemoveAll(root)

	if err := markdownio.ExportElysiumLayout(root, items); err != nil {
		writeError(w, err)
		return
	}
	var body bytes.Buffer
	if err := zipDirectory(root, &body); err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="elysium.zip"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body.Bytes())
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

func importedItemFromContent(item ContentItem, sections []EntrySection, relations []EntryRelation) markdownio.ImportedItem {
	importedSections := make([]markdownio.ImportedSection, 0, len(sections))
	for _, section := range sections {
		importedSections = append(importedSections, markdownio.ImportedSection{
			Heading:      section.Heading,
			BodyMarkdown: section.BodyMarkdown,
			SortOrder:    int(section.SortOrder),
		})
	}
	importedRelations := make([]markdownio.ImportedRelation, 0, len(relations))
	for _, relation := range relations {
		importedRelations = append(importedRelations, markdownio.ImportedRelation{
			TargetTitle:  relation.TargetTitle,
			RelationType: relation.RelationType,
		})
	}
	return markdownio.ImportedItem{
		Kind:         string(item.Kind),
		Title:        item.Title,
		BodyMarkdown: item.BodyMarkdown,
		MetadataJSON: item.MetadataJSON,
		Sections:     importedSections,
		Relations:    importedRelations,
	}
}

func unzipRequestBody(w http.ResponseWriter, r *http.Request) (string, error) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxElysiumZipBytes))
	if err != nil {
		return "", err
	}
	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return "", err
	}
	root, err := os.MkdirTemp("", "writer-elysium-import-*")
	if err != nil {
		return "", err
	}
	if err := unzipToDirectory(reader, root); err != nil {
		_ = os.RemoveAll(root)
		return "", err
	}
	return root, nil
}

func unzipToDirectory(reader *zip.Reader, root string) error {
	var fileCount int
	var totalUncompressed uint64
	for _, file := range reader.File {
		if !filepath.IsLocal(file.Name) || strings.Contains(file.Name, `\`) {
			return errors.New("unsafe zip path")
		}
		fileCount++
		if fileCount > maxElysiumFiles {
			return errors.New("too many files in zip")
		}
		totalUncompressed += file.UncompressedSize64
		if totalUncompressed > maxElysiumUncompressedBytes {
			return errors.New("zip expands too large")
		}
		target := filepath.Join(root, filepath.FromSlash(file.Name))
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		source, err := file.Open()
		if err != nil {
			return err
		}
		destination, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			_ = source.Close()
			return err
		}
		limited := &io.LimitedReader{R: source, N: int64(file.UncompressedSize64) + 1}
		_, copyErr := io.Copy(destination, limited)
		closeErr := errors.Join(source.Close(), destination.Close())
		if copyErr != nil {
			return copyErr
		}
		if limited.N == 0 {
			return errors.New("zip entry exceeds declared size")
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func zipDirectory(root string, writer io.Writer) error {
	zipWriter := zip.NewWriter(writer)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		header, err := zipWriter.Create(filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		return writeZipFile(header, path)
	})
	return errors.Join(err, zipWriter.Close())
}

func writeZipFile(writer io.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(writer, file)
	closeErr := file.Close()
	return errors.Join(copyErr, closeErr)
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
