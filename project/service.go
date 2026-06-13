package project

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"git.inkyquill.net/inky/writer/markdownio"
	"git.inkyquill.net/inky/writer/store"
	"github.com/mattn/go-sqlite3"
)

var ErrConflict = errors.New("content revision conflict")

type Service struct {
	db      *sql.DB
	queries *store.Queries
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db:      db,
		queries: store.New(db),
	}
}

func (s *Service) CreateProject(ctx context.Context, input CreateProjectInput) (StoryProject, error) {
	now := nowString()
	project := StoryProject{
		ID:       newID("project"),
		Title:    input.Title,
		Slug:     slugify(input.Title),
		Language: input.Language,
	}

	if err := s.queries.CreateStoryProject(ctx, store.CreateStoryProjectParams{
		ID:        project.ID,
		AuthorID:  input.AuthorID,
		Title:     project.Title,
		Slug:      project.Slug,
		Language:  project.Language,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		if isSQLiteConstraint(err, sqlite3.ErrConstraintUnique) {
			return StoryProject{}, ErrConflict
		}
		return StoryProject{}, fmt.Errorf("create story project: %w", err)
	}

	return project, nil
}

func (s *Service) ImportElysiumProject(ctx context.Context, authorID string, title string, language string, items []markdownio.ImportedItem) (StoryProject, error) {
	now := nowString()
	project := StoryProject{
		ID:       newID("project"),
		Title:    title,
		Slug:     slugify(title),
		Language: language,
	}

	if err := s.inTx(ctx, func(queries *store.Queries) error {
		if err := queries.CreateStoryProject(ctx, store.CreateStoryProjectParams{
			ID:        project.ID,
			AuthorID:  authorID,
			Title:     project.Title,
			Slug:      project.Slug,
			Language:  project.Language,
			CreatedAt: now,
			UpdatedAt: now,
		}); err != nil {
			if isSQLiteConstraint(err, sqlite3.ErrConstraintUnique) {
				return ErrConflict
			}
			return fmt.Errorf("create story project: %w", err)
		}

		for index, imported := range items {
			item := ContentItem{
				ID:              newID("content"),
				ProjectID:       project.ID,
				Kind:            ContentKind(imported.Kind),
				Title:           imported.Title,
				Slug:            slugify(imported.Title),
				BodyMarkdown:    imported.BodyMarkdown,
				MetadataJSON:    defaultJSON(imported.MetadataJSON),
				SortOrder:       int64(index),
				CurrentRevision: 1,
			}
			if err := queries.CreateContentItem(ctx, store.CreateContentItemParams{
				ID:              item.ID,
				ProjectID:       item.ProjectID,
				Kind:            string(item.Kind),
				Title:           item.Title,
				Slug:            item.Slug,
				BodyMarkdown:    item.BodyMarkdown,
				MetadataJson:    item.MetadataJSON,
				SortOrder:       item.SortOrder,
				CurrentRevision: item.CurrentRevision,
				CreatedAt:       now,
				UpdatedAt:       now,
			}); err != nil {
				if isSQLiteConstraint(err, sqlite3.ErrConstraintUnique) {
					return ErrConflict
				}
				return fmt.Errorf("create content item: %w", err)
			}
			if err := queries.CreateRevision(ctx, store.CreateRevisionParams{
				ID:             newID("revision"),
				ContentItemID:  item.ID,
				RevisionNumber: item.CurrentRevision,
				BodyMarkdown:   item.BodyMarkdown,
				MetadataJson:   item.MetadataJSON,
				Reason:         "Elysium import",
				CreatedBy:      "import",
				CreatedAt:      now,
				AgentSessionID: sql.NullString{},
				ActionKind:     "",
				ModelVariantID: sql.NullString{},
				SkillID:        "",
			}); err != nil {
				return fmt.Errorf("create initial revision: %w", err)
			}
			for _, section := range imported.Sections {
				if err := queries.CreateEntrySection(ctx, store.CreateEntrySectionParams{
					ID:            newID("section"),
					ContentItemID: item.ID,
					Heading:       section.Heading,
					BodyMarkdown:  section.BodyMarkdown,
					SortOrder:     int64(section.SortOrder),
				}); err != nil {
					return fmt.Errorf("create entry section: %w", err)
				}
			}
			for _, relation := range imported.Relations {
				if err := queries.CreateEntryRelation(ctx, store.CreateEntryRelationParams{
					ID:           newID("relation"),
					ProjectID:    project.ID,
					SourceItemID: item.ID,
					TargetItemID: sql.NullString{},
					TargetTitle:  relation.TargetTitle,
					RelationType: relation.RelationType,
					CreatedAt:    now,
				}); err != nil {
					return fmt.Errorf("create entry relation: %w", err)
				}
			}
		}
		return nil
	}); err != nil {
		return StoryProject{}, err
	}

	return project, nil
}

func (s *Service) ListProjects(ctx context.Context, authorID string) ([]StoryProject, error) {
	projects, err := s.queries.ListStoryProjects(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("list story projects: %w", err)
	}

	result := make([]StoryProject, 0, len(projects))
	for _, project := range projects {
		result = append(result, storyProjectFromStore(project))
	}
	return result, nil
}

func (s *Service) CreateContent(ctx context.Context, input CreateContentInput) (ContentItem, error) {
	now := nowString()
	createdBy, err := createdBy(input.CreatedBy)
	if err != nil {
		return ContentItem{}, err
	}
	item := ContentItem{
		ID:              newID("content"),
		ProjectID:       input.ProjectID,
		Kind:            input.Kind,
		Title:           input.Title,
		Slug:            slugify(input.Title),
		BodyMarkdown:    input.BodyMarkdown,
		MetadataJSON:    defaultJSON(input.MetadataJSON),
		SortOrder:       input.SortOrder,
		CurrentRevision: 1,
	}

	if err := s.inTx(ctx, func(queries *store.Queries) error {
		if _, err := queries.GetStoryProjectByID(ctx, input.ProjectID); err != nil {
			return fmt.Errorf("get story project: %w", err)
		}

		if err := queries.CreateContentItem(ctx, store.CreateContentItemParams{
			ID:              item.ID,
			ProjectID:       item.ProjectID,
			Kind:            string(item.Kind),
			Title:           item.Title,
			Slug:            item.Slug,
			BodyMarkdown:    item.BodyMarkdown,
			MetadataJson:    item.MetadataJSON,
			SortOrder:       item.SortOrder,
			CurrentRevision: item.CurrentRevision,
			CreatedAt:       now,
			UpdatedAt:       now,
		}); err != nil {
			if isSQLiteConstraint(err, sqlite3.ErrConstraintUnique) {
				return ErrConflict
			}
			return fmt.Errorf("create content item: %w", err)
		}

		if err := queries.CreateRevision(ctx, store.CreateRevisionParams{
			ID:             newID("revision"),
			ContentItemID:  item.ID,
			RevisionNumber: item.CurrentRevision,
			BodyMarkdown:   item.BodyMarkdown,
			MetadataJson:   item.MetadataJSON,
			Reason:         emptyDefault(input.Reason, "initial content"),
			CreatedBy:      createdBy,
			CreatedAt:      now,
			AgentSessionID: sql.NullString{},
			ActionKind:     "",
			ModelVariantID: sql.NullString{},
			SkillID:        "",
		}); err != nil {
			return fmt.Errorf("create initial revision: %w", err)
		}

		return nil
	}); err != nil {
		return ContentItem{}, err
	}

	return item, nil
}

func (s *Service) ListContent(ctx context.Context, projectID string, kind ContentKind) ([]ContentItem, error) {
	if _, err := s.queries.GetStoryProjectByID(ctx, projectID); err != nil {
		return nil, fmt.Errorf("get story project: %w", err)
	}

	items, err := s.queries.ListContentItems(ctx, store.ListContentItemsParams{
		ProjectID: projectID,
		Kind:      string(kind),
	})
	if err != nil {
		return nil, fmt.Errorf("list content items: %w", err)
	}

	result := make([]ContentItem, 0, len(items))
	for _, item := range items {
		result = append(result, contentItemFromStore(item))
	}
	return result, nil
}

func (s *Service) ListProjectContent(ctx context.Context, projectID string) ([]ContentItem, error) {
	if _, err := s.queries.GetStoryProjectByID(ctx, projectID); err != nil {
		return nil, fmt.Errorf("get story project: %w", err)
	}

	items, err := s.queries.ListProjectContentItems(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project content items: %w", err)
	}

	result := make([]ContentItem, 0, len(items))
	for _, item := range items {
		result = append(result, contentItemFromStore(item))
	}
	return result, nil
}

func (s *Service) ProjectMap(ctx context.Context, projectID string) (ProjectMap, error) {
	storyProject, err := s.queries.GetStoryProjectByID(ctx, projectID)
	if err != nil {
		return ProjectMap{}, fmt.Errorf("get story project: %w", err)
	}

	items, err := s.ListProjectContent(ctx, projectID)
	if err != nil {
		return ProjectMap{}, err
	}

	content := make([]ProjectMapContentItem, 0, len(items))
	for _, item := range items {
		mapped := ProjectMapContentItem{ContentItem: item}
		if item.Kind == KindStoryBibleEntry {
			sections, err := s.ListEntrySections(ctx, projectID, item.ID)
			if err != nil {
				return ProjectMap{}, err
			}
			relations, err := s.ListEntryRelations(ctx, projectID, item.ID)
			if err != nil {
				return ProjectMap{}, err
			}
			mapped.Sections = sections
			mapped.Relations = relations
		}
		content = append(content, mapped)
	}

	return ProjectMap{
		Project: storyProjectFromStore(storyProject),
		Content: content,
	}, nil
}

func (s *Service) SearchContent(ctx context.Context, input SearchContentInput) ([]ContentItem, error) {
	if _, err := s.queries.GetStoryProjectByID(ctx, input.ProjectID); err != nil {
		return nil, fmt.Errorf("get story project: %w", err)
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}

	var (
		items []store.ContentItem
		err   error
	)
	if strings.TrimSpace(input.Query) == "" {
		if input.Kind != "" {
			items, err = s.queries.ListContentItems(ctx, store.ListContentItemsParams{
				ProjectID: input.ProjectID,
				Kind:      string(input.Kind),
			})
		} else {
			items, err = s.queries.ListProjectContentItems(ctx, input.ProjectID)
		}
	} else {
		items, err = s.queries.SearchContent(ctx, store.SearchContentParams{
			Query:     input.Query,
			ProjectID: input.ProjectID,
			Limit:     limit,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("search content: %w", err)
	}

	result := make([]ContentItem, 0, len(items))
	for _, item := range items {
		content := contentItemFromStore(item)
		if input.Kind != "" && content.Kind != input.Kind {
			continue
		}
		matches, err := contentMatchesMetadata(content.MetadataJSON, input.MetadataFilters, input.Tags)
		if err != nil {
			return nil, err
		}
		if !matches {
			continue
		}
		result = append(result, content)
		if int64(len(result)) >= limit {
			break
		}
	}
	return result, nil
}

func (s *Service) GetContent(ctx context.Context, projectID, contentID string) (ContentItem, error) {
	item, err := s.queries.GetContentItem(ctx, store.GetContentItemParams{
		ID:        contentID,
		ProjectID: projectID,
	})
	if err != nil {
		return ContentItem{}, fmt.Errorf("get content item: %w", err)
	}
	return contentItemFromStore(item), nil
}

func (s *Service) UpdateContent(ctx context.Context, input UpdateContentInput) (ContentItem, error) {
	createdBy, err := createdBy(input.CreatedBy)
	if err != nil {
		return ContentItem{}, err
	}

	var updated ContentItem
	if err := s.inTx(ctx, func(queries *store.Queries) error {
		item, err := queries.GetContentItem(ctx, store.GetContentItemParams{
			ID:        input.ContentID,
			ProjectID: input.ProjectID,
		})
		if err != nil {
			return fmt.Errorf("get content item: %w", err)
		}
		if item.CurrentRevision != input.ExpectedRevision {
			return ErrConflict
		}

		nextRevision := input.ExpectedRevision + 1
		metadataJSON := defaultJSON(input.MetadataJSON)
		now := nowString()

		affected, err := queries.UpdateContentItemBody(ctx, store.UpdateContentItemBodyParams{
			BodyMarkdown:     input.BodyMarkdown,
			MetadataJson:     metadataJSON,
			NextRevision:     nextRevision,
			UpdatedAt:        now,
			ID:               input.ContentID,
			ProjectID:        input.ProjectID,
			ExpectedRevision: input.ExpectedRevision,
		})
		if err != nil {
			return fmt.Errorf("update content item: %w", err)
		}
		if affected == 0 {
			return ErrConflict
		}

		if err := queries.CreateRevision(ctx, store.CreateRevisionParams{
			ID:             newID("revision"),
			ContentItemID:  input.ContentID,
			RevisionNumber: nextRevision,
			BodyMarkdown:   input.BodyMarkdown,
			MetadataJson:   metadataJSON,
			Reason:         emptyDefault(input.Reason, "content update"),
			CreatedBy:      createdBy,
			CreatedAt:      now,
			AgentSessionID: sql.NullString{},
			ActionKind:     "",
			ModelVariantID: sql.NullString{},
			SkillID:        "",
		}); err != nil {
			return fmt.Errorf("create revision: %w", err)
		}

		updated = contentItemFromStore(store.ContentItem{
			ID:              item.ID,
			ProjectID:       item.ProjectID,
			Kind:            item.Kind,
			Title:           item.Title,
			Slug:            item.Slug,
			BodyMarkdown:    input.BodyMarkdown,
			MetadataJson:    metadataJSON,
			SortOrder:       item.SortOrder,
			CurrentRevision: nextRevision,
		})
		return nil
	}); err != nil {
		return ContentItem{}, err
	}

	return updated, nil
}

func (s *Service) CreateEntrySection(ctx context.Context, input CreateEntrySectionInput) error {
	if _, err := s.queries.GetContentItem(ctx, store.GetContentItemParams{
		ID:        input.ContentItemID,
		ProjectID: input.ProjectID,
	}); err != nil {
		return fmt.Errorf("get content item: %w", err)
	}
	if err := s.queries.CreateEntrySection(ctx, store.CreateEntrySectionParams{
		ID:            newID("section"),
		ContentItemID: input.ContentItemID,
		Heading:       input.Heading,
		BodyMarkdown:  input.BodyMarkdown,
		SortOrder:     input.SortOrder,
	}); err != nil {
		return fmt.Errorf("create entry section: %w", err)
	}
	return nil
}

func (s *Service) ListEntrySections(ctx context.Context, projectID, contentID string) ([]EntrySection, error) {
	sections, err := s.queries.ListEntrySections(ctx, store.ListEntrySectionsParams{
		ContentItemID: contentID,
		ProjectID:     projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("list entry sections: %w", err)
	}

	result := make([]EntrySection, 0, len(sections))
	for _, section := range sections {
		result = append(result, EntrySection{
			Heading:      section.Heading,
			BodyMarkdown: section.BodyMarkdown,
			SortOrder:    section.SortOrder,
		})
	}
	return result, nil
}

func (s *Service) CreateAttachedNote(ctx context.Context, input CreateAttachedNoteInput) (AttachedNote, error) {
	if _, err := s.queries.GetStoryProjectByID(ctx, input.ProjectID); err != nil {
		return AttachedNote{}, fmt.Errorf("get story project: %w", err)
	}
	if input.ContentItemID != "" {
		if _, err := s.GetContent(ctx, input.ProjectID, input.ContentItemID); err != nil {
			return AttachedNote{}, err
		}
	}

	source := emptyDefault(input.Source, "author")
	if source != "author" && source != "read_and_check" {
		return AttachedNote{}, fmt.Errorf("invalid attached note source %q", source)
	}

	now := nowString()
	note := AttachedNote{
		ID:             newID("note"),
		ProjectID:      input.ProjectID,
		ContentItemID:  input.ContentItemID,
		SelectionStart: input.SelectionStart,
		SelectionEnd:   input.SelectionEnd,
		Title:          input.Title,
		BodyMarkdown:   input.BodyMarkdown,
		Source:         source,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO attached_notes (
			id, project_id, content_item_id, selection_start, selection_end,
			title, body_markdown, source, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		note.ID,
		note.ProjectID,
		nullString(note.ContentItemID),
		nullInt64(note.SelectionStart),
		nullInt64(note.SelectionEnd),
		note.Title,
		note.BodyMarkdown,
		note.Source,
		note.CreatedAt,
		note.UpdatedAt,
	)
	if err != nil {
		return AttachedNote{}, fmt.Errorf("create attached note: %w", err)
	}

	return note, nil
}

func (s *Service) CreateEntryRelation(ctx context.Context, input CreateEntryRelationInput) error {
	if err := s.queries.CreateEntryRelation(ctx, store.CreateEntryRelationParams{
		ID:           newID("relation"),
		ProjectID:    input.ProjectID,
		SourceItemID: input.SourceItemID,
		TargetItemID: sql.NullString{},
		TargetTitle:  input.TargetTitle,
		RelationType: input.RelationType,
		CreatedAt:    nowString(),
	}); err != nil {
		return fmt.Errorf("create entry relation: %w", err)
	}
	return nil
}

func (s *Service) ListEntryRelations(ctx context.Context, projectID, contentID string) ([]EntryRelation, error) {
	relations, err := s.queries.ListEntryRelations(ctx, store.ListEntryRelationsParams{
		ProjectID:    projectID,
		SourceItemID: contentID,
	})
	if err != nil {
		return nil, fmt.Errorf("list entry relations: %w", err)
	}

	result := make([]EntryRelation, 0, len(relations))
	for _, relation := range relations {
		result = append(result, EntryRelation{
			TargetTitle:  relation.TargetTitle,
			RelationType: relation.RelationType,
		})
	}
	return result, nil
}

func (s *Service) ListRevisions(ctx context.Context, projectID, contentID string) ([]Revision, error) {
	if _, err := s.GetContent(ctx, projectID, contentID); err != nil {
		return nil, err
	}

	revisions, err := s.queries.ListRevisions(ctx, store.ListRevisionsParams{
		ContentItemID: contentID,
		ProjectID:     projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("list revisions: %w", err)
	}

	result := make([]Revision, 0, len(revisions))
	for _, revision := range revisions {
		result = append(result, revisionFromStore(revision))
	}
	return result, nil
}

func (s *Service) inTx(ctx context.Context, fn func(*store.Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := fn(s.queries.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	committed = true

	return nil
}

func storyProjectFromStore(project store.StoryProject) StoryProject {
	return StoryProject{
		ID:       project.ID,
		Title:    project.Title,
		Slug:     project.Slug,
		Language: project.Language,
	}
}

func contentItemFromStore(item store.ContentItem) ContentItem {
	return ContentItem{
		ID:              item.ID,
		ProjectID:       item.ProjectID,
		Kind:            ContentKind(item.Kind),
		Title:           item.Title,
		Slug:            item.Slug,
		BodyMarkdown:    item.BodyMarkdown,
		MetadataJSON:    item.MetadataJson,
		SortOrder:       item.SortOrder,
		CurrentRevision: item.CurrentRevision,
	}
}

func revisionFromStore(revision store.Revision) Revision {
	return Revision{
		ID:             revision.ID,
		ContentItemID:  revision.ContentItemID,
		RevisionNumber: revision.RevisionNumber,
		BodyMarkdown:   revision.BodyMarkdown,
		MetadataJSON:   revision.MetadataJson,
		Reason:         revision.Reason,
		CreatedBy:      revision.CreatedBy,
		CreatedAt:      revision.CreatedAt,
	}
}

func contentMatchesMetadata(metadataJSON string, filters map[string]string, tags []string) (bool, error) {
	if len(filters) == 0 && len(tags) == 0 {
		return true, nil
	}

	var metadata map[string]any
	if err := json.Unmarshal([]byte(defaultJSON(metadataJSON)), &metadata); err != nil {
		return false, err
	}

	for key, want := range filters {
		got, ok := metadata[key]
		if !ok {
			return false, nil
		}
		if gotString, ok := got.(string); !ok || gotString != want {
			return false, nil
		}
	}

	if len(tags) > 0 {
		values, ok := metadata["tags"].([]any)
		if !ok {
			return false, nil
		}
		available := make(map[string]bool, len(values))
		for _, value := range values {
			if tag, ok := value.(string); ok {
				available[tag] = true
			}
		}
		for _, tag := range tags {
			if !available[tag] {
				return false, nil
			}
		}
	}

	return true, nil
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value != 0}
}

func defaultJSON(value string) string {
	if value == "" {
		return "{}"
	}
	return value
}

func emptyDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func createdBy(value string) (string, error) {
	value = emptyDefault(value, "author")
	switch value {
	case "author", "import", "restore", "agent":
		return value, nil
	default:
		return "", fmt.Errorf("invalid created_by %q", value)
	}
}

func slugify(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(value)), "-")
}

var idCounter uint64

func newID(prefix string) string {
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UTC().UnixNano(), atomic.AddUint64(&idCounter, 1))
}

func nowString() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func isSQLiteConstraint(err error, code sqlite3.ErrNoExtended) bool {
	var sqliteErr sqlite3.Error
	return errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == code
}
