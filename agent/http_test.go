package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"git.inkyquill.net/inky/writer/project"
	"github.com/go-chi/chi/v5"
)

func TestProviderConfigHTTPEndpoints(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	service := NewService(db, project.NewService(db), nil)
	handler := newTestAgentHTTP(service)

	provider := postJSON[ProviderConfig](t, handler, "/api/provider-configs", `{
		"name": "DeepSeek",
		"baseUrl": "https://api.deepseek.com",
		"apiKey": "secret-key"
	}`, http.StatusCreated)
	if provider.AuthorID != placeholderAuthorID {
		t.Fatalf("provider author ID = %q, want %q", provider.AuthorID, placeholderAuthorID)
	}
	if provider.Name != "DeepSeek" || provider.BaseURL != "https://api.deepseek.com" {
		t.Fatalf("provider = %#v", provider)
	}

	updated := putJSON[ProviderConfig](t, handler, "/api/provider-configs/"+provider.ID, `{
		"baseUrl": "https://proxy.deepseek.test/v1",
		"apiKey": "rotated-key"
	}`, http.StatusOK)
	if updated.ID != provider.ID || updated.Name != provider.Name {
		t.Fatalf("updated provider = %#v, want same ID/name as %#v", updated, provider)
	}
	if updated.BaseURL != "https://proxy.deepseek.test/v1" {
		t.Fatalf("updated base URL = %q", updated.BaseURL)
	}

	providers := getJSON[[]ProviderConfig](t, handler, "/api/provider-configs", http.StatusOK)
	if len(providers) != 1 || providers[0].ID != provider.ID {
		t.Fatalf("providers = %#v, want created provider", providers)
	}

	model := postJSON[ModelVariant](t, handler, "/api/provider-configs/"+provider.ID+"/model-variants", `{
		"name": "DeepSeek Chat",
		"model": "deepseek-chat",
		"temperature": 0.4,
		"maxOutputTokens": 4096,
		"contextWindowTokens": 64000,
		"inputPricePerMillion": 0.27,
		"outputPricePerMillion": 1.10,
		"cacheReadPricePerMillion": 0.07,
		"cacheWritePricePerMillion": 0.27,
		"requestTokenField": "max_tokens",
		"reasoningFormat": "deepseek",
		"compatibilityJson": {"supportsTools": true}
	}`, http.StatusCreated)
	if model.ProviderConfigID != provider.ID || model.Model != "deepseek-chat" {
		t.Fatalf("model = %#v", model)
	}

	models := getJSON[[]ModelVariant](t, handler, "/api/provider-configs/"+provider.ID+"/model-variants", http.StatusOK)
	if len(models) != 1 || models[0].ID != model.ID {
		t.Fatalf("models = %#v, want created model", models)
	}

	var storedKey string
	if err := db.QueryRowContext(ctx, `SELECT api_key_encrypted FROM provider_configs WHERE id = ?`, provider.ID).Scan(&storedKey); err != nil {
		t.Fatalf("read stored API key: %v", err)
	}
	if storedKey != "rotated-key" {
		t.Fatalf("stored key = %q, want rotated-key", storedKey)
	}
}

