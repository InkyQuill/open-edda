package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"git.inkyquill.net/inky/writer/project"
	"git.inkyquill.net/inky/writer/store"
)

func TestContextToolDefinitionsExposeExplicitSchemas(t *testing.T) {
	tools := ContextToolDefinitions()
	names := make(map[string]CompletionTool, len(tools))
	for _, tool := range tools {
		names[tool.Function.Name] = tool
	}

	for _, name := range []string{
		"project_map",
		"search_content",
		"read_content",
		"read_chapter",
		"read_story_bible_entry",
		"read_entry_section",
		"list_revisions",
	} {
		tool, ok := names[name]
		if !ok {
			t.Fatalf("ContextToolDefinitions() missing %q", name)
		}
		if tool.Type != "function" {
			t.Fatalf("%s type = %q, want function", name, tool.Type)
		}
		if got := tool.Function.Parameters["type"]; got != "object" {
			t.Fatalf("%s schema type = %#v, want object", name, got)
		}
		if _, ok := tool.Function.Parameters["properties"].(map[string]any); !ok {
			t.Fatalf("%s schema properties missing or wrong type: %#v", name, tool.Function.Parameters["properties"])
		}
	}

	searchProps := names["search_content"].Function.Parameters["properties"].(map[string]any)
	for _, prop := range []string{"query", "kind", "metadataFilters", "tags", "limit"} {
		if _, ok := searchProps[prop]; !ok {
			t.Fatalf("search_content schema missing %q", prop)
		}
	}
}

func TestExecuteContextToolsRecordsActivityAndArtifact(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		arguments func(seed toolSeed) string
		assert    func(t *testing.T, result ToolResult)
	}{
		{
			name:     "project_map",
			toolName: "project_map",
			arguments: func(seed toolSeed) string {
				return `{}`
			},
			assert: func(t *testing.T, result ToolResult) {
				assertFullJSONContains(t, result.FullResultJSON, "content", "Opening", "Mira")
				assertMarkdownContains(t, result.ModelVisibleMarkdown, "Opening", "Mira")
			},
		},
		{
			name:     "search_content",
			toolName: "search_content",
			arguments: func(seed toolSeed) string {
				return `{"query":"alchemy","kind":"story_bible_entry","metadataFilters":{"type":"character","status":"draft"},"tags":["alchemy"],"limit":10}`
			},
			assert: func(t *testing.T, result ToolResult) {
				assertFullJSONContains(t, result.FullResultJSON, "Mira")
				assertFullJSONNotContains(t, result.FullResultJSON, "Ash Compass")
				assertFullJSONNotContains(t, result.FullResultJSON, "Sol")
			},
		},
		{
			name:     "read_content",
			toolName: "read_content",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Chapter.ID + `"}`
			},
			assert: func(t *testing.T, result ToolResult) {
				assertFullJSONContains(t, result.FullResultJSON, "The lantern burned blue.")
				assertMarkdownContains(t, result.ModelVisibleMarkdown, "The lantern burned blue.")
			},
		},
		{
			name:     "read_chapter",
			toolName: "read_chapter",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Chapter.ID + `"}`
			},
			assert: func(t *testing.T, result ToolResult) {
				assertFullJSONContains(t, result.FullResultJSON, "chapter")
				assertMarkdownContains(t, result.ModelVisibleMarkdown, "Opening")
			},
		},
		{
			name:     "read_story_bible_entry",
			toolName: "read_story_bible_entry",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Character.ID + `"}`
			},
			assert: func(t *testing.T, result ToolResult) {
				assertFullJSONContains(t, result.FullResultJSON, "story_bible_entry")
				assertMarkdownContains(t, result.ModelVisibleMarkdown, "Mira")
			},
		},
		{
			name:     "read_entry_section",
			toolName: "read_entry_section",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Character.ID + `","heading":"Motivation"}`
			},
			assert: func(t *testing.T, result ToolResult) {
				assertFullJSONContains(t, result.FullResultJSON, "Find the ash compass.")
				assertMarkdownContains(t, result.ModelVisibleMarkdown, "Find the ash compass.")
			},
		},
		{
			name:     "list_revisions",
			toolName: "list_revisions",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Chapter.ID + `"}`
			},
			assert: func(t *testing.T, result ToolResult) {
				assertFullJSONContains(t, result.FullResultJSON, "opening update", "initial draft")
				assertMarkdownContains(t, result.ModelVisibleMarkdown, "opening update")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openMigratedTestDB(t)
			ctx := context.Background()
			projectService := project.NewService(db)
			storyProject := createTestProject(t, ctx, projectService, "author-1", "Tool Project")
			service := NewService(db, projectService, nil)
			session := createTestSession(t, ctx, service, storyProject.ID)
			seed := seedToolContent(t, ctx, projectService, storyProject.ID)

			result, err := service.ExecuteTool(ctx, ToolCallInput{
				ProjectID:     storyProject.ID,
				SessionID:     session.ID,
				ToolCallID:    "call-" + tt.name,
				ToolName:      tt.toolName,
				ArgumentsJSON: tt.arguments(seed),
			})
			if err != nil {
				t.Fatalf("ExecuteTool() error = %v", err)
			}

			if result.ToolName != tt.toolName {
				t.Fatalf("ToolName = %q, want %q", result.ToolName, tt.toolName)
			}
			if result.ToolCallID != "call-"+tt.name {
				t.Fatalf("ToolCallID = %q", result.ToolCallID)
			}
			if !json.Valid([]byte(result.FullResultJSON)) {
				t.Fatalf("FullResultJSON is not valid JSON: %s", result.FullResultJSON)
			}
			if result.ModelVisibleMarkdown == "" {
				t.Fatal("ModelVisibleMarkdown is empty")
			}
			tt.assert(t, result)

			assertToolActivityAndArtifact(t, ctx, db, storyProject.ID, session.ID, tt.toolName, result)
		})
	}
}

