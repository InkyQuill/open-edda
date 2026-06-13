package agent

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"git.inkyquill.net/inky/writer/project"
)

func TestPromptProfileStoresAndRetrievesProjectPreferences(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Prompt Profile Project")
	service := NewService(db, projectService, nil)

	saved, err := service.UpsertPromptProfile(ctx, UpsertPromptProfileInput{
		ProjectID:                 storyProject.ID,
		Genre:                     "secondary-world fantasy",
		Tense:                     "past tense",
		POV:                       "close third person",
		Voice:                     "lyrical but restrained",
		InstructionsMarkdown:      "Favor sensory continuity and keep dialogue subtextual.",
		PromptRecordRetentionDays: 45,
	})
	if err != nil {
		t.Fatalf("UpsertPromptProfile() error = %v", err)
	}

	got, err := service.GetPromptProfile(ctx, storyProject.ID)
	if err != nil {
		t.Fatalf("GetPromptProfile() error = %v", err)
	}

	if got.ID == "" {
		t.Fatal("profile ID is empty")
	}
	if got.ID != saved.ID {
		t.Fatalf("retrieved profile ID = %q, want saved ID %q", got.ID, saved.ID)
	}
	if got.ProjectID != storyProject.ID {
		t.Fatalf("project ID = %q, want %q", got.ProjectID, storyProject.ID)
	}
	if got.Genre != "secondary-world fantasy" {
		t.Fatalf("genre = %q", got.Genre)
	}
	if got.Tense != "past tense" {
		t.Fatalf("tense = %q", got.Tense)
	}
	if got.POV != "close third person" {
		t.Fatalf("POV = %q", got.POV)
	}
	if got.Voice != "lyrical but restrained" {
		t.Fatalf("voice = %q", got.Voice)
	}
	if got.InstructionsMarkdown != "Favor sensory continuity and keep dialogue subtextual." {
		t.Fatalf("instructions = %q", got.InstructionsMarkdown)
	}
	if got.PromptRecordRetentionDays != 45 {
		t.Fatalf("retention days = %d, want 45", got.PromptRecordRetentionDays)
	}
}

