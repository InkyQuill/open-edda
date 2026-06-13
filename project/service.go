package project

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"git.inkyquill.net/inky/writer/store"
)

var ErrConflict = errors.New("content revision conflict")

type Service struct {
	queries *store.Queries
}

func NewService(queries *store.Queries) *Service {
	return &Service{queries: queries}
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
		return StoryProject{}, fmt.Errorf("create story project: %w", err)
	}

	return project, nil
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

	if err := s.queries.CreateContentItem(ctx, store.CreateContentItemParams{
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
		return ContentItem{}, fmt.Errorf("create content item: %w", err)
	}

	if err := s.queries.CreateRevision(ctx, store.CreateRevisionParams{
		ID:             newID("revision"),
		ContentItemID:  item.ID,
		RevisionNumber: item.CurrentRevision,
		BodyMarkdown:   item.BodyMarkdown,
		MetadataJson:   item.MetadataJSON,
		Reason:         emptyDefault(input.Reason, "initial content"),
		CreatedBy:      createdBy,
		CreatedAt:      now,
	}); err != nil {
		return ContentItem{}, fmt.Errorf("create initial revision: %w", err)
	}

	return item, nil
}

func (s *Service) UpdateContent(ctx context.Context, input UpdateContentInput) (ContentItem, error) {
	createdBy, err := createdBy(input.CreatedBy)
	if err != nil {
		return ContentItem{}, err
	}

	item, err := s.queries.GetContentItem(ctx, store.GetContentItemParams{
		ID:        input.ContentID,
		ProjectID: input.ProjectID,
	})
	if err != nil {
		return ContentItem{}, fmt.Errorf("get content item: %w", err)
	}
	if item.CurrentRevision != input.ExpectedRevision {
		return ContentItem{}, ErrConflict
	}

	nextRevision := input.ExpectedRevision + 1
	metadataJSON := defaultJSON(input.MetadataJSON)
	now := nowString()

	affected, err := s.queries.UpdateContentItemBody(ctx, store.UpdateContentItemBodyParams{
		BodyMarkdown:     input.BodyMarkdown,
		MetadataJson:     metadataJSON,
		NextRevision:     nextRevision,
		UpdatedAt:        now,
		ID:               input.ContentID,
		ProjectID:        input.ProjectID,
		ExpectedRevision: input.ExpectedRevision,
	})
	if err != nil {
		return ContentItem{}, fmt.Errorf("update content item: %w", err)
	}
	if affected == 0 {
		return ContentItem{}, ErrConflict
	}

	if err := s.queries.CreateRevision(ctx, store.CreateRevisionParams{
		ID:             newID("revision"),
		ContentItemID:  input.ContentID,
		RevisionNumber: nextRevision,
		BodyMarkdown:   input.BodyMarkdown,
		MetadataJson:   metadataJSON,
		Reason:         emptyDefault(input.Reason, "content update"),
		CreatedBy:      createdBy,
		CreatedAt:      now,
	}); err != nil {
		return ContentItem{}, fmt.Errorf("create revision: %w", err)
	}

	return contentItemFromStore(store.ContentItem{
		ID:              item.ID,
		ProjectID:       item.ProjectID,
		Kind:            item.Kind,
		Title:           item.Title,
		Slug:            item.Slug,
		BodyMarkdown:    input.BodyMarkdown,
		MetadataJson:    metadataJSON,
		SortOrder:       item.SortOrder,
		CurrentRevision: nextRevision,
	}), nil
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