func TestExecuteToolRejectsSessionOutsideProject(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	firstProject := createTestProject(t, ctx, projectService, "author-1", "First Tool Project")
	secondProject := createTestProject(t, ctx, projectService, "author-1", "Second Tool Project")
	service := NewService(db, projectService, nil)
	session := createTestSession(t, ctx, service, firstProject.ID)

	_, err := service.ExecuteTool(ctx, ToolCallInput{
		ProjectID:     secondProject.ID,
		SessionID:     session.ID,
		ToolCallID:    "call-wrong-project",
		ToolName:      "project_map",
		ArgumentsJSON: `{}`,
	})
	if err == nil {
		t.Fatal("ExecuteTool() error = nil, want project/session scope error")
	}
}

func TestContextToolMarkdownIncludesActionableContentIdentifiers(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Identifier Tool Project")
	service := NewService(db, projectService, nil)
	session := createTestSession(t, ctx, service, storyProject.ID)
	seed := seedToolContent(t, ctx, projectService, storyProject.ID)

	projectMap, err := service.ExecuteTool(ctx, ToolCallInput{
		ProjectID:     storyProject.ID,
		SessionID:     session.ID,
		ToolCallID:    "call-project-map-identifiers",
		ToolName:      "project_map",
		ArgumentsJSON: `{}`,
	})
	if err != nil {
		t.Fatalf("ExecuteTool(project_map) error = %v", err)
	}
	assertMarkdownContains(t, projectMap.ModelVisibleMarkdown,
		"contentId: "+seed.Chapter.ID,
		"contentId: "+seed.Character.ID,
		"section: Motivation",
	)

	search, err := service.ExecuteTool(ctx, ToolCallInput{
		ProjectID:     storyProject.ID,
		SessionID:     session.ID,
		ToolCallID:    "call-search-identifiers",
		ToolName:      "search_content",
		ArgumentsJSON: `{"query":"alchemy","kind":"story_bible_entry","metadataFilters":{"type":"character"},"limit":1}`,
	})
	if err != nil {
		t.Fatalf("ExecuteTool(search_content) error = %v", err)
	}
	assertMarkdownContains(t, search.ModelVisibleMarkdown,
		"contentId: "+seed.Character.ID,
		"Mira",
	)
}

