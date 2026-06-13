package agent

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"time"

	"git.inkyquill.net/inky/writer/project"
	"git.inkyquill.net/inky/writer/store"
)

type Provider interface {
	Name() string
}

type Service struct {
	db             *sql.DB
	queries        *store.Queries
	projectService *project.Service
	provider       Provider
}

func NewService(db *sql.DB, projectService *project.Service, provider Provider) *Service {
	return &Service{
		db:             db,
		queries:        store.New(db),
		projectService: projectService,
		provider:       provider,
	}
}

func (s *Service) CreateSession(ctx context.Context, input CreateSessionInput) (Session, error) {
	if err := validateActionKind(input.ActionKind); err != nil {
		return Session{}, err
	}
	if err := validateApplyMode(input.ApplyMode); err != nil {
		return Session{}, err
	}
	if _, err := s.queries.GetStoryProjectByID(ctx, input.ProjectID); err != nil {
		return Session{}, fmt.Errorf("get story project: %w", err)
	}
	if input.ModelVariantID != "" {
		if _, err := s.queries.GetModelVariantForProject(ctx, store.GetModelVariantForProjectParams{
			ProjectID:      input.ProjectID,
			ModelVariantID: input.ModelVariantID,
		}); err != nil {
			return Session{}, fmt.Errorf("get model variant for project: %w", err)
		}
	}

	now := nowString()
	session := Session{
		ID:             newID("session"),
		ProjectID:      input.ProjectID,
		Title:          input.Title,
		ActionKind:     input.ActionKind,
		ModelVariantID: input.ModelVariantID,
		ApplyMode:      input.ApplyMode,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.queries.CreateAgentSession(ctx, store.CreateAgentSessionParams{
		ID:             session.ID,
		ProjectID:      session.ProjectID,
		Title:          session.Title,
		ActionKind:     string(session.ActionKind),
		ModelVariantID: nullString(session.ModelVariantID),
		ApplyMode:      string(session.ApplyMode),
		CreatedAt:      session.CreatedAt,
		UpdatedAt:      session.UpdatedAt,
	}); err != nil {
		return Session{}, fmt.Errorf("create agent session: %w", err)
	}

	return session, nil
}

func (s *Service) ListSessions(ctx context.Context, input ListSessionsInput) ([]Session, error) {
	if _, err := s.queries.GetStoryProjectByID(ctx, input.ProjectID); err != nil {
		return nil, fmt.Errorf("get story project: %w", err)
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}
	sessions, err := s.queries.ListAgentSessions(ctx, store.ListAgentSessionsParams{
		ProjectID: input.ProjectID,
		Limit:     limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list agent sessions: %w", err)
	}

	result := make([]Session, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, sessionFromStore(session))
	}
	return result, nil
}

func (s *Service) GetSession(ctx context.Context, projectID, sessionID string) (Session, error) {
	session, err := s.queries.GetAgentSession(ctx, store.GetAgentSessionParams{
		ProjectID: projectID,
		ID:        sessionID,
	})
	if err != nil {
		return Session{}, fmt.Errorf("get agent session: %w", err)
	}
	return sessionFromStore(session), nil
}

func (s *Service) AppendMessage(ctx context.Context, input AppendMessageInput) (Message, error) {
	if err := validateMessageRole(input.Role); err != nil {
		return Message{}, err
	}
	if _, err := s.GetSession(ctx, input.ProjectID, input.SessionID); err != nil {
		return Message{}, err
	}

	message := Message{
		ID:           newID("message"),
		SessionID:    input.SessionID,
		Role:         input.Role,
		BodyMarkdown: input.BodyMarkdown,
		MetadataJSON: defaultJSON(input.MetadataJSON),
		CreatedAt:    nowString(),
	}

	affected, err := s.queries.CreateAgentMessageForProject(ctx, store.CreateAgentMessageForProjectParams{
		ID:           message.ID,
		Role:         string(message.Role),
		BodyMarkdown: message.BodyMarkdown,
		MetadataJson: message.MetadataJSON,
		CreatedAt:    message.CreatedAt,
		SessionID:    message.SessionID,
		ProjectID:    input.ProjectID,
	})
	if err != nil {
		return Message{}, fmt.Errorf("create agent message: %w", err)
	}
	if affected != 1 {
		return Message{}, fmt.Errorf("create agent message: affected rows = %d, want 1", affected)
	}

	return message, nil
}

func (s *Service) ListMessages(ctx context.Context, projectID, sessionID string) ([]Message, error) {
	if _, err := s.GetSession(ctx, projectID, sessionID); err != nil {
		return nil, err
	}

	messages, err := s.queries.ListAgentMessagesForProject(ctx, store.ListAgentMessagesForProjectParams{
		SessionID: sessionID,
		ProjectID: projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("list agent messages: %w", err)
	}

	result := make([]Message, 0, len(messages))
	for _, message := range messages {
		result = append(result, messageFromStore(message))
	}
	return result, nil
}

func sessionFromStore(session store.AgentSession) Session {
	return Session{
		ID:             session.ID,
		ProjectID:      session.ProjectID,
		Title:          session.Title,
		ActionKind:     ActionKind(session.ActionKind),
		ModelVariantID: valueString(session.ModelVariantID),
		ApplyMode:      ApplyMode(session.ApplyMode),
		CreatedAt:      session.CreatedAt,
		UpdatedAt:      session.UpdatedAt,
	}
}

func messageFromStore(message store.AgentMessage) Message {
	return Message{
		ID:           message.ID,
		SessionID:    message.SessionID,
		Role:         MessageRole(message.Role),
		BodyMarkdown: message.BodyMarkdown,
		MetadataJSON: message.MetadataJson,
		CreatedAt:    message.CreatedAt,
	}
}

func validateActionKind(value ActionKind) error {
	switch value {
	case ActionKindChat, ActionKindContinuation, ActionKindRewrite, ActionKindReadCheck:
		return nil
	default:
		return fmt.Errorf("invalid action kind %q", value)
	}
}

func validateApplyMode(value ApplyMode) error {
	switch value {
	case ApplyModePreview, ApplyModeDirectApply:
		return nil
	default:
		return fmt.Errorf("invalid apply mode %q", value)
	}
}

func validateMessageRole(value MessageRole) error {
	switch value {
	case MessageRoleUser, MessageRoleAssistant, MessageRoleTool, MessageRoleSystem:
		return nil
	default:
		return fmt.Errorf("invalid message role %q", value)
	}
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func valueString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func defaultJSON(value string) string {
	if value == "" {
		return "{}"
	}
	return value
}

var idCounter uint64
var lastTimestamp int64

func newID(prefix string) string {
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UTC().UnixNano(), atomic.AddUint64(&idCounter, 1))
}

func nowString() string {
	for {
		next := time.Now().UTC().UnixNano()
		last := atomic.LoadInt64(&lastTimestamp)
		if next <= last {
			next = last + 1
		}
		if atomic.CompareAndSwapInt64(&lastTimestamp, last, next) {
			return time.Unix(0, next).UTC().Format(time.RFC3339Nano)
		}
	}
}
