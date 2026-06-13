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

func (s *Service) CreateProviderConfig(ctx context.Context, input CreateProviderConfigInput) (ProviderConfig, error) {
	now := nowString()
	provider := ProviderConfig{
		ID:        newID("provider"),
		AuthorID:  input.AuthorID,
		Name:      input.Name,
		BaseURL:   input.BaseURL,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Encryption is handled in the later auth/security milestone.
	if err := s.queries.CreateProviderConfig(ctx, store.CreateProviderConfigParams{
		ID:              provider.ID,
		AuthorID:        provider.AuthorID,
		Name:            provider.Name,
		BaseUrl:         provider.BaseURL,
		ApiKeyEncrypted: input.APIKey,
		CreatedAt:       provider.CreatedAt,
		UpdatedAt:       provider.UpdatedAt,
	}); err != nil {
		return ProviderConfig{}, fmt.Errorf("create provider config: %w", err)
	}

	return provider, nil
}

func (s *Service) ListProviderConfigs(ctx context.Context, authorID string) ([]ProviderConfig, error) {
	providers, err := s.queries.ListProviderConfigs(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("list provider configs: %w", err)
	}

	result := make([]ProviderConfig, 0, len(providers))
	for _, provider := range providers {
		result = append(result, providerConfigFromStore(provider))
	}
	return result, nil
}

func (s *Service) GetProviderConfig(ctx context.Context, authorID, providerID string) (ProviderConfig, error) {
	provider, err := s.queries.GetProviderConfig(ctx, store.GetProviderConfigParams{
		ID:       providerID,
		AuthorID: authorID,
	})
	if err != nil {
		return ProviderConfig{}, fmt.Errorf("get provider config: %w", err)
	}
	return providerConfigFromStore(provider), nil
}

func (s *Service) CreateModelVariant(ctx context.Context, input CreateModelVariantInput) (ModelVariant, error) {
	if _, err := s.GetProviderConfig(ctx, input.AuthorID, input.ProviderConfigID); err != nil {
		return ModelVariant{}, err
	}

	now := nowString()
	model := ModelVariant{
		ID:                        newID("model"),
		ProviderConfigID:          input.ProviderConfigID,
		Name:                      input.Name,
		Model:                     input.Model,
		Temperature:               input.Temperature,
		MaxOutputTokens:           input.MaxOutputTokens,
		ContextWindowTokens:       input.ContextWindowTokens,
		InputPricePerMillion:      input.InputPricePerMillion,
		OutputPricePerMillion:     input.OutputPricePerMillion,
		CacheReadPricePerMillion:  input.CacheReadPricePerMillion,
		CacheWritePricePerMillion: input.CacheWritePricePerMillion,
		RequestTokenField:         defaultRequestTokenField(input.RequestTokenField),
		ReasoningFormat:           input.ReasoningFormat,
		CompatibilityJSON:         defaultJSON(input.CompatibilityJSON),
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}

	if err := s.queries.CreateModelVariant(ctx, store.CreateModelVariantParams{
		ID:                        model.ID,
		ProviderConfigID:          model.ProviderConfigID,
		Name:                      model.Name,
		Model:                     model.Model,
		Temperature:               model.Temperature,
		MaxOutputTokens:           model.MaxOutputTokens,
		ContextWindowTokens:       model.ContextWindowTokens,
		InputPricePerMillion:      model.InputPricePerMillion,
		OutputPricePerMillion:     model.OutputPricePerMillion,
		CacheReadPricePerMillion:  model.CacheReadPricePerMillion,
		CacheWritePricePerMillion: model.CacheWritePricePerMillion,
		RequestTokenField:         model.RequestTokenField,
		ReasoningFormat:           model.ReasoningFormat,
		CompatibilityJson:         model.CompatibilityJSON,
		CreatedAt:                 model.CreatedAt,
		UpdatedAt:                 model.UpdatedAt,
	}); err != nil {
		return ModelVariant{}, fmt.Errorf("create model variant: %w", err)
	}

	return model, nil
}

func (s *Service) ListModelVariantsByProvider(ctx context.Context, authorID, providerID string) ([]ModelVariant, error) {
	models, err := s.queries.ListModelVariantsByProvider(ctx, store.ListModelVariantsByProviderParams{
		ProviderConfigID: providerID,
		AuthorID:         authorID,
	})
	if err != nil {
		return nil, fmt.Errorf("list model variants by provider: %w", err)
	}

	result := make([]ModelVariant, 0, len(models))
	for _, model := range models {
		result = append(result, modelVariantFromStore(model))
	}
	return result, nil
}

func (s *Service) ListModelVariants(ctx context.Context, authorID string) ([]ModelVariant, error) {
	models, err := s.queries.ListModelVariantsByAuthor(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("list model variants: %w", err)
	}

	result := make([]ModelVariant, 0, len(models))
	for _, model := range models {
		result = append(result, modelVariantFromStore(model))
	}
	return result, nil
}

func (s *Service) GetModelVariant(ctx context.Context, authorID, modelID string) (ModelVariant, error) {
	model, err := s.queries.GetModelVariant(ctx, store.GetModelVariantParams{
		ID:       modelID,
		AuthorID: authorID,
	})
	if err != nil {
		return ModelVariant{}, fmt.Errorf("get model variant: %w", err)
	}
	return modelVariantFromStore(model), nil
}

func (s *Service) GetModelVariantForProject(ctx context.Context, projectID, modelID string) (ModelVariant, error) {
	model, err := s.queries.GetModelVariantForProject(ctx, store.GetModelVariantForProjectParams{
		ProjectID:      projectID,
		ModelVariantID: modelID,
	})
	if err != nil {
		return ModelVariant{}, fmt.Errorf("get model variant for project: %w", err)
	}
	return modelVariantFromStore(model), nil
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

func providerConfigFromStore(provider store.ProviderConfig) ProviderConfig {
	return ProviderConfig{
		ID:        provider.ID,
		AuthorID:  provider.AuthorID,
		Name:      provider.Name,
		BaseURL:   provider.BaseUrl,
		CreatedAt: provider.CreatedAt,
		UpdatedAt: provider.UpdatedAt,
	}
}

func modelVariantFromStore(model store.ModelVariant) ModelVariant {
	return ModelVariant{
		ID:                        model.ID,
		ProviderConfigID:          model.ProviderConfigID,
		Name:                      model.Name,
		Model:                     model.Model,
		Temperature:               model.Temperature,
		MaxOutputTokens:           model.MaxOutputTokens,
		ContextWindowTokens:       model.ContextWindowTokens,
		InputPricePerMillion:      model.InputPricePerMillion,
		OutputPricePerMillion:     model.OutputPricePerMillion,
		CacheReadPricePerMillion:  model.CacheReadPricePerMillion,
		CacheWritePricePerMillion: model.CacheWritePricePerMillion,
		RequestTokenField:         model.RequestTokenField,
		ReasoningFormat:           model.ReasoningFormat,
		CompatibilityJSON:         model.CompatibilityJson,
		CreatedAt:                 model.CreatedAt,
		UpdatedAt:                 model.UpdatedAt,
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

func defaultRequestTokenField(value string) string {
	if value == "" {
		return "max_tokens"
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