func TestExecuteToolValidatesSearchContentArguments(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Search Validation Project")
	service := NewService(db, projectService, nil)
	session := createTestSession(t, ctx, service, storyProject.ID)

	tests := []struct {
		name      string
		arguments string
	}{
		{name: "invalid kind", arguments: `{"kind":"invalid"}`},
		{name: "zero limit", arguments: `{"limit":0}`},
		{name: "negative limit", arguments: `{"limit":-1}`},
		{name: "too large limit", arguments: `{"limit":101}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ExecuteTool(ctx, ToolCallInput{
				ProjectID:     storyProject.ID,
				SessionID:     session.ID,
				ToolCallID:    "call-" + strings.ReplaceAll(tt.name, " ", "-"),
				ToolName:      "search_content",
				ArgumentsJSON: tt.arguments,
			})
			if err == nil {
				t.Fatal("ExecuteTool() error = nil, want validation error")
			}
		})
	}
}

func TestExecuteToolValidatesReadEntrySectionArgumentsAndKind(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Section Validation Project")
	service := NewService(db, projectService, nil)
	session := createTestSession(t, ctx, service, storyProject.ID)
	seed := seedToolContent(t, ctx, projectService, storyProject.ID)

	tests := []struct {
		name      string
		arguments string
	}{
		{name: "missing content ID", arguments: `{"heading":"Motivation"}`},
		{name: "missing heading", arguments: `{"contentId":"` + seed.Character.ID + `"}`},
		{name: "wrong kind", arguments: `{"contentId":"` + seed.Chapter.ID + `","heading":"Motivation"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ExecuteTool(ctx, ToolCallInput{
				ProjectID:     storyProject.ID,
				SessionID:     session.ID,
				ToolCallID:    "call-" + strings.ReplaceAll(tt.name, " ", "-"),
				ToolName:      "read_entry_section",
				ArgumentsJSON: tt.arguments,
			})
			if err == nil {
				t.Fatal("ExecuteTool() error = nil, want validation error")
			}
		})
	}
}

