package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"git.inkyquill.net/inky/writer/project"
	"git.inkyquill.net/inky/writer/skill"
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
		"skill",
		"append_to_chapter",
		"insert_into_chapter",
		"replace_selection",
		"update_story_bible_entry",
		"update_entry_section",
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

	skillSchema := names["skill"].Function.Parameters
	required, ok := skillSchema["required"].([]string)
	if !ok {
		t.Fatalf("skill schema required = %#v, want []string", skillSchema["required"])
	}
	if len(required) != 1 || required[0] != "skillId" {
		t.Fatalf("skill schema required = %#v, want [skillId]", required)
	}
	skillProps := skillSchema["properties"].(map[string]any)
	if _, ok := skillProps["skillId"]; !ok {
		t.Fatal("skill schema missing skillId property")
	}
}

func TestExecuteWriteToolsRecordRevisionActivityAndArtifact(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		arguments    func(seed toolSeed) string
		wantContent  func(seed toolSeed) (contentID string, body string, revision int64)
		wantMetadata map[string]any
	}{
		{
			name:     "append_to_chapter",
			toolName: "append_to_chapter",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Chapter.ID + `","expectedRevision":2,"generatedMarkdown":"\nThe glass city answered.","reason":"continue scene"}`
			},
			wantContent: func(seed toolSeed) (string, string, int64) {
				return seed.Chapter.ID, "The lantern burned blue.\nMira listened.\nThe glass city answered.", 3
			},
			wantMetadata: map[string]any{
				"operationKind":    "append_to_chapter",
				"previousRevision": float64(2),
				"nextRevision":     float64(3),
			},
		},
		{
			name:     "insert_into_chapter",
			toolName: "insert_into_chapter",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Chapter.ID + `","expectedRevision":2,"insertPosition":11,"generatedMarkdown":" quietly","reason":"insert adverb"}`
			},
			wantContent: func(seed toolSeed) (string, string, int64) {
				return seed.Chapter.ID, "The lantern quietly burned blue.\nMira listened.", 3
			},
			wantMetadata: map[string]any{
				"operationKind":    "insert_into_chapter",
				"previousRevision": float64(2),
				"nextRevision":     float64(3),
			},
		},
		{
			name:     "replace_selection",
			toolName: "replace_selection",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Chapter.ID + `","expectedRevision":2,"selectionStart":12,"selectionEnd":23,"generatedMarkdown":"glowed green","reason":"replace image"}`
			},
			wantContent: func(seed toolSeed) (string, string, int64) {
				return seed.Chapter.ID, "The lantern glowed green.\nMira listened.", 3
			},
			wantMetadata: map[string]any{
				"operationKind":    "replace_selection",
				"previousRevision": float64(2),
				"nextRevision":     float64(3),
			},
		},
		{
			name:     "update_story_bible_entry",
			toolName: "update_story_bible_entry",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Character.ID + `","expectedRevision":1,"generatedMarkdown":"Mira keeps precise notes about impossible lanterns.","reason":"update character entry"}`
			},
			wantContent: func(seed toolSeed) (string, string, int64) {
				return seed.Character.ID, "Mira keeps precise notes about impossible lanterns.", 2
			},
			wantMetadata: map[string]any{
				"operationKind":    "update_story_bible_entry",
				"previousRevision": float64(1),
				"nextRevision":     float64(2),
			},
		},
		{
			name:     "update_entry_section",
			toolName: "update_entry_section",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Character.ID + `","heading":"Motivation","generatedMarkdown":"Protect the glass city.","reason":"update motivation"}`
			},
			wantContent: func(seed toolSeed) (string, string, int64) {
				return seed.Character.ID, "Alchemy notes mention the lantern.", 1
			},
			wantMetadata: map[string]any{
				"operationKind": "update_entry_section",
				"heading":       "Motivation",
				"reason":        "update motivation",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openMigratedTestDB(t)
			ctx := context.Background()
			projectService := project.NewService(db)
			storyProject := createTestProject(t, ctx, projectService, "author-1", "Write Tool Project")
			service := NewService(db, projectService, nil)
			modelID := seedToolModelVariant(t, db, "author-1")
			session := createTestSessionWithModel(t, ctx, service, storyProject.ID, ActionKindRewrite, modelID)
			seed := seedToolContent(t, ctx, projectService, storyProject.ID)

			result, err := service.ExecuteTool(ctx, ToolCallInput{
				ProjectID:     storyProject.ID,
				SessionID:     session.ID,
				ToolCallID:    "call-" + tt.name,
				ToolName:      tt.toolName,
				ArgumentsJSON: tt.arguments(seed),
			})
			if err != nil {
				t.Fatalf("ExecuteTool(%s) error = %v", tt.toolName, err)
			}

			contentID, wantBody, wantRevision := tt.wantContent(seed)
			item, err := projectService.GetContent(ctx, storyProject.ID, contentID)
			if err != nil {
				t.Fatalf("GetContent() error = %v", err)
			}
			if item.BodyMarkdown != wantBody {
				t.Fatalf("body = %q, want %q", item.BodyMarkdown, wantBody)
			}
			if item.CurrentRevision != wantRevision {
				t.Fatalf("revision = %d, want %d", item.CurrentRevision, wantRevision)
			}
			if tt.toolName == "update_entry_section" {
				sections, err := projectService.ListEntrySections(ctx, storyProject.ID, seed.Character.ID)
				if err != nil {
					t.Fatalf("ListEntrySections() error = %v", err)
				}
				if len(sections) != 1 || sections[0].BodyMarkdown != "Protect the glass city." {
					t.Fatalf("sections = %#v", sections)
				}
			}

			assertWriteToolActivityAndArtifact(t, ctx, db, storyProject.ID, session, tt.toolName, contentID, result, tt.wantMetadata)
		})
	}
}

