package agent

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"git.inkyquill.net/inky/writer/project"
	"git.inkyquill.net/inky/writer/store"
)

const maxToolRounds = 4

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

func (s *Service) GetPromptProfile(ctx context.Context, projectID string) (PromptProfile, error) {
	profile, err := s.queries.GetPromptProfile(ctx, projectID)
	if err != nil {
		return PromptProfile{}, fmt.Errorf("get prompt profile: %w", err)
	}
	return promptProfileFromStore(profile), nil
}

func (s *Service) UpsertPromptProfile(ctx context.Context, input UpsertPromptProfileInput) (PromptProfile, error) {
	if _, err := s.queries.GetStoryProjectByID(ctx, input.ProjectID); err != nil {
		return PromptProfile{}, fmt.Errorf("get story project: %w", err)
	}

	now := nowString()
	existing, err := s.queries.GetPromptProfile(ctx, input.ProjectID)
	id := newID("profile")
	createdAt := now
	if err == nil {
		id = existing.ID
		createdAt = existing.CreatedAt
	} else if !errors.Is(err, sql.ErrNoRows) {
		return PromptProfile{}, fmt.Errorf("get prompt profile: %w", err)
	}

	if err := s.queries.UpsertPromptProfile(ctx, store.UpsertPromptProfileParams{
		ID:                        id,
		ProjectID:                 input.ProjectID,
		Genre:                     input.Genre,
		Tense:                     input.Tense,
		Pov:                       input.POV,
		Voice:                     input.Voice,
		InstructionsMarkdown:      input.InstructionsMarkdown,
		PromptRecordRetentionDays: input.PromptRecordRetentionDays,
		CreatedAt:                 createdAt,
		UpdatedAt:                 now,
	}); err != nil {
		return PromptProfile{}, fmt.Errorf("upsert prompt profile: %w", err)
	}

	profile, err := s.queries.GetPromptProfile(ctx, input.ProjectID)
	if err != nil {
		return PromptProfile{}, fmt.Errorf("get prompt profile: %w", err)
	}
	return promptProfileFromStore(profile), nil
}