func TestExecuteToolRollsBackActivityWhenArtifactInsertFails(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Rollback Tool Project")
	service := NewService(db, projectService, nil)
	session := createTestSession(t, ctx, service, storyProject.ID)
	seedToolContent(t, ctx, projectService, storyProject.ID)

	if _, err := db.Exec(`
		CREATE TRIGGER block_tool_result_artifacts
		BEFORE INSERT ON tool_result_artifacts
		BEGIN
			SELECT RAISE(ABORT, 'artifact insert blocked');
		END;
	`); err != nil {
		t.Fatalf("create artifact-blocking trigger: %v", err)
	}

	_, err := service.ExecuteTool(ctx, ToolCallInput{
		ProjectID:     storyProject.ID,
		SessionID:     session.ID,
		ToolCallID:    "call-artifact-fails",
		ToolName:      "project_map",
		ArgumentsJSON: `{}`,
	})
	if err == nil {
		t.Fatal("ExecuteTool() error = nil, want artifact failure")
	}

	events, err := store.New(db).ListActivityEvents(ctx, store.ListActivityEventsParams{
		ProjectID: storyProject.ID,
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("ListActivityEvents() error = %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("activity event count after rollback = %d, want 0", len(events))
	}
}

func TestExecuteToolBoundsModelVisibleOutput(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Large Tool Project")
	service := NewService(db, projectService, nil)
	session := createTestSession(t, ctx, service, storyProject.ID)

	body := "BEGIN\n" + strings.Repeat("middle line with enough text to overflow the output budget\n", 1400) + "END"
	chapter, err := projectService.CreateContent(ctx, project.CreateContentInput{
		ProjectID:    storyProject.ID,
		Kind:         project.KindChapter,
		Title:        "Large Chapter",
		BodyMarkdown: body,
		CreatedBy:    "author",
	})
	if err != nil {
		t.Fatalf("CreateContent() error = %v", err)
	}

	result, err := service.ExecuteTool(ctx, ToolCallInput{
		ProjectID:     storyProject.ID,
		SessionID:     session.ID,
		ToolCallID:    "call-large-read",
		ToolName:      "read_content",
		ArgumentsJSON: `{"contentId":"` + chapter.ID + `"}`,
	})
	if err != nil {
		t.Fatalf("ExecuteTool() error = %v", err)
	}

	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
	if len([]byte(result.ModelVisibleMarkdown)) > 50*1024 {
		t.Fatalf("model-visible bytes = %d, want <= 51200", len([]byte(result.ModelVisibleMarkdown)))
	}
	assertMarkdownContains(t, result.ModelVisibleMarkdown, "BEGIN", "END", "truncated")
	assertFullJSONContains(t, result.FullResultJSON, "middle line with enough text")
}

func TestExecuteToolBoundsByteHeavySingleLineDirectRead(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Single Line Project")
	service := NewService(db, projectService, nil)
	session := createTestSession(t, ctx, service, storyProject.ID)

	body := "BEGIN-" + strings.Repeat("Ж", 40_000) + "-END"
	chapter, err := projectService.CreateContent(ctx, project.CreateContentInput{
		ProjectID:    storyProject.ID,
		Kind:         project.KindChapter,
		Title:        "Single Line",
		BodyMarkdown: body,
		CreatedBy:    "author",
	})
	if err != nil {
		t.Fatalf("CreateContent() error = %v", err)
	}

	result, err := service.ExecuteTool(ctx, ToolCallInput{
		ProjectID:     storyProject.ID,
		SessionID:     session.ID,
		ToolCallID:    "call-single-line",
		ToolName:      "read_content",
		ArgumentsJSON: `{"contentId":"` + chapter.ID + `"}`,
	})
	if err != nil {
		t.Fatalf("ExecuteTool() error = %v", err)
	}

	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
	if len([]byte(result.ModelVisibleMarkdown)) > 50*1024 {
		t.Fatalf("model-visible bytes = %d, want <= 51200", len([]byte(result.ModelVisibleMarkdown)))
	}
	if !json.Valid([]byte(result.FullResultJSON)) {
		t.Fatal("full result JSON is invalid")
	}
	assertMarkdownContains(t, result.ModelVisibleMarkdown, "BEGIN-", "-END", "truncated")
}

func TestExecuteToolBoundsByteHeavySingleLineListResult(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Single Line List Project")
	service := NewService(db, projectService, nil)
	session := createTestSession(t, ctx, service, storyProject.ID)

	for _, input := range []project.CreateContentInput{
		{
			ProjectID:    storyProject.ID,
			Kind:         project.KindProjectNote,
			Title:        "BEGIN item",
			BodyMarkdown: "needle " + strings.Repeat("Ж", 35_000),
			CreatedBy:    "author",
		},
		{
			ProjectID:    storyProject.ID,
			Kind:         project.KindProjectNote,
			Title:        "END item",
			BodyMarkdown: "needle " + strings.Repeat("Я", 35_000),
			CreatedBy:    "author",
		},
	} {
		if _, err := projectService.CreateContent(ctx, input); err != nil {
			t.Fatalf("CreateContent(%q) error = %v", input.Title, err)
		}
	}

	result, err := service.ExecuteTool(ctx, ToolCallInput{
		ProjectID:     storyProject.ID,
		SessionID:     session.ID,
		ToolCallID:    "call-list-single-line",
		ToolName:      "search_content",
		ArgumentsJSON: `{"query":"needle","kind":"project_note","limit":2}`,
	})
	if err != nil {
		t.Fatalf("ExecuteTool() error = %v", err)
	}

	if !result.Truncated {
		t.Fatal("Truncated = false, want true")
	}
	if len([]byte(result.ModelVisibleMarkdown)) > 50*1024 {
		t.Fatalf("model-visible bytes = %d, want <= 51200", len([]byte(result.ModelVisibleMarkdown)))
	}
	assertMarkdownContains(t, result.ModelVisibleMarkdown, "BEGIN item", "END item", "truncated")
}

type toolSeed struct {
	Chapter   project.ContentItem
	Character project.ContentItem
}

func seedToolContent(t *testing.T, ctx context.Context, service *project.Service, projectID string) toolSeed {
	t.Helper()

	chapter, err := service.CreateContent(ctx, project.CreateContentInput{
		ProjectID:    projectID,
		Kind:         project.KindChapter,
		Title:        "Opening",
		BodyMarkdown: "The lantern burned blue.",
		MetadataJSON: `{"status":"draft"}`,
		Reason:       "initial draft",
		CreatedBy:    "author",
	})
	if err != nil {
		t.Fatalf("CreateContent(chapter) error = %v", err)
	}
	if _, err := service.UpdateContent(ctx, project.UpdateContentInput{
		ProjectID:        projectID,
		ContentID:        chapter.ID,
		ExpectedRevision: 1,
		BodyMarkdown:     "The lantern burned blue.\nMira listened.",
		MetadataJSON:     `{"status":"draft"}`,
		Reason:           "opening update",
		CreatedBy:        "author",
	}); err != nil {
		t.Fatalf("UpdateContent(chapter) error = %v", err)
	}

	character, err := service.CreateContent(ctx, project.CreateContentInput{
		ProjectID:    projectID,
		Kind:         project.KindStoryBibleEntry,
		Title:        "Mira",
		BodyMarkdown: "Alchemy notes mention the lantern.",
		MetadataJSON: `{"type":"character","status":"draft","tags":["alchemy"]}`,
		CreatedBy:    "author",
	})
	if err != nil {
		t.Fatalf("CreateContent(character) error = %v", err)
	}
	if err := service.CreateEntrySection(ctx, project.CreateEntrySectionInput{
		ProjectID:     projectID,
		ContentItemID: character.ID,
		Heading:       "Motivation",
		BodyMarkdown:  "Find the ash compass.",
		SortOrder:     1,
	}); err != nil {
		t.Fatalf("CreateEntrySection() error = %v", err)
	}
	if err := service.CreateEntryRelation(ctx, project.CreateEntryRelationInput{
		ProjectID:    projectID,
		SourceItemID: character.ID,
		TargetTitle:  "Ash Compass",
		RelationType: "seeks",
	}); err != nil {
		t.Fatalf("CreateEntryRelation() error = %v", err)
	}
	for _, input := range []project.CreateContentInput{
		{
			ProjectID:    projectID,
			Kind:         project.KindStoryBibleEntry,
			Title:        "Ash Compass",
			BodyMarkdown: "Alchemy notes mention the lantern.",
			MetadataJSON: `{"type":"artifact","status":"draft","tags":["alchemy"]}`,
			CreatedBy:    "author",
		},
		{
			ProjectID:    projectID,
			Kind:         project.KindStoryBibleEntry,
			Title:        "Sol",
			BodyMarkdown: "Alchemy notes mention the lantern.",
			MetadataJSON: `{"type":"character","status":"final","tags":["alchemy"]}`,
			CreatedBy:    "author",
		},
	} {
		if _, err := service.CreateContent(ctx, input); err != nil {
			t.Fatalf("CreateContent(%q) error = %v", input.Title, err)
		}
	}

	return toolSeed{Chapter: chapter, Character: character}
}

func assertToolActivityAndArtifact(t *testing.T, ctx context.Context, db *sql.DB, projectID, sessionID, toolName string, result ToolResult) {
	t.Helper()

	queries := store.New(db)
	events, err := queries.ListActivityEvents(ctx, store.ListActivityEventsParams{
		ProjectID: projectID,
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("ListActivityEvents() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("activity event count = %d, want 1", len(events))
	}
	event := events[0]
	if event.EventType == "" {
		t.Fatal("activity event type is empty")
	}
	if !event.SessionID.Valid || event.SessionID.String != sessionID {
		t.Fatalf("activity session = %#v, want %q", event.SessionID, sessionID)
	}
	if !strings.Contains(event.Summary, toolName) {
		t.Fatalf("activity summary = %q, want tool name %q", event.Summary, toolName)
	}
	var metadata map[string]any
	if err := json.Unmarshal([]byte(event.MetadataJson), &metadata); err != nil {
		t.Fatalf("unmarshal activity metadata: %v", err)
	}
	if metadata["toolName"] != toolName {
		t.Fatalf("activity metadata toolName = %#v, want %q", metadata["toolName"], toolName)
	}
	if metadata["toolCallId"] != result.ToolCallID {
		t.Fatalf("activity metadata toolCallId = %#v, want %q", metadata["toolCallId"], result.ToolCallID)
	}

	artifacts, err := queries.ListToolResultArtifacts(ctx, store.ListToolResultArtifactsParams{
		ProjectID: projectID,
		SessionID: sql.NullString{String: sessionID, Valid: true},
	})
	if err != nil {
		t.Fatalf("ListToolResultArtifacts() error = %v", err)
	}
	if len(artifacts) != 1 {
		t.Fatalf("artifact count = %d, want 1", len(artifacts))
	}
	artifact := artifacts[0]
	if artifact.ToolName != toolName {
		t.Fatalf("artifact tool = %q, want %q", artifact.ToolName, toolName)
	}
	if artifact.ToolCallID != result.ToolCallID {
		t.Fatalf("artifact tool call ID = %q, want %q", artifact.ToolCallID, result.ToolCallID)
	}
	if artifact.FullResultJson != result.FullResultJSON {
		t.Fatal("artifact full JSON differs from returned full JSON")
	}
	if artifact.ModelVisibleMarkdown != result.ModelVisibleMarkdown {
		t.Fatal("artifact model-visible markdown differs from returned markdown")
	}
	if (artifact.Truncated != 0) != result.Truncated {
		t.Fatalf("artifact truncated = %d, result truncated = %v", artifact.Truncated, result.Truncated)
	}
	if artifact.FullResultBytes != int64(len([]byte(result.FullResultJSON))) {
		t.Fatalf("artifact full bytes = %d, want %d", artifact.FullResultBytes, len([]byte(result.FullResultJSON)))
	}
}

func assertFullJSONContains(t *testing.T, value string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(value, needle) {
			t.Fatalf("full JSON missing %q: %.500s", needle, value)
		}
	}
}

func assertFullJSONNotContains(t *testing.T, value string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			t.Fatalf("full JSON contains %q unexpectedly: %.500s", needle, value)
		}
	}
}

func assertMarkdownContains(t *testing.T, value string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(value, needle) {
			t.Fatalf("markdown missing %q: %.500s", needle, value)
		}
	}
}
