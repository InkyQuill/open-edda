package project

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

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

func (s *Service) ListEntrySections(ctx context.Context, contentID string) ([]EntrySection, error) {
	sections, err := s.queries.ListEntrySections(ctx, contentID)
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