func (s *Service) RunChatTurn(ctx context.Context, input ChatTurnInput) (ChatTurnResult, error) {
	if strings.TrimSpace(input.BodyMarkdown) == "" {
		return ChatTurnResult{}, fmt.Errorf("body markdown is required")
	}
	if s.provider == nil {
		return ChatTurnResult{}, fmt.Errorf("provider is required")
	}
	session, err := s.GetSession(ctx, input.ProjectID, input.SessionID)
	if err != nil {
		return ChatTurnResult{}, err
	}
	if session.ActionKind != ActionKindChat {
		return ChatTurnResult{}, fmt.Errorf("session action kind = %q, want %q", session.ActionKind, ActionKindChat)
	}
	if session.ModelVariantID == "" {
		return ChatTurnResult{}, fmt.Errorf("session model variant ID is required")
	}
	model, err := s.queries.GetModelVariantForProject(ctx, store.GetModelVariantForProjectParams{
		ProjectID:      input.ProjectID,
		ModelVariantID: session.ModelVariantID,
	})
	if err != nil {
		return ChatTurnResult{}, fmt.Errorf("get model variant for project: %w", err)
	}
	providerConfig, err := s.providerConfigForModel(ctx, model.ProviderConfigID)
	if err != nil {
		return ChatTurnResult{}, err
	}
	profile, err := s.GetPromptProfile(ctx, input.ProjectID)
	if err != nil {
		return ChatTurnResult{}, err
	}

	userMessage, err := s.AppendMessage(ctx, AppendMessageInput{
		ProjectID:    input.ProjectID,
		SessionID:    input.SessionID,
		Role:         MessageRoleUser,
		BodyMarkdown: input.BodyMarkdown,
	})
	if err != nil {
		return ChatTurnResult{}, err
	}
	transcript, err := s.ListMessages(ctx, input.ProjectID, input.SessionID)
	if err != nil {
		return ChatTurnResult{}, err
	}

	contextSources, err := chatContextSources(profile, transcript)
	if err != nil {
		return ChatTurnResult{}, err
	}
	messages := chatPromptMessages(profile, transcript)
	tools := ContextToolDefinitions()
	requests := make([]CompletionRequest, 0, 2)
	responses := make([]CompletionResponse, 0, 2)
	usage := Usage{}
	usageMissing := false
	var final CompletionResponse

	for round := 0; round <= maxToolRounds; round++ {
		request := CompletionRequest{
			Model:           model.Model,
			Messages:        messages,
			Tools:           tools,
			Temperature:     &model.Temperature,
			MaxOutputTokens: model.MaxOutputTokens,
		}
		response, err := s.provider.Complete(ctx, request)
		if err != nil {
			return ChatTurnResult{}, fmt.Errorf("complete chat turn: %w", err)
		}
		requests = append(requests, request)
		responses = append(responses, response)
		if response.UsageAvailable {
			usage = addUsage(usage, response.Usage)
		} else {
			usageMissing = true
			usage = Usage{}
		}

		if len(response.Message.ToolCalls) == 0 {
			final = response
			break
		}
		if round == maxToolRounds {
			return ChatTurnResult{}, fmt.Errorf("tool call rounds exceeded %d", maxToolRounds)
		}
		messages = append(messages, response.Message)
		for _, toolCall := range response.Message.ToolCalls {
			toolResult, err := s.ExecuteTool(ctx, ToolCallInput{
				ProjectID:     input.ProjectID,
				SessionID:     input.SessionID,
				ToolCallID:    toolCall.ID,
				ToolName:      toolCall.Function.Name,
				ArgumentsJSON: toolCall.Function.Arguments,
			})
			if err != nil {
				return ChatTurnResult{}, err
			}
			toolMetadata, err := marshalJSON(map[string]any{
				"toolCallId": toolResult.ToolCallID,
				"toolName":   toolResult.ToolName,
				"truncated":  toolResult.Truncated,
			})
			if err != nil {
				return ChatTurnResult{}, err
			}
			if _, err := s.AppendMessage(ctx, AppendMessageInput{
				ProjectID:    input.ProjectID,
				SessionID:    input.SessionID,
				Role:         MessageRoleTool,
				BodyMarkdown: toolResult.ModelVisibleMarkdown,
				MetadataJSON: toolMetadata,
			}); err != nil {
				return ChatTurnResult{}, err
			}
			messages = append(messages, CompletionMessage{
				Role:       MessageRoleTool,
				Content:    toolResult.ModelVisibleMarkdown,
				ToolCallID: toolResult.ToolCallID,
			})
		}
	}

	if final.Message.Content == "" {
		return ChatTurnResult{}, fmt.Errorf("provider did not return a final assistant message")
	}
	assistantMetadata, err := marshalJSON(map[string]any{
		"finishReason": final.FinishReason,
		"completionId": final.ID,
	})
	if err != nil {
		return ChatTurnResult{}, err
	}
	assistantMessage, err := s.AppendMessage(ctx, AppendMessageInput{
		ProjectID:    input.ProjectID,
		SessionID:    input.SessionID,
		Role:         MessageRoleAssistant,
		BodyMarkdown: final.Message.Content,
		MetadataJSON: assistantMetadata,
	})
	if err != nil {
		return ChatTurnResult{}, err
	}

	promptRecordID := ""
	if profile.PromptRecordRetentionDays > 0 {
		if usageMissing {
			usage = Usage{}
		}
		requestJSON, err := marshalJSON(map[string]any{
			"providerName": providerConfig.Name,
			"modelName":    model.Model,
			"requests":     requests,
		})
		if err != nil {
			return ChatTurnResult{}, fmt.Errorf("marshal prompt request: %w", err)
		}
		responseJSON, err := marshalJSON(map[string]any{
			"providerName": providerConfig.Name,
			"modelName":    model.Model,
			"responses":    responses,
		})
		if err != nil {
			return ChatTurnResult{}, fmt.Errorf("marshal prompt response: %w", err)
		}
		promptRecordID, err = s.storePromptRecord(ctx, session, providerConfig.Name, modelVariantFromStore(model), usage, requestJSON, responseJSON, contextSources)
		if err != nil {
			return ChatTurnResult{}, err
		}
	}
	if usageMissing {
		if err := s.createActivityEvent(ctx, input.ProjectID, input.SessionID, "usage_missing", "Provider response omitted token usage", "{}"); err != nil {
			return ChatTurnResult{}, err
		}
	}

	return ChatTurnResult{
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
		PromptRecordID:   promptRecordID,
	}, nil
}

