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