func TestExecuteWriteToolRequiresSession(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Write Session Project")
	service := NewService(db, projectService, nil)
	seed := seedToolContent(t, ctx, projectService, storyProject.ID)

	_, err := service.ExecuteTool(ctx, ToolCallInput{
		ProjectID:     storyProject.ID,
		ToolCallID:    "call-no-session",
		ToolName:      "append_to_chapter",
		ArgumentsJSON: `{"contentId":"` + seed.Chapter.ID + `","expectedRevision":2,"generatedMarkdown":"\nNo session.","reason":"append without session"}`,
	})
	if err == nil {
		t.Fatal("ExecuteTool(write without session) error = nil, want validation error")
	}
}

func TestExecuteWriteToolValidatesTargetKind(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Write Kind Project")
	service := NewService(db, projectService, nil)
	session := createTestSession(t, ctx, service, storyProject.ID)
	seed := seedToolContent(t, ctx, projectService, storyProject.ID)

	tests := []struct {
		name      string
		toolName  string
		arguments string
	}{
		{
			name:      "append rejects story bible entry",
			toolName:  "append_to_chapter",
			arguments: `{"contentId":"` + seed.Character.ID + `","expectedRevision":1,"generatedMarkdown":"\nWrong kind.","reason":"bad append"}`,
		},
		{
			name:      "entry update rejects chapter",
			toolName:  "update_story_bible_entry",
			arguments: `{"contentId":"` + seed.Chapter.ID + `","expectedRevision":2,"generatedMarkdown":"Wrong kind.","reason":"bad entry update"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ExecuteTool(ctx, ToolCallInput{
				ProjectID:     storyProject.ID,
				SessionID:     session.ID,
				ToolCallID:    "call-" + strings.ReplaceAll(tt.name, " ", "-"),
				ToolName:      tt.toolName,
				ArgumentsJSON: tt.arguments,
			})
			if err == nil {
				t.Fatal("ExecuteTool() error = nil, want target kind validation error")
			}
		})
	}
}

func TestExecuteWriteToolRejectsBlankGeneratedMarkdown(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		arguments func(seed toolSeed) string
	}{
		{
			name:     "append_to_chapter",
			toolName: "append_to_chapter",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Chapter.ID + `","expectedRevision":2,"generatedMarkdown":"   ","reason":"blank append"}`
			},
		},
		{
			name:     "insert_into_chapter",
			toolName: "insert_into_chapter",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Chapter.ID + `","expectedRevision":2,"insertPosition":3,"generatedMarkdown":"","reason":"blank insert"}`
			},
		},
		{
			name:     "replace_selection",
			toolName: "replace_selection",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Chapter.ID + `","expectedRevision":2,"selectionStart":0,"selectionEnd":3,"generatedMarkdown":"\n\t ","reason":"blank replace"}`
			},
		},
		{
			name:     "update_story_bible_entry",
			toolName: "update_story_bible_entry",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Character.ID + `","expectedRevision":1,"generatedMarkdown":" ","reason":"blank entry update"}`
			},
		},
		{
			name:     "update_entry_section",
			toolName: "update_entry_section",
			arguments: func(seed toolSeed) string {
				return `{"contentId":"` + seed.Character.ID + `","heading":"Motivation","generatedMarkdown":"  ","reason":"blank section update"}`
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openMigratedTestDB(t)
			ctx := context.Background()
			projectService := project.NewService(db)
			storyProject := createTestProject(t, ctx, projectService, "author-1", "Blank Write Project")
			service := NewService(db, projectService, nil)
			session := createTestSession(t, ctx, service, storyProject.ID)
			seed := seedToolContent(t, ctx, projectService, storyProject.ID)

			_, err := service.ExecuteTool(ctx, ToolCallInput{
				ProjectID:     storyProject.ID,
				SessionID:     session.ID,
				ToolCallID:    "call-blank-" + tt.name,
				ToolName:      tt.toolName,
				ArgumentsJSON: tt.arguments(seed),
			})
			if err == nil {
				t.Fatal("ExecuteTool() error = nil, want blank generatedMarkdown validation error")
			}
			if !strings.Contains(err.Error(), "generatedMarkdown is required") {
				t.Fatalf("ExecuteTool() error = %v, want generatedMarkdown validation error", err)
			}

			assertNoToolActivityOrArtifact(t, ctx, db, storyProject.ID, session.ID)
		})
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

