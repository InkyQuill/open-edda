package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"git.inkyquill.net/inky/writer/project"
	"git.inkyquill.net/inky/writer/store"
	"github.com/pressly/goose/v3"
)

func TestRunChatTurnStoresMessagesToolActivityAndPromptRecord(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Chat Turn Project")
	provider := &fakeChatProvider{
		responses: []CompletionResponse{
			{
				ID: "completion-tool",
				Message: CompletionMessage{
					Role: MessageRoleAssistant,
					ToolCalls: []CompletionToolCall{
						{
							ID:   "tool-call-1",
							Type: "function",
							Function: CompletionToolCallFunction{
								Name:      "project_map",
								Arguments: `{}`,
							},
						},
					},
				},
				FinishReason: "tool_calls",
				Usage: Usage{
					InputTokens:      90,
					OutputTokens:     12,
					CacheReadTokens:  10,
					CacheWriteTokens: 5,
					TotalTokens:      117,
				},
				UsageAvailable: true,
			},
			{
				ID: "completion-final",
				Message: CompletionMessage{
					Role:    MessageRoleAssistant,
					Content: "The city map shows one opening chapter.",
				},
				FinishReason: "stop",
				Usage: Usage{
					InputTokens:      110,
					OutputTokens:     30,
					CacheReadTokens:  20,
					CacheWriteTokens: 10,
					TotalTokens:      170,
				},
				UsageAvailable: true,
			},
		},
	}
	service := NewService(db, projectService, provider)
	model := createTestProviderAndModel(t, ctx, service, "author-1")
	session, err := service.CreateSession(ctx, CreateSessionInput{
		ProjectID:      storyProject.ID,
		Title:          "Opening chat",
		ActionKind:     ActionKindChat,
		ModelVariantID: model.ID,
		ApplyMode:      ApplyModePreview,
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	createPromptProfileWithRetention(t, ctx, service, storyProject.ID, 30)
	if _, err := projectService.CreateContent(ctx, project.CreateContentInput{
		ProjectID:    storyProject.ID,
		Kind:         project.KindChapter,
		Title:        "Opening",
		BodyMarkdown: "The glass city waited.",
		Reason:       "seed",
		CreatedBy:    "author",
	}); err != nil {
		t.Fatalf("CreateContent() error = %v", err)
	}

	result, err := service.RunChatTurn(ctx, ChatTurnInput{
		ProjectID:    storyProject.ID,
		SessionID:    session.ID,
		BodyMarkdown: "What context should I keep in mind?",
	})
	if err != nil {
		t.Fatalf("RunChatTurn() error = %v", err)
	}

	if result.AssistantMessage.BodyMarkdown != "The city map shows one opening chapter." {
		t.Fatalf("assistant body = %q", result.AssistantMessage.BodyMarkdown)
	}
	if result.PromptRecordID == "" {
		t.Fatal("PromptRecordID is empty")
	}
	if len(provider.requests) != 2 {
		t.Fatalf("provider request count = %d, want 2", len(provider.requests))
	}
	if len(provider.requests[0].Tools) == 0 {
		t.Fatal("first provider request did not include tools")
	}
	if got := provider.requests[1].Messages[len(provider.requests[1].Messages)-1]; got.Role != MessageRoleTool || !strings.Contains(got.Content, "Chat Turn Project") {
		t.Fatalf("last second-round message = %#v, want project map tool result", got)
	}

	messages, err := service.ListMessages(ctx, storyProject.ID, session.ID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages) != 3 {
		t.Fatalf("message count = %d, want 3", len(messages))
	}
	if messages[0].Role != MessageRoleUser || messages[0].BodyMarkdown != "What context should I keep in mind?" {
		t.Fatalf("user message = %#v", messages[0])
	}
	if messages[1].Role != MessageRoleTool || !strings.Contains(messages[1].BodyMarkdown, "project_map") {
		t.Fatalf("tool message = %#v", messages[1])
	}
	if messages[2].Role != MessageRoleAssistant || messages[2].BodyMarkdown != "The city map shows one opening chapter." {
		t.Fatalf("assistant message = %#v", messages[2])
	}

	queries := store.New(db)
	events, err := queries.ListActivityEvents(ctx, store.ListActivityEventsParams{ProjectID: storyProject.ID, Limit: 10})
	if err != nil {
		t.Fatalf("ListActivityEvents() error = %v", err)
	}
	if len(events) == 0 {
		t.Fatal("activity event count = 0, want at least one")
	}

	records, err := queries.ListPromptRecords(ctx, store.ListPromptRecordsParams{ProjectID: storyProject.ID, Limit: 10})
	if err != nil {
		t.Fatalf("ListPromptRecords() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("prompt record count = %d, want 1", len(records))
	}
	record := records[0]
	if record.ID != result.PromptRecordID {
		t.Fatalf("prompt record ID = %q, want %q", record.ID, result.PromptRecordID)
	}
	if record.ProviderName != "Fake Provider" {
		t.Fatalf("provider name = %q, want Fake Provider", record.ProviderName)
	}
	if record.ModelName != "fake-chat" {
		t.Fatalf("model name = %q, want fake-chat", record.ModelName)
	}
	if record.InputTokens != 200 || record.OutputTokens != 42 || record.CacheReadTokens != 30 || record.CacheWriteTokens != 15 || record.TotalTokens != 287 {
		t.Fatalf("stored usage = input %d output %d cache read %d cache write %d total %d", record.InputTokens, record.OutputTokens, record.CacheReadTokens, record.CacheWriteTokens, record.TotalTokens)
	}
	assertFloat(t, record.InputCost, 0.0002)
	assertFloat(t, record.OutputCost, 0.000084)
	assertFloat(t, record.CacheReadCost, 0.000003)
	assertFloat(t, record.CacheWriteCost, 0.00003)
	assertFloat(t, record.TotalCost, 0.000317)
	if strings.Contains(record.RequestJson, "secret-key") || strings.Contains(record.RequestJson, "Authorization") {
		t.Fatalf("request JSON contains secret material: %s", record.RequestJson)
	}
	if !json.Valid([]byte(record.RequestJson)) {
		t.Fatalf("request JSON is invalid: %s", record.RequestJson)
	}
	if !json.Valid([]byte(record.ResponseJson)) {
		t.Fatalf("response JSON is invalid: %s", record.ResponseJson)
	}

	snapshots, err := queries.ListPromptContextSnapshots(ctx, record.ID)
	if err != nil {
		t.Fatalf("ListPromptContextSnapshots() error = %v", err)
	}
	if len(snapshots) == 0 {
		t.Fatal("snapshot count = 0, want prompt context snapshots")
	}
	assertSnapshotSource(t, snapshots, "prompt_profile")
	assertSnapshotSource(t, snapshots, "transcript")
}

func TestRunChatTurnStoresZeroUsageAndEventWhenProviderOmitsUsage(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Usage Missing Project")
	provider := &fakeChatProvider{
		responses: []CompletionResponse{
			{
				ID: "completion-no-usage",
				Message: CompletionMessage{
					Role:    MessageRoleAssistant,
					Content: "No usage came back.",
				},
				FinishReason: "stop",
			},
		},
	}
	service := NewService(db, projectService, provider)
	model := createTestProviderAndModel(t, ctx, service, "author-1")
	session := createChatTurnTestSessionWithModel(t, ctx, service, storyProject.ID, model.ID)
	createPromptProfileWithRetention(t, ctx, service, storyProject.ID, 30)

	if _, err := service.RunChatTurn(ctx, ChatTurnInput{
		ProjectID:    storyProject.ID,
		SessionID:    session.ID,
		BodyMarkdown: "Please answer.",
	}); err != nil {
		t.Fatalf("RunChatTurn() error = %v", err)
	}

	queries := store.New(db)
	records, err := queries.ListPromptRecords(ctx, store.ListPromptRecordsParams{ProjectID: storyProject.ID, Limit: 10})
	if err != nil {
		t.Fatalf("ListPromptRecords() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("prompt record count = %d, want 1", len(records))
	}
	if records[0].TotalTokens != 0 || records[0].TotalCost != 0 {
		t.Fatalf("usage/cost = tokens %d cost %f, want zeros", records[0].TotalTokens, records[0].TotalCost)
	}
	events, err := queries.ListActivityEvents(ctx, store.ListActivityEventsParams{ProjectID: storyProject.ID, Limit: 10})
	if err != nil {
		t.Fatalf("ListActivityEvents() error = %v", err)
	}
	if !hasEventType(events, "usage_missing") {
		t.Fatalf("events did not include usage_missing: %#v", events)
	}
}

func TestPromptRecordRetentionZeroDisablesStorage(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Retention Zero Project")
	provider := &fakeChatProvider{
		responses: []CompletionResponse{
			{
				ID: "completion-final",
				Message: CompletionMessage{
					Role:    MessageRoleAssistant,
					Content: "Storage is disabled.",
				},
				FinishReason:   "stop",
				UsageAvailable: true,
			},
		},
	}
	service := NewService(db, projectService, provider)
	model := createTestProviderAndModel(t, ctx, service, "author-1")
	session := createChatTurnTestSessionWithModel(t, ctx, service, storyProject.ID, model.ID)
	createPromptProfileWithRetention(t, ctx, service, storyProject.ID, 0)

	if _, err := service.RunChatTurn(ctx, ChatTurnInput{
		ProjectID:    storyProject.ID,
		SessionID:    session.ID,
		BodyMarkdown: "Do not store a prompt record.",
	}); err != nil {
		t.Fatalf("RunChatTurn() error = %v", err)
	}
	records, err := store.New(db).ListPromptRecords(ctx, store.ListPromptRecordsParams{ProjectID: storyProject.ID, Limit: 10})
	if err != nil {
		t.Fatalf("ListPromptRecords() error = %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("prompt record count = %d, want 0", len(records))
	}
}

func TestPrunePromptRecordsDeletesRecordsOlderThanRetentionWindow(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Prune Project")
	service := NewService(db, projectService, nil)
	createPromptProfileWithRetention(t, ctx, service, storyProject.ID, 30)
	insertPromptRecordForPrune(t, ctx, db, storyProject.ID, "old-record", "2000-01-01T00:00:00Z")
	insertPromptRecordForPrune(t, ctx, db, storyProject.ID, "new-record", nowString())

	deleted, err := service.PrunePromptRecords(ctx, storyProject.ID)
	if err != nil {
		t.Fatalf("PrunePromptRecords() error = %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted rows = %d, want 1", deleted)
	}
	records, err := store.New(db).ListPromptRecords(ctx, store.ListPromptRecordsParams{ProjectID: storyProject.ID, Limit: 10})
	if err != nil {
		t.Fatalf("ListPromptRecords() error = %v", err)
	}
	if len(records) != 1 || records[0].ID != "new-record" {
		t.Fatalf("remaining records = %#v, want only new-record", records)
	}

	createPromptProfileWithRetention(t, ctx, service, storyProject.ID, 0)
	deleted, err = service.PrunePromptRecords(ctx, storyProject.ID)
	if err != nil {
		t.Fatalf("PrunePromptRecords(retention zero) error = %v", err)
	}
	if deleted != 1 {
		t.Fatalf("retention zero deleted rows = %d, want 1", deleted)
	}
}

func TestCreateSessionAndMessages(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()

	projectService := project.NewService(db)
	storyProject, err := projectService.CreateProject(ctx, project.CreateProjectInput{
		AuthorID: "author-1",
		Title:    "Agent Test Project",
		Language: "en",
	})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	service := NewService(db, projectService, nil)
	session, err := service.CreateSession(ctx, CreateSessionInput{
		ProjectID:  storyProject.ID,
		Title:      "Opening chat",
		ActionKind: ActionKindChat,
		ApplyMode:  ApplyModePreview,
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	userMessage, err := service.AppendMessage(ctx, AppendMessageInput{
		ProjectID:    storyProject.ID,
		SessionID:    session.ID,
		Role:         MessageRoleUser,
		BodyMarkdown: "Can you help with this opening?",
	})
	if err != nil {
		t.Fatalf("AppendMessage(user) error = %v", err)
	}

	assistantMessage, err := service.AppendMessage(ctx, AppendMessageInput{
		ProjectID:    storyProject.ID,
		SessionID:    session.ID,
		Role:         MessageRoleAssistant,
		BodyMarkdown: "Yes. Start with a sharper image.",
		MetadataJSON: `{"finishReason":"stop"}`,
	})
	if err != nil {
		t.Fatalf("AppendMessage(assistant) error = %v", err)
	}

	messages, err := service.ListMessages(ctx, storyProject.ID, session.ID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("ListMessages() count = %d, want 2", len(messages))
	}
	if messages[0].ID != userMessage.ID {
		t.Fatalf("first message ID = %q, want %q", messages[0].ID, userMessage.ID)
	}
	if messages[0].Role != MessageRoleUser {
		t.Fatalf("first message role = %q, want %q", messages[0].Role, MessageRoleUser)
	}
	if messages[0].BodyMarkdown != "Can you help with this opening?" {
		t.Fatalf("first message body = %q", messages[0].BodyMarkdown)
	}
	if messages[0].MetadataJSON != "{}" {
		t.Fatalf("first message metadata = %q, want {}", messages[0].MetadataJSON)
	}
	if messages[1].ID != assistantMessage.ID {
		t.Fatalf("second message ID = %q, want %q", messages[1].ID, assistantMessage.ID)
	}
	if messages[1].Role != MessageRoleAssistant {
		t.Fatalf("second message role = %q, want %q", messages[1].Role, MessageRoleAssistant)
	}
	if messages[1].BodyMarkdown != "Yes. Start with a sharper image." {
		t.Fatalf("second message body = %q", messages[1].BodyMarkdown)
	}
	if messages[1].MetadataJSON != `{"finishReason":"stop"}` {
		t.Fatalf("second message metadata = %q", messages[1].MetadataJSON)
	}
}

func TestCreateSessionRejectsInvalidInput(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Validation Project")
	service := NewService(db, projectService, nil)

	tests := []struct {
		name  string
		input CreateSessionInput
	}{
		{
			name: "invalid action kind",
			input: CreateSessionInput{
				ProjectID:  storyProject.ID,
				Title:      "Bad action",
				ActionKind: "invalid",
				ApplyMode:  ApplyModePreview,
			},
		},
		{
			name: "invalid apply mode",
			input: CreateSessionInput{
				ProjectID:  storyProject.ID,
				Title:      "Bad apply",
				ActionKind: ActionKindChat,
				ApplyMode:  "invalid",
			},
		},
		{
			name: "missing project",
			input: CreateSessionInput{
				ProjectID:  "missing-project",
				Title:      "Missing project",
				ActionKind: ActionKindChat,
				ApplyMode:  ApplyModePreview,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateSession(ctx, tt.input)
			if err == nil {
				t.Fatal("CreateSession() error = nil, want error")
			}
		})
	}
}

func TestAppendMessageRejectsInvalidRole(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Message Validation")
	service := NewService(db, projectService, nil)
	session := createTestSession(t, ctx, service, storyProject.ID)

	_, err := service.AppendMessage(ctx, AppendMessageInput{
		ProjectID:    storyProject.ID,
		SessionID:    session.ID,
		Role:         "invalid",
		BodyMarkdown: "not accepted",
	})
	if err == nil {
		t.Fatal("AppendMessage() error = nil, want invalid role error")
	}
}

func TestMessagesRequireProjectOwnership(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	firstProject := createTestProject(t, ctx, projectService, "author-1", "First Project")
	secondProject := createTestProject(t, ctx, projectService, "author-1", "Second Project")
	service := NewService(db, projectService, nil)
	session := createTestSession(t, ctx, service, firstProject.ID)

	_, err := service.AppendMessage(ctx, AppendMessageInput{
		ProjectID:    secondProject.ID,
		SessionID:    session.ID,
		Role:         MessageRoleUser,
		BodyMarkdown: "wrong project",
	})
	if err == nil {
		t.Fatal("AppendMessage() error = nil, want wrong-project error")
	}

	messages, err := service.ListMessages(ctx, secondProject.ID, session.ID)
	if err == nil {
		t.Fatalf("ListMessages() error = nil, messages = %v, want wrong-project error", messages)
	}
}

func TestCreateSessionRejectsCrossAuthorModelVariant(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	authorProject := createTestProject(t, ctx, projectService, "author-1", "Author Project")
	otherModelID := seedOtherAuthorModelVariant(t, db)
	service := NewService(db, projectService, nil)

	_, err := service.CreateSession(ctx, CreateSessionInput{
		ProjectID:      authorProject.ID,
		Title:          "Wrong model",
		ActionKind:     ActionKindChat,
		ModelVariantID: otherModelID,
		ApplyMode:      ApplyModePreview,
	})
	if err == nil {
		t.Fatal("CreateSession() error = nil, want cross-author model error")
	}
}

func TestProviderConfigServiceSavesConfigsWithoutReturningAPIKey(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	service := NewService(db, project.NewService(db), nil)

	provider, err := service.CreateProviderConfig(ctx, CreateProviderConfigInput{
		AuthorID: "author-1",
		Name:     "DeepSeek",
		BaseURL:  "https://api.deepseek.com",
		APIKey:   "secret-key",
	})
	if err != nil {
		t.Fatalf("CreateProviderConfig() error = %v", err)
	}

	if provider.Name != "DeepSeek" {
		t.Fatalf("provider name = %q, want DeepSeek", provider.Name)
	}
	if provider.BaseURL != "https://api.deepseek.com" {
		t.Fatalf("provider base URL = %q", provider.BaseURL)
	}

	providers, err := service.ListProviderConfigs(ctx, "author-1")
	if err != nil {
		t.Fatalf("ListProviderConfigs() error = %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("provider count = %d, want 1", len(providers))
	}
	if providers[0] != provider {
		t.Fatalf("listed provider = %#v, want %#v", providers[0], provider)
	}

	got, err := service.GetProviderConfig(ctx, "author-1", provider.ID)
	if err != nil {
		t.Fatalf("GetProviderConfig() error = %v", err)
	}
	if got != provider {
		t.Fatalf("got provider = %#v, want %#v", got, provider)
	}

	var storedKey string
	if err := db.QueryRowContext(ctx, `SELECT api_key_encrypted FROM provider_configs WHERE id = ?`, provider.ID).Scan(&storedKey); err != nil {
		t.Fatalf("read stored key: %v", err)
	}
	if storedKey != "secret-key" {
		t.Fatalf("stored key = %q, want plaintext key for this milestone", storedKey)
	}
}

func TestModelVariantServiceSavesPricingCompatibilityAndProjectSelection(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Provider Project")
	service := NewService(db, projectService, nil)

	provider, err := service.CreateProviderConfig(ctx, CreateProviderConfigInput{
		AuthorID: "author-1",
		Name:     "DeepSeek",
		BaseURL:  "https://api.deepseek.com",
		APIKey:   "secret-key",
	})
	if err != nil {
		t.Fatalf("CreateProviderConfig() error = %v", err)
	}

	chat, err := service.CreateModelVariant(ctx, CreateModelVariantInput{
		AuthorID:                  "author-1",
		ProviderConfigID:          provider.ID,
		Name:                      "DeepSeek Chat",
		Model:                     "deepseek-chat",
		Temperature:               0.5,
		MaxOutputTokens:           4096,
		ContextWindowTokens:       64000,
		InputPricePerMillion:      0.27,
		OutputPricePerMillion:     1.10,
		CacheReadPricePerMillion:  0.07,
		CacheWritePricePerMillion: 0.27,
		RequestTokenField:         "max_tokens",
		ReasoningFormat:           "deepseek",
		CompatibilityJSON:         `{"supportsTools":true}`,
	})
	if err != nil {
		t.Fatalf("CreateModelVariant(chat) error = %v", err)
	}
	coder, err := service.CreateModelVariant(ctx, CreateModelVariantInput{
		AuthorID:                  "author-1",
		ProviderConfigID:          provider.ID,
		Name:                      "DeepSeek Coder",
		Model:                     "deepseek-coder",
		Temperature:               0.2,
		MaxOutputTokens:           2048,
		ContextWindowTokens:       32000,
		InputPricePerMillion:      0.14,
		OutputPricePerMillion:     0.28,
		CacheReadPricePerMillion:  0.014,
		CacheWritePricePerMillion: 0.14,
		RequestTokenField:         "max_tokens",
		CompatibilityJSON:         "{}",
	})
	if err != nil {
		t.Fatalf("CreateModelVariant(coder) error = %v", err)
	}

	variants, err := service.ListModelVariantsByProvider(ctx, "author-1", provider.ID)
	if err != nil {
		t.Fatalf("ListModelVariantsByProvider() error = %v", err)
	}
	if len(variants) != 2 {
		t.Fatalf("variant count = %d, want 2", len(variants))
	}

	got, err := service.GetModelVariant(ctx, "author-1", chat.ID)
	if err != nil {
		t.Fatalf("GetModelVariant() error = %v", err)
	}
	if got.InputPricePerMillion != 0.27 ||
		got.OutputPricePerMillion != 1.10 ||
		got.CacheReadPricePerMillion != 0.07 ||
		got.CacheWritePricePerMillion != 0.27 {
		t.Fatalf("pricing fields not preserved: %#v", got)
	}
	if got.RequestTokenField != "max_tokens" {
		t.Fatalf("request token field = %q, want max_tokens", got.RequestTokenField)
	}
	if got.ReasoningFormat != "deepseek" {
		t.Fatalf("reasoning format = %q, want deepseek", got.ReasoningFormat)
	}
	if got.ContextWindowTokens != 64000 {
		t.Fatalf("context window tokens = %d, want 64000", got.ContextWindowTokens)
	}
	if got.MaxOutputTokens != 4096 {
		t.Fatalf("max output tokens = %d, want 4096", got.MaxOutputTokens)
	}

	selected, err := service.GetModelVariantForProject(ctx, storyProject.ID, chat.ID)
	if err != nil {
		t.Fatalf("GetModelVariantForProject() owner error = %v", err)
	}
	if selected.ID != chat.ID {
		t.Fatalf("selected variant ID = %q, want %q", selected.ID, chat.ID)
	}

	otherModelID := seedOtherAuthorModelVariant(t, db)
	if _, err := service.GetModelVariantForProject(ctx, storyProject.ID, otherModelID); err == nil {
		t.Fatal("GetModelVariantForProject() with cross-author model error = nil, want error")
	}

	if coder.ID == chat.ID {
		t.Fatal("two created variants have the same ID")
	}
}

func openMigratedTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbName := strings.ReplaceAll(t.Name(), "/", "-")
	db, err := store.Open("file:" + dbName + "?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}
	if err := goose.Up(db, filepath.Join("..", "migrations")); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO authors (id, email, password_hash, created_at)
		VALUES ('author-1', 'author@example.com', 'hash', '2026-06-13T00:00:00Z');
	`)
	if err != nil {
		t.Fatalf("seed author: %v", err)
	}

	return db
}

func createTestProject(t *testing.T, ctx context.Context, service *project.Service, authorID string, title string) project.StoryProject {
	t.Helper()

	storyProject, err := service.CreateProject(ctx, project.CreateProjectInput{
		AuthorID: authorID,
		Title:    title,
		Language: "en",
	})
	if err != nil {
		t.Fatalf("CreateProject(%q) error = %v", title, err)
	}
	return storyProject
}

func createTestSession(t *testing.T, ctx context.Context, service *Service, projectID string) Session {
	t.Helper()

	session, err := service.CreateSession(ctx, CreateSessionInput{
		ProjectID:  projectID,
		Title:      "Test session",
		ActionKind: ActionKindChat,
		ApplyMode:  ApplyModePreview,
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	return session
}

func createChatTurnTestSessionWithModel(t *testing.T, ctx context.Context, service *Service, projectID, modelID string) Session {
	t.Helper()

	session, err := service.CreateSession(ctx, CreateSessionInput{
		ProjectID:      projectID,
		Title:          "Test session",
		ActionKind:     ActionKindChat,
		ModelVariantID: modelID,
		ApplyMode:      ApplyModePreview,
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	return session
}

func createTestProviderAndModel(t *testing.T, ctx context.Context, service *Service, authorID string) ModelVariant {
	t.Helper()

	provider, err := service.CreateProviderConfig(ctx, CreateProviderConfigInput{
		AuthorID: authorID,
		Name:     "Fake Provider",
		BaseURL:  "https://fake-provider.test",
		APIKey:   "secret-key",
	})
	if err != nil {
		t.Fatalf("CreateProviderConfig() error = %v", err)
	}
	model, err := service.CreateModelVariant(ctx, CreateModelVariantInput{
		AuthorID:                  authorID,
		ProviderConfigID:          provider.ID,
		Name:                      "Fake Chat",
		Model:                     "fake-chat",
		Temperature:               0.3,
		MaxOutputTokens:           1024,
		InputPricePerMillion:      1,
		OutputPricePerMillion:     2,
		CacheReadPricePerMillion:  0.1,
		CacheWritePricePerMillion: 2,
		RequestTokenField:         "max_tokens",
		CompatibilityJSON:         "{}",
	})
	if err != nil {
		t.Fatalf("CreateModelVariant() error = %v", err)
	}
	return model
}

func createPromptProfileWithRetention(t *testing.T, ctx context.Context, service *Service, projectID string, retentionDays int64) {
	t.Helper()
	_, err := service.UpsertPromptProfile(ctx, UpsertPromptProfileInput{
		ProjectID:                 projectID,
		Genre:                     "fantasy",
		Tense:                     "past tense",
		POV:                       "third person",
		Voice:                     "precise",
		InstructionsMarkdown:      "Keep continuity tight.",
		PromptRecordRetentionDays: retentionDays,
	})
	if err != nil {
		t.Fatalf("UpsertPromptProfile() error = %v", err)
	}
}

func insertPromptRecordForPrune(t *testing.T, ctx context.Context, db *sql.DB, projectID, id, createdAt string) {
	t.Helper()
	_, err := db.ExecContext(ctx, `
		INSERT INTO prompt_records (
			id, project_id, provider_name, model_name, action_kind,
			request_json, response_json, created_at
		) VALUES (?, ?, 'Fake Provider', 'fake-chat', 'chat', '{}', '{}', ?);
	`, id, projectID, createdAt)
	if err != nil {
		t.Fatalf("insert prompt record %q: %v", id, err)
	}
}

type fakeChatProvider struct {
	requests  []CompletionRequest
	responses []CompletionResponse
}

func (p *fakeChatProvider) Complete(_ context.Context, request CompletionRequest) (CompletionResponse, error) {
	p.requests = append(p.requests, request)
	if len(p.responses) == 0 {
		return CompletionResponse{}, nil
	}
	response := p.responses[0]
	p.responses = p.responses[1:]
	return response, nil
}

func assertSnapshotSource(t *testing.T, snapshots []store.PromptContextSnapshot, sourceKey string) {
	t.Helper()
	for _, snapshot := range snapshots {
		if snapshot.SourceKey == sourceKey {
			return
		}
	}
	t.Fatalf("snapshots missing source %q: %#v", sourceKey, snapshots)
}

func hasEventType(events []store.ActivityEvent, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}

func assertFloat(t *testing.T, got, want float64) {
	t.Helper()
	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	if diff > 0.000000001 {
		t.Fatalf("float = %.12f, want %.12f", got, want)
	}
}

func seedOtherAuthorModelVariant(t *testing.T, db *sql.DB) string {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO authors (id, email, password_hash, created_at)
		VALUES ('author-2', 'other@example.com', 'hash', '2026-06-13T00:00:00Z');
		INSERT INTO story_projects (id, author_id, title, slug, language, created_at, updated_at)
		VALUES ('project-author-2', 'author-2', 'Other', 'other', 'en', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
		INSERT INTO provider_configs (id, author_id, name, base_url, api_key_encrypted, created_at, updated_at)
		VALUES ('provider-author-2', 'author-2', 'Other Provider', 'http://example.test', 'encrypted', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
		INSERT INTO model_variants (
			id, provider_config_id, name, model, temperature, max_output_tokens,
			context_window_tokens, input_price_per_million, output_price_per_million,
			cache_read_price_per_million, cache_write_price_per_million,
			request_token_field, reasoning_format, compatibility_json, created_at, updated_at
		) VALUES (
			'model-author-2', 'provider-author-2', 'Other Model', 'other-model', 0.7, 2048,
			8192, 0, 0, 0, 0, 'max_tokens', '', '{}', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'
		);
	`)
	if err != nil {
		t.Fatalf("seed other author model: %v", err)
	}

	return "model-author-2"
}
