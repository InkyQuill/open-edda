package agent

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"git.inkyquill.net/inky/writer/auth"
	"git.inkyquill.net/inky/writer/project"
	"git.inkyquill.net/inky/writer/skill"
	"github.com/go-chi/chi/v5"
)

func authorID(r *http.Request) string {
	if id, ok := auth.AuthorIDFromContext(r.Context()); ok {
		return id
	}
	return "author-1"
}

const defaultListLimit = 50

type httpHandler struct {
	service *Service
}

type createProviderConfigRequest struct {
	Name    string `json:"name"`
	BaseURL string `json:"baseUrl"`
	APIKey  string `json:"apiKey"`
}

type updateProviderConfigRequest struct {
	BaseURL string `json:"baseUrl"`
	APIKey  string `json:"apiKey"`
}

type createModelVariantRequest struct {
	Name                      string          `json:"name"`
	Model                     string          `json:"model"`
	Temperature               float64         `json:"temperature"`
	MaxOutputTokens           int64           `json:"maxOutputTokens"`
	ContextWindowTokens       int64           `json:"contextWindowTokens"`
	InputPricePerMillion      float64         `json:"inputPricePerMillion"`
	OutputPricePerMillion     float64         `json:"outputPricePerMillion"`
	CacheReadPricePerMillion  float64         `json:"cacheReadPricePerMillion"`
	CacheWritePricePerMillion float64         `json:"cacheWritePricePerMillion"`
	RequestTokenField         string          `json:"requestTokenField"`
	ReasoningFormat           string          `json:"reasoningFormat"`
	CompatibilityJSON         json.RawMessage `json:"compatibilityJson"`
}

type upsertPromptProfileRequest struct {
	Genre                     string `json:"genre"`
	Tense                     string `json:"tense"`
	POV                       string `json:"pov"`
	Voice                     string `json:"voice"`
	InstructionsMarkdown      string `json:"instructionsMarkdown"`
	PromptRecordRetentionDays int64  `json:"promptRecordRetentionDays"`
}

type createSessionRequest struct {
	Title          string     `json:"title"`
	ActionKind     ActionKind `json:"actionKind"`
	ModelVariantID string     `json:"modelVariantId"`
	ApplyMode      ApplyMode  `json:"applyMode"`
	SkillIDs       []string   `json:"skillIds"`
}

type chatMessageRequest struct {
	BodyMarkdown string `json:"bodyMarkdown"`
}

type continuationRequest struct {
	ContentID         string    `json:"contentId"`
	ModelVariantID    string    `json:"modelVariantId"`
	ApplyMode         ApplyMode `json:"applyMode"`
	Guidance          string    `json:"guidance"`
	ExpectedRevision  int64     `json:"expectedRevision"`
	InsertPosition    int64     `json:"insertPosition"`
	Insert            bool      `json:"insert"`
	ContinuationUnits string    `json:"continuationUnits"`
	ContinuationCount int64     `json:"continuationCount"`
	SkillIDs          []string  `json:"skillIds"`
}

type rewriteRequest struct {
	ContentID        string    `json:"contentId"`
	ModelVariantID   string    `json:"modelVariantId"`
	ApplyMode        ApplyMode `json:"applyMode"`
	Guidance         string    `json:"guidance"`
	ExpectedRevision int64     `json:"expectedRevision"`
	SelectionStart   int64     `json:"selectionStart"`
	SelectionEnd     int64     `json:"selectionEnd"`
	SkillIDs         []string  `json:"skillIds"`
}

type prunePromptRecordsResponse struct {
	Deleted int64 `json:"deleted"`
}