func (s *Service) PrunePromptRecords(ctx context.Context, projectID string) (int64, error) {
	profile, err := s.GetPromptProfile(ctx, projectID)
	if err != nil {
		return 0, err
	}
	if profile.PromptRecordRetentionDays == 0 {
		result, err := s.db.ExecContext(ctx, `DELETE FROM prompt_records WHERE project_id = ?`, projectID)
		if err != nil {
			return 0, fmt.Errorf("delete project prompt records: %w", err)
		}
		return result.RowsAffected()
	}
	cutoff := nowString()
	if profile.PromptRecordRetentionDays > 0 {
		cutoff = time.Now().UTC().Add(-time.Duration(profile.PromptRecordRetentionDays) * 24 * time.Hour).Format(time.RFC3339Nano)
	}
	deleted, err := s.queries.DeleteExpiredPromptRecords(ctx, store.DeleteExpiredPromptRecordsParams{
		ProjectID: projectID,
		Cutoff:    cutoff,
	})
	if err != nil {
		return 0, fmt.Errorf("delete expired prompt records: %w", err)
	}
	return deleted, nil
}

func (s *Service) providerConfigForModel(ctx context.Context, providerConfigID string) (store.ProviderConfig, error) {
	var provider store.ProviderConfig
	err := s.db.QueryRowContext(ctx, `
		SELECT id, author_id, name, base_url, api_key_encrypted, created_at, updated_at
		FROM provider_configs
		WHERE id = ?
	`, providerConfigID).Scan(
		&provider.ID,
		&provider.AuthorID,
		&provider.Name,
		&provider.BaseUrl,
		&provider.ApiKeyEncrypted,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)
	if err != nil {
		return store.ProviderConfig{}, fmt.Errorf("get provider config for model: %w", err)
	}
	return provider, nil
}

func chatPromptMessages(profile PromptProfile, transcript []Message) []CompletionMessage {
	messages := []CompletionMessage{
		{Role: MessageRoleSystem, Content: fictionAssistantSystemPrompt},
		{Role: MessageRoleSystem, Content: renderPromptProfile(profile)},
	}
	for _, message := range transcript {
		if message.Role == MessageRoleSystem {
			continue
		}
		messages = append(messages, CompletionMessage{
			Role:    message.Role,
			Content: message.BodyMarkdown,
		})
	}
	return messages
}

func chatContextSources(profile PromptProfile, transcript []Message) ([]ContextSourceSnapshot, error) {
	profileValue, err := marshalJSON(map[string]any{
		"id":                        profile.ID,
		"projectId":                 profile.ProjectID,
		"genre":                     profile.Genre,
		"tense":                     profile.Tense,
		"pov":                       profile.POV,
		"voice":                     profile.Voice,
		"instructionsMarkdown":      profile.InstructionsMarkdown,
		"promptRecordRetentionDays": profile.PromptRecordRetentionDays,
	})
	if err != nil {
		return nil, err
	}
	transcriptValue, err := marshalJSON(transcript)
	if err != nil {
		return nil, err
	}
	toolValue, err := marshalJSON(map[string]any{"tools": ContextToolDefinitions()})
	if err != nil {
		return nil, err
	}
	providerValue, err := marshalJSON(map[string]any{
		"wholeProjectInPrompt": false,
		"instruction":          "Use tools for additional project context instead of assuming the whole project is present in this prompt.",
	})
	if err != nil {
		return nil, err
	}
	return []ContextSourceSnapshot{
		{
			SourceKey:        "prompt_profile",
			SourceVersion:    profile.UpdatedAt,
			RenderedMarkdown: renderPromptProfile(profile),
			ValueJSON:        profileValue,
		},
		{
			SourceKey:        "transcript",
			SourceVersion:    transcriptVersion(transcript),
			RenderedMarkdown: renderTranscript(transcript),
			ValueJSON:        transcriptValue,
		},
		{
			SourceKey:        "provider_disclosure",
			RenderedMarkdown: "## Provider Disclosure\n\nUse tools for additional project context instead of assuming the whole project is present in this prompt.",
			ValueJSON:        providerValue,
		},
		{
			SourceKey:        "tool_catalog",
			RenderedMarkdown: "## Tool Catalog\n\nProject context tools are available for chat turns.",
			ValueJSON:        toolValue,
		},
	}, nil
}

func renderTranscript(messages []Message) string {
	var b strings.Builder
	b.WriteString("## Transcript\n\n")
	for _, message := range messages {
		b.WriteString("### ")
		b.WriteString(string(message.Role))
		b.WriteString("\n\n")
		b.WriteString(message.BodyMarkdown)
		b.WriteString("\n\n")
	}
	return b.String()
}

func transcriptVersion(messages []Message) string {
	if len(messages) == 0 {
		return ""
	}
	return messages[len(messages)-1].CreatedAt
}