func TestProjectAgentHTTPEndpoints(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "HTTP Agent Project")
	chapter := createTestChapter(t, ctx, projectService, storyProject.ID, "Opening", "The old city waited.")
	provider := &fakeChatProvider{responses: []CompletionResponse{
		quickActionResponse("chat-1", "Keep the opening focused on the city."),
		quickActionResponse("continuation-1", "\n\nA new paragraph arrived."),
		quickActionResponse("rewrite-1", "bright city"),
		quickActionResponse("read-check-1", "Continuity note: check the city name."),
	}}
	service := NewService(db, projectService, provider)
	model := createTestProviderAndModel(t, ctx, service, "author-1")
	handler := newTestAgentHTTP(service)

	profile := putJSON[PromptProfile](t, handler, "/api/projects/"+storyProject.ID+"/agent/prompt-profile", `{
		"genre": "fantasy",
		"tense": "past tense",
		"pov": "third person",
		"voice": "precise",
		"instructionsMarkdown": "Keep continuity tight.",
		"promptRecordRetentionDays": 30
	}`, http.StatusOK)
	if profile.ProjectID != storyProject.ID || profile.PromptRecordRetentionDays != 30 {
		t.Fatalf("profile = %#v", profile)
	}
	gotProfile := getJSON[PromptProfile](t, handler, "/api/projects/"+storyProject.ID+"/agent/prompt-profile", http.StatusOK)
	if gotProfile.ID != profile.ID {
		t.Fatalf("got profile ID = %q, want %q", gotProfile.ID, profile.ID)
	}

	session := postJSON[Session](t, handler, "/api/projects/"+storyProject.ID+"/agent/sessions", `{
		"title": "Opening chat",
		"actionKind": "chat",
		"modelVariantId": "`+model.ID+`",
		"applyMode": "preview"
	}`, http.StatusCreated)
	if session.ActionKind != ActionKindChat || session.ModelVariantID != model.ID {
		t.Fatalf("session = %#v", session)
	}
	sessions := getJSON[[]Session](t, handler, "/api/projects/"+storyProject.ID+"/agent/sessions?limit=5", http.StatusOK)
	if len(sessions) != 1 || sessions[0].ID != session.ID {
		t.Fatalf("sessions = %#v, want created session", sessions)
	}

	chat := postJSON[ChatTurnResult](t, handler, "/api/projects/"+storyProject.ID+"/agent/sessions/"+session.ID+"/messages", `{
		"bodyMarkdown": "What should I keep in mind?"
	}`, http.StatusCreated)
	if chat.UserMessage.Role != MessageRoleUser || chat.AssistantMessage.BodyMarkdown != "Keep the opening focused on the city." {
		t.Fatalf("chat result = %#v", chat)
	}
	if chat.PromptRecordID == "" {
		t.Fatal("chat prompt record ID is empty")
	}

	continuation := postJSON[ContinuationResult](t, handler, "/api/projects/"+storyProject.ID+"/agent/actions/continuation", `{
		"contentId": "`+chapter.ID+`",
		"modelVariantId": "`+model.ID+`",
		"applyMode": "preview",
		"expectedRevision": 1,
		"continuationUnits": "paragraph",
		"continuationCount": 1
	}`, http.StatusOK)
	if continuation.Candidate.ID == "" || continuation.Candidate.Status != "pending" {
		t.Fatalf("continuation = %#v", continuation)
	}
	accepted := postJSON[AcceptCandidateResult](t, handler, "/api/projects/"+storyProject.ID+"/agent/candidates/"+continuation.Candidate.ID+"/accept", `{}`, http.StatusOK)
	if accepted.Candidate.Status != "accepted" || !strings.Contains(accepted.Content.BodyMarkdown, "A new paragraph arrived.") {
		t.Fatalf("accepted = %#v", accepted)
	}

	start := strings.Index(chapter.BodyMarkdown, "old city")
	end := start + len("old city")
	rewrite := postJSON[RewriteResult](t, handler, "/api/projects/"+storyProject.ID+"/agent/actions/rewrite", `{
		"contentId": "`+chapter.ID+`",
		"modelVariantId": "`+model.ID+`",
		"applyMode": "preview",
		"expectedRevision": 2,
		"selectionStart": `+strconv.Itoa(start)+`,
		"selectionEnd": `+strconv.Itoa(end)+`
	}`, http.StatusOK)
	if rewrite.Candidate.ID == "" || rewrite.Candidate.GeneratedMarkdown != "bright city" {
		t.Fatalf("rewrite = %#v", rewrite)
	}
	rejected := postJSON[GenerationCandidate](t, handler, "/api/projects/"+storyProject.ID+"/agent/candidates/"+rewrite.Candidate.ID+"/reject", `{}`, http.StatusOK)
	if rejected.Status != "rejected" {
		t.Fatalf("rejected status = %q, want rejected", rejected.Status)
	}

	readCheck := postJSON[ReadAndCheckResult](t, handler, "/api/projects/"+storyProject.ID+"/agent/actions/read-check", `{
		"contentId": "`+chapter.ID+`",
		"modelVariantId": "`+model.ID+`",
		"expectedRevision": 2,
		"selectionStart": 4,
		"selectionEnd": 15,
		"guidance": "Check continuity."
	}`, http.StatusOK)
	if readCheck.AssistantMessage.BodyMarkdown != "Continuity note: check the city name." || readCheck.Note.Source != "read_and_check" {
		t.Fatalf("read check = %#v", readCheck)
	}

	events := getJSON[[]ActivityEvent](t, handler, "/api/projects/"+storyProject.ID+"/agent/activity?limit=10", http.StatusOK)
	if len(events) == 0 {
		t.Fatal("activity events = 0, want at least one")
	}

	records := getJSON[[]PromptRecord](t, handler, "/api/projects/"+storyProject.ID+"/agent/prompt-records?limit=10", http.StatusOK)
	if len(records) != 1 || records[0].ID != chat.PromptRecordID {
		t.Fatalf("prompt records = %#v, want chat prompt record", records)
	}
	pruned := postJSON[struct {
		Deleted int64 `json:"deleted"`
	}](t, handler, "/api/projects/"+storyProject.ID+"/agent/prompt-records/prune", `{}`, http.StatusOK)
	if pruned.Deleted != 0 {
		t.Fatalf("pruned deleted = %d, want 0 for fresh retained record", pruned.Deleted)
	}
}