func TestPromptBuildActionPromptAssemblesContinuationContext(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Prompt Assembly Project")
	service := NewService(db, projectService, nil)

	_, err := service.UpsertPromptProfile(ctx, UpsertPromptProfileInput{
		ProjectID:                 storyProject.ID,
		Genre:                     "solarpunk mystery",
		Tense:                     "present tense",
		POV:                       "first person",
		Voice:                     "wry, observant, intimate",
		InstructionsMarkdown:      "Keep clues fair and avoid naming the culprit early.",
		PromptRecordRetentionDays: 30,
	})
	if err != nil {
		t.Fatalf("UpsertPromptProfile() error = %v", err)
	}

	chapter := createPromptContent(t, ctx, projectService, project.CreateContentInput{
		ProjectID:    storyProject.ID,
		Kind:         project.KindChapter,
		Title:        "Chapter 3: The Glass Atrium",
		BodyMarkdown: "Mira stops at the broken conservatory doors.\n\nThe rain smells like copper.",
		SortOrder:    20,
	})
	createPromptContent(t, ctx, projectService, project.CreateContentInput{
		ProjectID:    storyProject.ID,
		Kind:         project.KindWritingBrief,
		Title:        "Project Synopsis",
		BodyMarkdown: "Project-wide synopsis: Mira investigates a city where gardens remember crimes.",
		MetadataJSON: `{}`,
		SortOrder:    1,
	})
	createPromptContent(t, ctx, projectService, project.CreateContentInput{
		ProjectID:    storyProject.ID,
		Kind:         project.KindWritingBrief,
		Title:        "Chapter 3 Brief",
		BodyMarkdown: "Chapter-specific brief: this scene should reveal the gardener's missing key.",
		MetadataJSON: `{"contentItemId":"` + chapter.ID + `","scope":"chapter"}`,
		SortOrder:    2,
	})
	createPromptContent(t, ctx, projectService, project.CreateContentInput{
		ProjectID:    storyProject.ID,
		Kind:         project.KindStoryBibleEntry,
		Title:        "Mira Vale",
		BodyMarkdown: "Detective with a botanist's memory.",
		SortOrder:    30,
	})

	bundle, err := service.BuildActionPrompt(ctx, BuildPromptInput{
		ProjectID:       storyProject.ID,
		ActionKind:      ActionKindContinuation,
		TargetContentID: chapter.ID,
		CursorSummary:   "Cursor is after the copper-rain sentence.",
		UserGuidance:    "Continue with Mira noticing one impossible flower.",
	})
	if err != nil {
		t.Fatalf("BuildActionPrompt() error = %v", err)
	}

	assertContains(t, bundle.SystemMessage, "You are a fiction writing assistant working inside Writer. Preserve the author's intent, respect established project facts, and use available tools to inspect project context before making claims. Do not invent durable worldbuilding facts unless the author asks you to brainstorm.")
	assertContains(t, bundle.UserMessage, "Continue the target chapter")
	assertContains(t, bundle.UserMessage, "Continue with Mira noticing one impossible flower.")

	contextMarkdown := bundle.DeveloperMessage
	assertContains(t, contextMarkdown, "Genre: solarpunk mystery")
	assertContains(t, contextMarkdown, "Tense: present tense")
	assertContains(t, contextMarkdown, "Point of view: first person")
	assertContains(t, contextMarkdown, "Voice: wry, observant, intimate")
	assertContains(t, contextMarkdown, "Project-wide synopsis: Mira investigates a city where gardens remember crimes.")
	assertContains(t, contextMarkdown, "Chapter-specific brief: this scene should reveal the gardener's missing key.")
	assertOrdered(t, contextMarkdown, "Project-wide synopsis", "Chapter-specific brief")
	assertContains(t, contextMarkdown, "Target chapter: Chapter 3: The Glass Atrium")
	assertContains(t, contextMarkdown, "Cursor is after the copper-rain sentence.")
	assertContains(t, contextMarkdown, "Use tools for additional project context instead of assuming the whole project is present in this prompt.")

	if !json.Valid([]byte(bundle.MetadataJSON)) {
		t.Fatalf("MetadataJSON is not valid JSON: %q", bundle.MetadataJSON)
	}

	sourceKeys := make([]string, 0, len(bundle.ContextSources))
	for _, source := range bundle.ContextSources {
		sourceKeys = append(sourceKeys, source.SourceKey)
		if source.RenderedMarkdown == "" {
			t.Fatalf("context source %q has empty rendered markdown", source.SourceKey)
		}
		if !json.Valid([]byte(source.ValueJSON)) {
			t.Fatalf("context source %q has invalid value JSON: %q", source.SourceKey, source.ValueJSON)
		}
	}
	for _, key := range []string{"prompt_profile", "writing_briefs", "target_content_window", "provider_disclosure", "tool_catalog"} {
		if !slices.Contains(sourceKeys, key) {
			t.Fatalf("context source keys = %v, want key %q", sourceKeys, key)
		}
	}
}

func createPromptContent(t *testing.T, ctx context.Context, service *project.Service, input project.CreateContentInput) project.ContentItem {
	t.Helper()

	input.CreatedBy = "author"
	created, err := service.CreateContent(ctx, input)
	if err != nil {
		t.Fatalf("CreateContent(%q) error = %v", input.Title, err)
	}
	return created
}

func assertContains(t *testing.T, value string, want string) {
	t.Helper()
	if !strings.Contains(value, want) {
		t.Fatalf("expected value to contain %q\nvalue:\n%s", want, value)
	}
}

func assertOrdered(t *testing.T, value string, first string, second string) {
	t.Helper()
	firstIndex := strings.Index(value, first)
	if firstIndex == -1 {
		t.Fatalf("missing first ordered value %q", first)
	}
	secondIndex := strings.Index(value, second)
	if secondIndex == -1 {
		t.Fatalf("missing second ordered value %q", second)
	}
	if firstIndex >= secondIndex {
		t.Fatalf("%q should appear before %q\nvalue:\n%s", first, second, value)
	}
}