func addUsage(left Usage, right Usage) Usage {
	result := Usage{
		InputTokens:      left.InputTokens + right.InputTokens,
		OutputTokens:     left.OutputTokens + right.OutputTokens,
		CacheReadTokens:  left.CacheReadTokens + right.CacheReadTokens,
		CacheWriteTokens: left.CacheWriteTokens + right.CacheWriteTokens,
		TotalTokens:      left.TotalTokens + right.TotalTokens,
		InputCost:        left.InputCost + right.InputCost,
		OutputCost:       left.OutputCost + right.OutputCost,
		CacheReadCost:    left.CacheReadCost + right.CacheReadCost,
		CacheWriteCost:   left.CacheWriteCost + right.CacheWriteCost,
	}
	result.TotalCost = result.InputCost + result.OutputCost + result.CacheReadCost + result.CacheWriteCost
	return result
}

func usageWithCosts(usage Usage, model ModelVariant) Usage {
	usage.InputCost = tokenCost(usage.InputTokens, model.InputPricePerMillion)
	usage.OutputCost = tokenCost(usage.OutputTokens, model.OutputPricePerMillion)
	usage.CacheReadCost = tokenCost(usage.CacheReadTokens, model.CacheReadPricePerMillion)
	usage.CacheWriteCost = tokenCost(usage.CacheWriteTokens, model.CacheWritePricePerMillion)
	usage.TotalCost = usage.InputCost + usage.OutputCost + usage.CacheReadCost + usage.CacheWriteCost
	return usage
}

func (s *Service) storePromptRecord(ctx context.Context, session Session, providerName string, model ModelVariant, usage Usage, requestJSON, responseJSON string, snapshots []ContextSourceSnapshot) (string, error) {
	usage = usageWithCosts(usage, model)
	now := nowString()
	promptRecordID := newID("prompt")
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin prompt record transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	queries := s.queries.WithTx(tx)
	if err := queries.CreatePromptRecord(ctx, store.CreatePromptRecordParams{
		ID:               promptRecordID,
		ProjectID:        session.ProjectID,
		SessionID:        sql.NullString{String: session.ID, Valid: session.ID != ""},
		ProviderName:     providerName,
		ModelName:        model.Model,
		ActionKind:       string(session.ActionKind),
		RequestJson:      requestJSON,
		ResponseJson:     responseJSON,
		InputTokens:      usage.InputTokens,
		OutputTokens:     usage.OutputTokens,
		CacheReadTokens:  usage.CacheReadTokens,
		CacheWriteTokens: usage.CacheWriteTokens,
		TotalTokens:      usage.TotalTokens,
		InputCost:        usage.InputCost,
		OutputCost:       usage.OutputCost,
		CacheReadCost:    usage.CacheReadCost,
		CacheWriteCost:   usage.CacheWriteCost,
		TotalCost:        usage.TotalCost,
		CreatedAt:        now,
	}); err != nil {
		return "", fmt.Errorf("create prompt record: %w", err)
	}
	for _, snapshot := range snapshots {
		if err := queries.CreatePromptContextSnapshot(ctx, store.CreatePromptContextSnapshotParams{
			ID:               newID("snapshot"),
			PromptRecordID:   promptRecordID,
			SourceKey:        snapshot.SourceKey,
			SourceVersion:    snapshot.SourceVersion,
			RenderedMarkdown: snapshot.RenderedMarkdown,
			ValueJson:        snapshot.ValueJSON,
			CreatedAt:        now,
		}); err != nil {
			return "", fmt.Errorf("create prompt context snapshot: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit prompt record transaction: %w", err)
	}
	committed = true
	return promptRecordID, nil
}

func (s *Service) createActivityEvent(ctx context.Context, projectID, sessionID, eventType, summary, metadataJSON string) error {
	if err := s.queries.CreateActivityEvent(ctx, store.CreateActivityEventParams{
		ID:           newID("event"),
		ProjectID:    projectID,
		SessionID:    sql.NullString{String: sessionID, Valid: sessionID != ""},
		EventType:    eventType,
		Summary:      summary,
		MetadataJson: defaultJSON(metadataJSON),
		CreatedAt:    nowString(),
	}); err != nil {
		return fmt.Errorf("create activity event: %w", err)
	}
	return nil
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

func promptProfileFromStore(profile store.PromptProfile) PromptProfile {
	return PromptProfile{
		ID:                        profile.ID,
		ProjectID:                 profile.ProjectID,
		Genre:                     profile.Genre,
		Tense:                     profile.Tense,
		POV:                       profile.Pov,
		Voice:                     profile.Voice,
		InstructionsMarkdown:      profile.InstructionsMarkdown,
		PromptRecordRetentionDays: profile.PromptRecordRetentionDays,
		CreatedAt:                 profile.CreatedAt,
		UpdatedAt:                 profile.UpdatedAt,
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
