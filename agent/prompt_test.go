package agent

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"git.inkyquill.net/inky/writer/project"
	"git.inkyquill.net/inky/writer/skill"
	"git.inkyquill.net/inky/writer/store"
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

	assertContains(t, bundle.SystemMessage, "You are a fiction writing assistant working inside Edda. Preserve the author's intent, respect established project facts, and use available tools to inspect project context before making claims. Do not invent durable worldbuilding facts unless the author asks you to brainstorm.")
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

func TestPromptBuildActionPromptAddsSkillGuidanceSources(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Prompt Skill Project")
	service := NewService(db, projectService, nil)
	skillService := skill.NewService(db)
	service.SetSkillService(skillService)
	createPromptProfile(t, ctx, service, storyProject.ID)
	chapter := createPromptContent(t, ctx, projectService, project.CreateContentInput{
		ProjectID:    storyProject.ID,
		Kind:         project.KindChapter,
		Title:        "Chapter 1",
		BodyMarkdown: "Opening.",
	})

	describedSkill := installPromptSkill(t, ctx, skillService, storyProject.ID, skill.ImportedSkill{
		Name:                 `style-pass`,
		Description:          `Use when rewriting prose for style & rhythm`,
		InstructionsMarkdown: "Full instructions must not be preloaded.",
		ScriptCount:          1,
		ScriptsDisabled:      true,
	})
	undescribedSkill := installPromptSkill(t, ctx, skillService, storyProject.ID, skill.ImportedSkill{
		Name:                 "private-notes",
		InstructionsMarkdown: "Private instructions.",
	})
	session, err := service.CreateSession(ctx, CreateSessionInput{
		ProjectID:  storyProject.ID,
		Title:      "Skill session",
		ActionKind: ActionKindContinuation,
		ApplyMode:  ApplyModePreview,
		SkillIDs:   []string{describedSkill.ID},
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	bundle, err := service.BuildActionPrompt(ctx, BuildPromptInput{
		ProjectID:       storyProject.ID,
		SessionID:       session.ID,
		ActionKind:      ActionKindContinuation,
		TargetContentID: chapter.ID,
	})
	if err != nil {
		t.Fatalf("BuildActionPrompt() error = %v", err)
	}

	sourceKeys := make([]string, 0, len(bundle.ContextSources))
	for _, source := range bundle.ContextSources {
		sourceKeys = append(sourceKeys, source.SourceKey)
	}
	for _, key := range []string{"prompt_profile", "writing_briefs", "target_content_window", "provider_disclosure", "tool_catalog", "available_skills", "selected_skills"} {
		if !slices.Contains(sourceKeys, key) {
			t.Fatalf("context source keys = %v, want key %q", sourceKeys, key)
		}
	}

	available := mustContextSource(t, bundle, "available_skills")
	assertContains(t, available.RenderedMarkdown, "Skills provide specialized instructions and workflows for specific writing tasks.")
	assertContains(t, available.RenderedMarkdown, "Use the skill tool to load a skill when the task matches its description.")
	assertContains(t, available.RenderedMarkdown, "<available_skills>")
	assertContains(t, available.RenderedMarkdown, "<id>"+describedSkill.ID+"</id>")
	assertContains(t, available.RenderedMarkdown, "<name>style-pass</name>")
	assertContains(t, available.RenderedMarkdown, "<description>Use when rewriting prose for style &amp; rhythm</description>")
	if strings.Contains(available.RenderedMarkdown, undescribedSkill.ID) || strings.Contains(available.RenderedMarkdown, "private-notes") {
		t.Fatalf("available_skills advertised undescribed skill:\n%s", available.RenderedMarkdown)
	}
	if strings.Contains(available.RenderedMarkdown, "Full instructions must not be preloaded.") {
		t.Fatalf("available_skills included full instructions:\n%s", available.RenderedMarkdown)
	}

	selected := mustContextSource(t, bundle, "selected_skills")
	assertContains(t, selected.RenderedMarkdown, "The author selected these skills for this session.")
	assertContains(t, selected.RenderedMarkdown, "<selected_skills>")
	assertContains(t, selected.RenderedMarkdown, "<id>"+describedSkill.ID+"</id>")
	assertContains(t, selected.RenderedMarkdown, `<script_status disabled="true" count="1">Scripts are available only as inert reference files.</script_status>`)
}

func TestPromptBuildActionPromptAddsStableAvailableSkillsSourceWhenNoDescribedSkills(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Prompt Undescribed Skill Project")
	service := NewService(db, projectService, nil)
	skillService := skill.NewService(db)
	service.SetSkillService(skillService)
	createPromptProfile(t, ctx, service, storyProject.ID)
	chapter := createPromptContent(t, ctx, projectService, project.CreateContentInput{
		ProjectID:    storyProject.ID,
		Kind:         project.KindChapter,
		Title:        "Chapter 1",
		BodyMarkdown: "Opening.",
	})
	installPromptSkill(t, ctx, skillService, storyProject.ID, skill.ImportedSkill{
		Name:                 "private-notes",
		InstructionsMarkdown: "Private instructions.",
	})

	bundle, err := service.BuildActionPrompt(ctx, BuildPromptInput{
		ProjectID:       storyProject.ID,
		ActionKind:      ActionKindContinuation,
		TargetContentID: chapter.ID,
	})
	if err != nil {
		t.Fatalf("BuildActionPrompt() error = %v", err)
	}

	available := mustContextSource(t, bundle, "available_skills")
	assertContains(t, available.RenderedMarkdown, "No described skills are available for this project.")
	if _, ok := contextSource(bundle, "selected_skills"); ok {
		t.Fatalf("selected_skills source rendered without selected skills")
	}
}

func TestPromptBuildActionPromptValidatesSessionProject(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	projectOne := createTestProject(t, ctx, projectService, "author-1", "Prompt Session Project One")
	projectTwo := createTestProject(t, ctx, projectService, "author-1", "Prompt Session Project Two")
	service := NewService(db, projectService, nil)
	createPromptProfile(t, ctx, service, projectOne.ID)
	createPromptProfile(t, ctx, service, projectTwo.ID)

	otherSession, err := service.CreateSession(ctx, CreateSessionInput{
		ProjectID:  projectTwo.ID,
		Title:      "Other session",
		ActionKind: ActionKindContinuation,
		ApplyMode:  ApplyModePreview,
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	chapter := createPromptContent(t, ctx, projectService, project.CreateContentInput{
		ProjectID:    projectOne.ID,
		Kind:         project.KindChapter,
		Title:        "Chapter 1",
		BodyMarkdown: "Opening.",
	})

	_, err = service.BuildActionPrompt(ctx, BuildPromptInput{
		ProjectID:       projectOne.ID,
		SessionID:       otherSession.ID,
		ActionKind:      ActionKindContinuation,
		TargetContentID: chapter.ID,
	})
	assertErrorContains(t, err, "get agent session")
}

func TestPromptBuildActionPromptRequiresChapterTarget(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Prompt Target Validation Project")
	service := NewService(db, projectService, nil)
	createPromptProfile(t, ctx, service, storyProject.ID)

	t.Run("missing target ID", func(t *testing.T) {
		_, err := service.BuildActionPrompt(ctx, BuildPromptInput{
			ProjectID:  storyProject.ID,
			ActionKind: ActionKindContinuation,
		})
		assertErrorContains(t, err, "target content ID is required")
	})

	t.Run("wrong target kind", func(t *testing.T) {
		brief := createPromptContent(t, ctx, projectService, project.CreateContentInput{
			ProjectID:    storyProject.ID,
			Kind:         project.KindWritingBrief,
			Title:        "Not a Chapter",
			BodyMarkdown: "Brief text.",
		})

		_, err := service.BuildActionPrompt(ctx, BuildPromptInput{
			ProjectID:       storyProject.ID,
			ActionKind:      ActionKindContinuation,
			TargetContentID: brief.ID,
		})
		assertErrorContains(t, err, "target content must be chapter")
	})
}

func TestPromptBuildActionPromptValidatesWritingBriefMetadata(t *testing.T) {
	tests := []struct {
		name         string
		metadataJSON string
		wantErr      string
	}{
		{
			name:         "project scope with content item",
			metadataJSON: `{"scope":"project","contentItemId":"chapter-1"}`,
			wantErr:      "project-scoped writing brief must not have contentItemId",
		},
		{
			name:         "chapter scope without content item",
			metadataJSON: `{"scope":"chapter"}`,
			wantErr:      "chapter-scoped writing brief must have contentItemId",
		},
		{
			name:         "unknown scope",
			metadataJSON: `{"scope":"scene"}`,
			wantErr:      `unknown writing brief scope "scene"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openMigratedTestDB(t)
			ctx := context.Background()
			projectService := project.NewService(db)
			storyProject := createTestProject(t, ctx, projectService, "author-1", "Prompt Brief Metadata Project")
			service := NewService(db, projectService, nil)
			createPromptProfile(t, ctx, service, storyProject.ID)
			chapter := createPromptContent(t, ctx, projectService, project.CreateContentInput{
				ProjectID:    storyProject.ID,
				Kind:         project.KindChapter,
				Title:        "Chapter 1",
				BodyMarkdown: "Opening.",
			})
			createPromptContent(t, ctx, projectService, project.CreateContentInput{
				ProjectID:    storyProject.ID,
				Kind:         project.KindWritingBrief,
				Title:        "Invalid Brief",
				BodyMarkdown: "Invalid metadata.",
				MetadataJSON: tt.metadataJSON,
			})

			_, err := service.BuildActionPrompt(ctx, BuildPromptInput{
				ProjectID:       storyProject.ID,
				ActionKind:      ActionKindContinuation,
				TargetContentID: chapter.ID,
			})
			assertErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestPromptBuildActionPromptTreatsContentItemIDWithoutScopeAsTargetBrief(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Prompt Brief Compatibility Project")
	service := NewService(db, projectService, nil)
	createPromptProfile(t, ctx, service, storyProject.ID)
	chapter := createPromptContent(t, ctx, projectService, project.CreateContentInput{
		ProjectID:    storyProject.ID,
		Kind:         project.KindChapter,
		Title:        "Chapter 1",
		BodyMarkdown: "Opening.",
	})
	createPromptContent(t, ctx, projectService, project.CreateContentInput{
		ProjectID:    storyProject.ID,
		Kind:         project.KindWritingBrief,
		Title:        "Legacy Chapter Brief",
		BodyMarkdown: "Legacy target-specific brief.",
		MetadataJSON: `{"contentItemId":"` + chapter.ID + `"}`,
	})

	bundle, err := service.BuildActionPrompt(ctx, BuildPromptInput{
		ProjectID:       storyProject.ID,
		ActionKind:      ActionKindContinuation,
		TargetContentID: chapter.ID,
	})
	if err != nil {
		t.Fatalf("BuildActionPrompt() error = %v", err)
	}

	assertContains(t, bundle.DeveloperMessage, "Legacy target-specific brief.")
}

func TestPromptBuildActionPromptOrdersWritingBriefsDeterministically(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	storyProject := createTestProject(t, ctx, projectService, "author-1", "Prompt Brief Ordering Project")
	service := NewService(db, projectService, nil)
	createPromptProfile(t, ctx, service, storyProject.ID)
	chapter := createPromptContent(t, ctx, projectService, project.CreateContentInput{
		ProjectID:    storyProject.ID,
		Kind:         project.KindChapter,
		Title:        "Chapter 1",
		BodyMarkdown: "Opening.",
	})

	insertPromptContentRow(t, ctx, service, store.CreateContentItemParams{
		ID:              "brief-b",
		ProjectID:       storyProject.ID,
		Kind:            string(project.KindWritingBrief),
		Title:           "Same Brief",
		Slug:            "same-brief-b",
		BodyMarkdown:    "Project brief B.",
		MetadataJson:    `{}`,
		SortOrder:       10,
		CurrentRevision: 1,
	})
	insertPromptContentRow(t, ctx, service, store.CreateContentItemParams{
		ID:              "brief-a",
		ProjectID:       storyProject.ID,
		Kind:            string(project.KindWritingBrief),
		Title:           "Same Brief",
		Slug:            "same-brief-a",
		BodyMarkdown:    "Project brief A.",
		MetadataJson:    `{}`,
		SortOrder:       10,
		CurrentRevision: 1,
	})
	insertPromptContentRow(t, ctx, service, store.CreateContentItemParams{
		ID:              "brief-target-b",
		ProjectID:       storyProject.ID,
		Kind:            string(project.KindWritingBrief),
		Title:           "Same Target Brief",
		Slug:            "same-target-brief-b",
		BodyMarkdown:    "Target brief B.",
		MetadataJson:    `{"scope":"chapter","contentItemId":"` + chapter.ID + `"}`,
		SortOrder:       20,
		CurrentRevision: 1,
	})
	insertPromptContentRow(t, ctx, service, store.CreateContentItemParams{
		ID:              "brief-target-a",
		ProjectID:       storyProject.ID,
		Kind:            string(project.KindWritingBrief),
		Title:           "Same Target Brief",
		Slug:            "same-target-brief-a",
		BodyMarkdown:    "Target brief A.",
		MetadataJson:    `{"scope":"chapter","contentItemId":"` + chapter.ID + `"}`,
		SortOrder:       20,
		CurrentRevision: 1,
	})

	bundle, err := service.BuildActionPrompt(ctx, BuildPromptInput{
		ProjectID:       storyProject.ID,
		ActionKind:      ActionKindContinuation,
		TargetContentID: chapter.ID,
	})
	if err != nil {
		t.Fatalf("BuildActionPrompt() error = %v", err)
	}

	assertOrdered(t, bundle.DeveloperMessage, "Project brief A.", "Project brief B.")
	assertOrdered(t, bundle.DeveloperMessage, "Project brief B.", "Target brief A.")
	assertOrdered(t, bundle.DeveloperMessage, "Target brief A.", "Target brief B.")
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

func createPromptProfile(t *testing.T, ctx context.Context, service *Service, projectID string) {
	t.Helper()
	_, err := service.UpsertPromptProfile(ctx, UpsertPromptProfileInput{
		ProjectID:                 projectID,
		Genre:                     "fantasy",
		Tense:                     "past tense",
		POV:                       "third person",
		Voice:                     "clear",
		PromptRecordRetentionDays: 30,
	})
	if err != nil {
		t.Fatalf("UpsertPromptProfile() error = %v", err)
	}
}

func installPromptSkill(t *testing.T, ctx context.Context, service *skill.Service, projectID string, imported skill.ImportedSkill) skill.Skill {
	t.Helper()

	installed, err := service.Install(ctx, skill.InstallInput{
		ProjectID:   projectID,
		SourceType:  skill.SourceTypeUpload,
		SourceLabel: imported.Name + ".zip",
		Imported:    imported,
	})
	if err != nil {
		t.Fatalf("Install(%q) error = %v", imported.Name, err)
	}
	return installed
}

func mustContextSource(t *testing.T, bundle PromptBundle, key string) ContextSourceSnapshot {
	t.Helper()
	source, ok := contextSource(bundle, key)
	if !ok {
		t.Fatalf("missing context source %q", key)
	}
	return source
}

func contextSource(bundle PromptBundle, key string) (ContextSourceSnapshot, bool) {
	for _, source := range bundle.ContextSources {
		if source.SourceKey == key {
			return source, true
		}
	}
	return ContextSourceSnapshot{}, false
}

func insertPromptContentRow(t *testing.T, ctx context.Context, service *Service, input store.CreateContentItemParams) {
	t.Helper()
	now := nowString()
	input.CreatedAt = now
	input.UpdatedAt = now
	if err := service.queries.CreateContentItem(ctx, input); err != nil {
		t.Fatalf("CreateContentItem(%q) error = %v", input.ID, err)
	}
}

func assertContains(t *testing.T, value string, want string) {
	t.Helper()
	if !strings.Contains(value, want) {
		t.Fatalf("expected value to contain %q\nvalue:\n%s", want, value)
	}
}

func assertErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want containing %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %q, want containing %q", err.Error(), want)
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
