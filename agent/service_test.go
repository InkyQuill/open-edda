package agent

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"git.inkyquill.net/inky/writer/project"
	"git.inkyquill.net/inky/writer/store"
	"github.com/pressly/goose/v3"
)

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