func TestExecuteSkillToolReturnsRenderedSkillAndRecordsActivityAndArtifact(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	skillService := skill.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Skill Tool Project")
	service := NewService(db, projectService, nil)
	service.SetSkillService(skillService)
	session := createTestSession(t, ctx, service, storyProject.ID)

	installed := installToolSkill(t, ctx, skillService, storyProject.ID)
	result, err := service.ExecuteTool(ctx, ToolCallInput{
		ProjectID:     storyProject.ID,
		SessionID:     session.ID,
		ToolCallID:    "call-skill",
		ToolName:      "skill",
		ArgumentsJSON: `{"skillId":"` + installed.ID + `"}`,
	})
	if err != nil {
		t.Fatalf("ExecuteTool(skill) error = %v", err)
	}

	assertMarkdownContains(t, result.ModelVisibleMarkdown,
		`<skill_content name="style-pass">`,
		`Always tighten prose.`,
		`<file path="templates/rewrite.md" purpose="template">Rewrite template</file>`,
	)
	assertFullJSONContains(t, result.FullResultJSON,
		`"modelVisibleMarkdown"`,
		`"skill"`,
		`"instructionsMarkdown": "Always tighten prose."`,
		`"bodyText": "Rewrite template"`,
	)
	assertSkillToolActivityAndArtifact(t, ctx, db, storyProject.ID, session.ID, result, installed.ID)
}

func TestExecuteSkillToolReturnsErrorForMissingSkill(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	skillService := skill.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Missing Skill Tool Project")
	service := NewService(db, projectService, nil)
	service.SetSkillService(skillService)
	session := createTestSession(t, ctx, service, storyProject.ID)

	_, err := service.ExecuteTool(ctx, ToolCallInput{
		ProjectID:     storyProject.ID,
		SessionID:     session.ID,
		ToolCallID:    "call-missing-skill",
		ToolName:      "skill",
		ArgumentsJSON: `{"skillId":"skill-missing"}`,
	})
	if err == nil {
		t.Fatal("ExecuteTool(skill missing) error = nil, want error")
	}

	assertNoToolActivityOrArtifact(t, ctx, db, storyProject.ID, session.ID)
}