func TestAgentHTTPMapsClientErrors(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db, project.NewService(db), nil)
	handler := newTestAgentHTTP(service)

	req := httptest.NewRequest(http.MethodPost, "/api/provider-configs", bytes.NewBufferString(`{"name":"Bad"} trailing`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed JSON status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/projects/missing/agent/prompt-profile", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing profile status = %d, want %d; body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestQuickActionHTTPValidationErrorsReturnBadRequest(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Quick Action Validation Project")
	chapter := createTestChapter(t, ctx, projectService, storyProject.ID, "Opening", "The old city waited.")
	utf8Chapter := createTestChapter(t, ctx, projectService, storyProject.ID, "Accent", "éclair")
	provider := &fakeChatProvider{responses: []CompletionResponse{
		quickActionResponse("unused-1", "unused"),
		quickActionResponse("unused-2", "unused"),
	}}
	service := NewService(db, projectService, provider)
	model := createTestProviderAndModel(t, ctx, service, "author-1")
	handler := newTestAgentHTTP(service)
	entry, err := projectService.CreateContent(ctx, project.CreateContentInput{
		ProjectID:    storyProject.ID,
		Kind:         project.KindStoryBibleEntry,
		Title:        "Mira",
		BodyMarkdown: "An alchemist.",
		Reason:       "seed",
		CreatedBy:    "author",
	})
	if err != nil {
		t.Fatalf("CreateContent(entry) error = %v", err)
	}

	tests := []struct {
		name string
		path string
		body string
	}{
		{
			name: "continuation missing model variant",
			path: "/api/projects/" + storyProject.ID + "/agent/actions/continuation",
			body: `{"contentId":"` + chapter.ID + `","expectedRevision":1}`,
		},
		{
			name: "continuation invalid apply mode",
			path: "/api/projects/" + storyProject.ID + "/agent/actions/continuation",
			body: `{"contentId":"` + chapter.ID + `","modelVariantId":"model-1","applyMode":"invalid","expectedRevision":1}`,
		},
		{
			name: "rewrite missing model variant",
			path: "/api/projects/" + storyProject.ID + "/agent/actions/rewrite",
			body: `{"contentId":"` + chapter.ID + `","expectedRevision":1,"selectionStart":4,"selectionEnd":12}`,
		},
		{
			name: "rewrite invalid apply mode",
			path: "/api/projects/" + storyProject.ID + "/agent/actions/rewrite",
			body: `{"contentId":"` + chapter.ID + `","modelVariantId":"model-1","applyMode":"invalid","expectedRevision":1,"selectionStart":4,"selectionEnd":12}`,
		},
		{
			name: "read check missing model variant",
			path: "/api/projects/" + storyProject.ID + "/agent/actions/read-check",
			body: `{"contentId":"` + chapter.ID + `","expectedRevision":1,"selectionStart":4,"selectionEnd":12}`,
		},
		{
			name: "read check invalid apply mode",
			path: "/api/projects/" + storyProject.ID + "/agent/actions/read-check",
			body: `{"contentId":"` + chapter.ID + `","modelVariantId":"model-1","applyMode":"invalid","expectedRevision":1,"selectionStart":4,"selectionEnd":12}`,
		},
		{
			name: "rewrite invalid selection range",
			path: "/api/projects/" + storyProject.ID + "/agent/actions/rewrite",
			body: `{"contentId":"` + chapter.ID + `","modelVariantId":"` + model.ID + `","expectedRevision":1,"selectionStart":12,"selectionEnd":4}`,
		},
		{
			name: "read check utf8 boundary",
			path: "/api/projects/" + storyProject.ID + "/agent/actions/read-check",
			body: `{"contentId":"` + utf8Chapter.ID + `","modelVariantId":"` + model.ID + `","expectedRevision":1,"selectionStart":1,"selectionEnd":2}`,
		},
		{
			name: "rewrite non chapter target",
			path: "/api/projects/" + storyProject.ID + "/agent/actions/rewrite",
			body: `{"contentId":"` + entry.ID + `","modelVariantId":"` + model.ID + `","expectedRevision":1,"selectionStart":0,"selectionEnd":2}`,
		},
		{
			name: "continuation preview invalid insert position",
			path: "/api/projects/" + storyProject.ID + "/agent/actions/continuation",
			body: `{"contentId":"` + chapter.ID + `","modelVariantId":"` + model.ID + `","applyMode":"preview","expectedRevision":1,"insert":true,"insertPosition":999}`,
		},
		{
			name: "continuation direct apply invalid insert position",
			path: "/api/projects/" + storyProject.ID + "/agent/actions/continuation",
			body: `{"contentId":"` + chapter.ID + `","modelVariantId":"` + model.ID + `","applyMode":"direct_apply","expectedRevision":1,"insert":true,"insertPosition":999}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("content type = %q, want application/json", got)
			}
		})
	}
	if len(provider.requests) != 0 {
		t.Fatalf("provider request count = %d, want 0 for invalid quick-action requests", len(provider.requests))
	}
}

func TestProviderAndModelHTTPValidationAndConflicts(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db, project.NewService(db), nil)
	handler := newTestAgentHTTP(service)

	badProviderRequests := []string{
		`{"name":"","baseUrl":"https://provider.test","apiKey":"secret"}`,
		`{"name":"Provider","baseUrl":"","apiKey":"secret"}`,
		`{"name":"Provider","baseUrl":"://bad-url","apiKey":"secret"}`,
		`{"name":"Provider","baseUrl":"https://provider.test","apiKey":""}`,
	}
	for _, body := range badProviderRequests {
		req := httptest.NewRequest(http.MethodPost, "/api/provider-configs", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("provider validation status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	}

	provider := postJSON[ProviderConfig](t, handler, "/api/provider-configs", `{
		"name": "Duplicate Provider",
		"baseUrl": "https://provider.test",
		"apiKey": "secret"
	}`, http.StatusCreated)
	_ = postJSON[errorResponse](t, handler, "/api/provider-configs", `{
		"name": "Duplicate Provider",
		"baseUrl": "https://other-provider.test",
		"apiKey": "secret"
	}`, http.StatusConflict)

	req := httptest.NewRequest(http.MethodPut, "/api/provider-configs/"+provider.ID, bytes.NewBufferString(`{"baseUrl":"","apiKey":"secret"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("provider update validation status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	badModelRequests := []string{
		`{"name":"","model":"chat-model"}`,
		`{"name":"Chat","model":""}`,
	}
	for _, body := range badModelRequests {
		req := httptest.NewRequest(http.MethodPost, "/api/provider-configs/"+provider.ID+"/model-variants", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("model validation status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	}

	_ = postJSON[ModelVariant](t, handler, "/api/provider-configs/"+provider.ID+"/model-variants", `{
		"name": "Duplicate Model",
		"model": "chat-model"
	}`, http.StatusCreated)
	_ = postJSON[errorResponse](t, handler, "/api/provider-configs/"+provider.ID+"/model-variants", `{
		"name": "Duplicate Model",
		"model": "other-chat-model"
	}`, http.StatusConflict)
}

func newTestAgentHTTP(service *Service) http.Handler {
	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		RegisterRoutes(r, service)
	})
	return r
}

func getJSON[T any](t *testing.T, handler http.Handler, path string, wantStatus int) T {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return decodeTestResponse[T](t, rec, wantStatus)
}

func postJSON[T any](t *testing.T, handler http.Handler, path string, body string, wantStatus int) T {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return decodeTestResponse[T](t, rec, wantStatus)
}

func putJSON[T any](t *testing.T, handler http.Handler, path string, body string, wantStatus int) T {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return decodeTestResponse[T](t, rec, wantStatus)
}

func decodeTestResponse[T any](t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) T {
	t.Helper()
	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, wantStatus, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q, want application/json", got)
	}
	var value T
	if err := json.NewDecoder(rec.Body).Decode(&value); err != nil {
		t.Fatalf("decode response: %v; body = %s", err, rec.Body.String())
	}
	return value
}