// RegisterRoutes mounts agent core routes on an /api router.
func RegisterRoutes(r chi.Router, service *Service) {
	if service == nil {
		return
	}

	h := httpHandler{service: service}
	r.Get("/provider-configs", h.listProviderConfigs)
	r.Post("/provider-configs", h.createProviderConfig)
	r.Put("/provider-configs/{providerID}", h.updateProviderConfig)
	r.Get("/provider-configs/{providerID}/model-variants", h.listModelVariants)
	r.Post("/provider-configs/{providerID}/model-variants", h.createModelVariant)

	r.Get("/projects/{projectID}/agent/prompt-profile", h.getPromptProfile)
	r.Put("/projects/{projectID}/agent/prompt-profile", h.upsertPromptProfile)
	r.Get("/projects/{projectID}/agent/sessions", h.listSessions)
	r.Post("/projects/{projectID}/agent/sessions", h.createSession)
	r.Post("/projects/{projectID}/agent/sessions/{sessionID}/messages", h.createSessionMessage)
	r.Post("/projects/{projectID}/agent/actions/continuation", h.runContinuation)
	r.Post("/projects/{projectID}/agent/actions/rewrite", h.runRewrite)
	r.Post("/projects/{projectID}/agent/actions/read-check", h.runReadCheck)
	r.Post("/projects/{projectID}/agent/candidates/{candidateID}/accept", h.acceptCandidate)
	r.Post("/projects/{projectID}/agent/candidates/{candidateID}/reject", h.rejectCandidate)
	r.Get("/projects/{projectID}/agent/activity", h.listActivity)
	r.Get("/projects/{projectID}/agent/prompt-records", h.listPromptRecords)
	r.Post("/projects/{projectID}/agent/prompt-records/prune", h.prunePromptRecords)
}

func (h httpHandler) listProviderConfigs(w http.ResponseWriter, r *http.Request) {
	providers, err := h.service.ListProviderConfigs(r.Context(), authorID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, providers)
}

func (h httpHandler) createProviderConfig(w http.ResponseWriter, r *http.Request) {
	var input createProviderConfigRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
		return
	}
	if !validateProviderConfigRequest(w, input.Name, input.BaseURL, input.APIKey, true) {
		return
	}

	provider, err := h.service.CreateProviderConfig(r.Context(), CreateProviderConfigInput{
		AuthorID: authorID(r),
		Name:     input.Name,
		BaseURL:  input.BaseURL,
		APIKey:   input.APIKey,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, provider)
}

func (h httpHandler) updateProviderConfig(w http.ResponseWriter, r *http.Request) {
	var input updateProviderConfigRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
		return
	}
	if !validateProviderConfigRequest(w, "existing provider", input.BaseURL, input.APIKey, false) {
		return
	}

	provider, err := h.service.UpdateProviderConfig(r.Context(), UpdateProviderConfigInput{
		AuthorID:   authorID(r),
		ProviderID: chi.URLParam(r, "providerID"),
		BaseURL:    input.BaseURL,
		APIKey:     input.APIKey,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, provider)
}