func TestExecuteSkillToolScriptFilesAreDisabledAndNotExecuted(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	skillService := skill.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Script Skill Tool Project")
	service := NewService(db, projectService, nil)
	service.SetSkillService(skillService)
	session := createTestSession(t, ctx, service, storyProject.ID)

	installed := installToolSkill(t, ctx, skillService, storyProject.ID)
	result, err := service.ExecuteTool(ctx, ToolCallInput{
		ProjectID:     storyProject.ID,
		SessionID:     session.ID,
		ToolCallID:    "call-script-skill",
		ToolName:      "skill",
		ArgumentsJSON: `{"skillId":"` + installed.ID + `"}`,
	})
	if err != nil {
		t.Fatalf("ExecuteTool(skill) error = %v", err)
	}

	assertMarkdownContains(t, result.ModelVisibleMarkdown, `disabled="true"`)
	if strings.Contains(result.ModelVisibleMarkdown, "echo should-not-execute") {
		t.Fatalf("ModelVisibleMarkdown included shell execution output:\n%s", result.ModelVisibleMarkdown)
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

func installToolSkill(t *testing.T, ctx context.Context, service *skill.Service, projectID string) skill.Skill {
	t.Helper()

	installed, err := service.Install(ctx, skill.InstallInput{
		ProjectID:   projectID,
		SourceType:  skill.SourceTypeUpload,
		SourceLabel: "style-pass.zip",
		Imported: skill.ImportedSkill{
			Name:                 "style-pass",
			DisplayName:          "Style Pass",
			Description:          "Tightens style",
			InstructionsMarkdown: "Always tighten prose.",
			MetadataJSON:         "{}",
			SourceLabel:          "style-pass.zip",
			ScriptCount:          1,
			ScriptsDisabled:      true,
			Files: []skill.ImportedSkillFile{
				{
					RelativePath: "templates/rewrite.md",
					Purpose:      skill.FilePurposeTemplate,
					MediaType:    "text/markdown; charset=utf-8",
					BodyText:     "Rewrite template",
					Bytes:        int64(len("Rewrite template")),
				},
				{
					RelativePath:   "scripts/analyze.sh",
					Purpose:        skill.FilePurposeScript,
					MediaType:      "text/x-shellscript; charset=utf-8",
					BodyText:       "echo should-not-execute",
					Bytes:          int64(len("echo should-not-execute")),
					ScriptDisabled: true,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Install(skill) error = %v", err)
	}
	return installed
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

func assertSkillToolActivityAndArtifact(t *testing.T, ctx context.Context, db *sql.DB, projectID, sessionID string, result ToolResult, skillID string) {
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
	if event.EventType != "skill_loaded" {
		t.Fatalf("activity event type = %q, want skill_loaded", event.EventType)
	}
	if !event.SessionID.Valid || event.SessionID.String != sessionID {
		t.Fatalf("activity session = %#v, want %q", event.SessionID, sessionID)
	}
	if event.Summary != "Loaded skill style-pass" {
		t.Fatalf("activity summary = %q, want Loaded skill style-pass", event.Summary)
	}
	var metadata map[string]any
	if err := json.Unmarshal([]byte(event.MetadataJson), &metadata); err != nil {
		t.Fatalf("unmarshal activity metadata: %v", err)
	}
	for key, want := range map[string]any{
		"toolName":        "skill",
		"toolCallId":      result.ToolCallID,
		"skillId":         skillID,
		"name":            "style-pass",
		"scriptCount":     float64(1),
		"scriptsDisabled": true,
	} {
		if metadata[key] != want {
			t.Fatalf("activity metadata %s = %#v, want %#v", key, metadata[key], want)
		}
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
	if artifact.ToolName != "skill" {
		t.Fatalf("artifact tool = %q, want skill", artifact.ToolName)
	}
	if artifact.FullResultJson != result.FullResultJSON {
		t.Fatal("artifact full JSON differs from returned full JSON")
	}
	if artifact.ModelVisibleMarkdown != result.ModelVisibleMarkdown {
		t.Fatal("artifact model-visible markdown differs from returned markdown")
	}
	if artifact.Truncated != 0 || result.Truncated {
		t.Fatalf("truncated = artifact %d result %v, want false", artifact.Truncated, result.Truncated)
	}
	if len([]byte(artifact.ModelVisibleMarkdown)) > maxToolVisibleBytes {
		t.Fatalf("artifact model-visible bytes = %d, want <= %d", len([]byte(artifact.ModelVisibleMarkdown)), maxToolVisibleBytes)
	}
	assertFullJSONContains(t, artifact.FullResultJson, `"scriptCount": 1`, `"scriptsDisabled": true`)
	assertMarkdownContains(t, artifact.ModelVisibleMarkdown, `<skill_content name="style-pass">`)
}

func assertWriteToolActivityAndArtifact(t *testing.T, ctx context.Context, db *sql.DB, projectID string, session Session, toolName string, targetContentID string, result ToolResult, want map[string]any) {
	t.Helper()

	queries := store.New(db)
	events, err := queries.ListActivityEvents(ctx, store.ListActivityEventsParams{
		ProjectID: projectID,
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("ListActivityEvents() error = %v", err)
	}
	var event store.ActivityEvent
	for _, candidate := range events {
		var metadata map[string]any
		if err := json.Unmarshal([]byte(candidate.MetadataJson), &metadata); err != nil {
			t.Fatalf("unmarshal activity metadata: %v", err)
		}
		if metadata["toolName"] == toolName && metadata["toolCallId"] == result.ToolCallID {
			event = candidate
			break
		}
	}
	if event.ID == "" {
		t.Fatalf("write tool activity for %s not found in %#v", toolName, events)
	}
	if !event.SessionID.Valid || event.SessionID.String != session.ID {
		t.Fatalf("activity session = %#v, want %q", event.SessionID, session.ID)
	}

	var metadata map[string]any
	if err := json.Unmarshal([]byte(event.MetadataJson), &metadata); err != nil {
		t.Fatalf("unmarshal selected activity metadata: %v", err)
	}
	if metadata["targetContentId"] != targetContentID {
		t.Fatalf("targetContentId = %#v, want %q", metadata["targetContentId"], targetContentID)
	}
	if metadata["actionKind"] != string(session.ActionKind) {
		t.Fatalf("actionKind = %#v, want %q", metadata["actionKind"], session.ActionKind)
	}
	if metadata["modelVariantId"] != session.ModelVariantID {
		t.Fatalf("modelVariantId = %#v, want %q", metadata["modelVariantId"], session.ModelVariantID)
	}
	if metadata["agentSessionId"] != session.ID {
		t.Fatalf("agentSessionId = %#v, want %q", metadata["agentSessionId"], session.ID)
	}
	for key, value := range want {
		if metadata[key] != value {
			t.Fatalf("metadata[%s] = %#v, want %#v", key, metadata[key], value)
		}
	}

	artifacts, err := queries.ListToolResultArtifacts(ctx, store.ListToolResultArtifactsParams{
		ProjectID: projectID,
		SessionID: sql.NullString{String: session.ID, Valid: true},
	})
	if err != nil {
		t.Fatalf("ListToolResultArtifacts() error = %v", err)
	}
	var artifact store.ToolResultArtifact
	for _, candidate := range artifacts {
		if candidate.ToolName == toolName && candidate.ToolCallID == result.ToolCallID {
			artifact = candidate
			break
		}
	}
	if artifact.ID == "" {
		t.Fatalf("artifact for %s not found in %#v", toolName, artifacts)
	}
	if artifact.FullResultJson != result.FullResultJSON {
		t.Fatal("artifact full JSON differs from returned full JSON")
	}
}

func assertNoToolActivityOrArtifact(t *testing.T, ctx context.Context, db *sql.DB, projectID, sessionID string) {
	t.Helper()

	queries := store.New(db)
	events, err := queries.ListActivityEvents(ctx, store.ListActivityEventsParams{
		ProjectID: projectID,
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("ListActivityEvents() error = %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("activity event count = %d, want 0", len(events))
	}

	artifacts, err := queries.ListToolResultArtifacts(ctx, store.ListToolResultArtifactsParams{
		ProjectID: projectID,
		SessionID: sql.NullString{String: sessionID, Valid: true},
	})
	if err != nil {
		t.Fatalf("ListToolResultArtifacts() error = %v", err)
	}
	if len(artifacts) != 0 {
		t.Fatalf("artifact count = %d, want 0", len(artifacts))
	}
}

func seedToolModelVariant(t *testing.T, db *sql.DB, authorID string) string {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO provider_configs (id, author_id, name, base_url, api_key_encrypted, created_at, updated_at)
		VALUES ('tool-provider-1', ?, 'Tool Provider', 'http://example.test', 'encrypted', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
		INSERT INTO model_variants (
			id, provider_config_id, name, model, temperature, max_output_tokens,
			context_window_tokens, input_price_per_million, output_price_per_million,
			cache_read_price_per_million, cache_write_price_per_million,
			request_token_field, reasoning_format, compatibility_json, created_at, updated_at
		) VALUES (
			'tool-model-1', 'tool-provider-1', 'Tool Model', 'tool-model', 0.7, 2048,
			8192, 0, 0, 0, 0, 'max_tokens', '', '{}', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'
		);
	`, authorID)
	if err != nil {
		t.Fatalf("seed tool model variant: %v", err)
	}
	return "tool-model-1"
}

func createTestSessionWithModel(t *testing.T, ctx context.Context, service *Service, projectID string, actionKind ActionKind, modelID string) Session {
	t.Helper()

	session, err := service.CreateSession(ctx, CreateSessionInput{
		ProjectID:      projectID,
		Title:          "Write session",
		ActionKind:     actionKind,
		ModelVariantID: modelID,
		ApplyMode:      ApplyModeDirectApply,
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	return session
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