func (h httpHandler) listModelVariants(w http.ResponseWriter, r *http.Request) {
	models, err := h.service.ListModelVariantsByProvider(r.Context(), authorID(r), chi.URLParam(r, "providerID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, models)
}

func (h httpHandler) createModelVariant(w http.ResponseWriter, r *http.Request) {
	var input createModelVariantRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
		return
	}
	if !validateModelVariantRequest(w, input.Name, input.Model) {
		return
	}

	model, err := h.service.CreateModelVariant(r.Context(), CreateModelVariantInput{
		AuthorID:                  authorID(r),
		ProviderConfigID:          chi.URLParam(r, "providerID"),
		Name:                      input.Name,
		Model:                     input.Model,
		Temperature:               input.Temperature,
		MaxOutputTokens:           input.MaxOutputTokens,
		ContextWindowTokens:       input.ContextWindowTokens,
		InputPricePerMillion:      input.InputPricePerMillion,
		OutputPricePerMillion:     input.OutputPricePerMillion,
		CacheReadPricePerMillion:  input.CacheReadPricePerMillion,
		CacheWritePricePerMillion: input.CacheWritePricePerMillion,
		RequestTokenField:         input.RequestTokenField,
		ReasoningFormat:           input.ReasoningFormat,
		CompatibilityJSON:         string(input.CompatibilityJSON),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, model)
}

func (h httpHandler) getPromptProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := h.service.GetPromptProfile(r.Context(), chi.URLParam(r, "projectID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (h httpHandler) upsertPromptProfile(w http.ResponseWriter, r *http.Request) {
	var input upsertPromptProfileRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
		return
	}

	profile, err := h.service.UpsertPromptProfile(r.Context(), UpsertPromptProfileInput{
		ProjectID:                 chi.URLParam(r, "projectID"),
		Genre:                     input.Genre,
		Tense:                     input.Tense,
		POV:                       input.POV,
		Voice:                     input.Voice,
		InstructionsMarkdown:      input.InstructionsMarkdown,
		PromptRecordRetentionDays: input.PromptRecordRetentionDays,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func (h httpHandler) listSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.service.ListSessions(r.Context(), ListSessionsInput{
		ProjectID: chi.URLParam(r, "projectID"),
		Limit:     limitFromQuery(r),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (h httpHandler) createSession(w http.ResponseWriter, r *http.Request) {
	var input createSessionRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
		return
	}
	if input.ActionKind == "" {
		input.ActionKind = ActionKindChat
	}
	if input.ApplyMode == "" {
		input.ApplyMode = ApplyModePreview
	}
	if err := validateActionKind(input.ActionKind); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid action kind"})
		return
	}
	if err := validateApplyMode(input.ApplyMode); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid apply mode"})
		return
	}

	session, err := h.service.CreateSession(r.Context(), CreateSessionInput{
		ProjectID:      chi.URLParam(r, "projectID"),
		Title:          input.Title,
		ActionKind:     input.ActionKind,
		ModelVariantID: input.ModelVariantID,
		ApplyMode:      input.ApplyMode,
		SkillIDs:       input.SkillIDs,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (h httpHandler) createSessionMessage(w http.ResponseWriter, r *http.Request) {
	var input chatMessageRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
		return
	}
	if strings.TrimSpace(input.BodyMarkdown) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "body markdown is required"})
		return
	}

	result, err := h.service.RunChatTurn(r.Context(), ChatTurnInput{
		ProjectID:    chi.URLParam(r, "projectID"),
		SessionID:    chi.URLParam(r, "sessionID"),
		BodyMarkdown: input.BodyMarkdown,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (h httpHandler) runContinuation(w http.ResponseWriter, r *http.Request) {
	var input continuationRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
		return
	}
	if !validateQuickActionRequest(w, input.ContentID, input.ModelVariantID, input.ApplyMode) {
		return
	}

	result, err := h.service.RunContinuation(r.Context(), ContinuationInput{
		ProjectID:         chi.URLParam(r, "projectID"),
		ContentID:         input.ContentID,
		ModelVariantID:    input.ModelVariantID,
		ApplyMode:         input.ApplyMode,
		Guidance:          input.Guidance,
		ExpectedRevision:  input.ExpectedRevision,
		InsertPosition:    input.InsertPosition,
		Insert:            input.Insert,
		ContinuationUnits: input.ContinuationUnits,
		ContinuationCount: input.ContinuationCount,
		SkillIDs:          input.SkillIDs,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h httpHandler) runRewrite(w http.ResponseWriter, r *http.Request) {
	var input rewriteRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
		return
	}
	if !validateQuickActionRequest(w, input.ContentID, input.ModelVariantID, input.ApplyMode) {
		return
	}

	result, err := h.service.RunRewrite(r.Context(), RewriteInput{
		ProjectID:        chi.URLParam(r, "projectID"),
		ContentID:        input.ContentID,
		ModelVariantID:   input.ModelVariantID,
		ApplyMode:        input.ApplyMode,
		Guidance:         input.Guidance,
		ExpectedRevision: input.ExpectedRevision,
		SelectionStart:   input.SelectionStart,
		SelectionEnd:     input.SelectionEnd,
		SkillIDs:         input.SkillIDs,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h httpHandler) runReadCheck(w http.ResponseWriter, r *http.Request) {
	var input rewriteRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
		return
	}
	if !validateQuickActionRequest(w, input.ContentID, input.ModelVariantID, input.ApplyMode) {
		return
	}

	result, err := h.service.RunReadAndCheck(r.Context(), ReadAndCheckInput{
		ProjectID:        chi.URLParam(r, "projectID"),
		ContentID:        input.ContentID,
		ModelVariantID:   input.ModelVariantID,
		ApplyMode:        input.ApplyMode,
		Guidance:         input.Guidance,
		ExpectedRevision: input.ExpectedRevision,
		SelectionStart:   input.SelectionStart,
		SelectionEnd:     input.SelectionEnd,
		SkillIDs:         input.SkillIDs,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h httpHandler) acceptCandidate(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.AcceptCandidate(r.Context(), AcceptCandidateInput{
		ProjectID:   chi.URLParam(r, "projectID"),
		CandidateID: chi.URLParam(r, "candidateID"),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h httpHandler) rejectCandidate(w http.ResponseWriter, r *http.Request) {
	candidate, err := h.service.RejectCandidate(r.Context(), RejectCandidateInput{
		ProjectID:   chi.URLParam(r, "projectID"),
		CandidateID: chi.URLParam(r, "candidateID"),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, candidate)
}

func (h httpHandler) listActivity(w http.ResponseWriter, r *http.Request) {
	events, err := h.service.ListActivityEvents(r.Context(), chi.URLParam(r, "projectID"), limitFromQuery(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (h httpHandler) listPromptRecords(w http.ResponseWriter, r *http.Request) {
	records, err := h.service.ListPromptRecords(r.Context(), chi.URLParam(r, "projectID"), limitFromQuery(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (h httpHandler) prunePromptRecords(w http.ResponseWriter, r *http.Request) {
	deleted, err := h.service.PrunePromptRecords(r.Context(), chi.URLParam(r, "projectID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, prunePromptRecordsResponse{Deleted: deleted})
}

func validateQuickActionRequest(w http.ResponseWriter, contentID, modelVariantID string, applyMode ApplyMode) bool {
	if strings.TrimSpace(contentID) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "content ID is required"})
		return false
	}
	if strings.TrimSpace(modelVariantID) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "model variant ID is required"})
		return false
	}
	if applyMode != "" {
		if err := validateApplyMode(applyMode); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid apply mode"})
			return false
		}
	}
	return true
}

func validateProviderConfigRequest(w http.ResponseWriter, name, baseURL, apiKey string, validateName bool) bool {
	if validateName && strings.TrimSpace(name) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "provider name is required"})
		return false
	}
	if strings.TrimSpace(baseURL) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "provider base URL is required"})
		return false
	}
	if err := validateBaseURL(baseURL); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid provider base URL"})
		return false
	}
	if strings.TrimSpace(apiKey) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "provider API key is required"})
		return false
	}
	return true
}

func validateBaseURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("base URL scheme must be http or https")
	}
	if parsed.Host == "" {
		return errors.New("base URL host is required")
	}
	return nil
}

func validateModelVariantRequest(w http.ResponseWriter, name, model string) bool {
	if strings.TrimSpace(name) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "model variant name is required"})
		return false
	}
	if strings.TrimSpace(model) == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "model is required"})
		return false
	}
	return true
}

func limitFromQuery(r *http.Request) int64 {
	value, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil || value <= 0 {
		return defaultListLimit
	}
	if value > 200 {
		return 200
	}
	return value
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
	case errors.Is(err, ErrInvalidInput), errors.Is(err, skill.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid input"})
	case errors.Is(err, project.ErrConflict):
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
