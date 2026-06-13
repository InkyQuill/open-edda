package agent

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

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
	providerConfig, err := s.queries.GetProviderConfigForProjectModel(ctx, store.GetProviderConfigForProjectModelParams{
		ProjectID:      input.ProjectID,
		ModelVariantID: session.ModelVariantID,
	})
	if err != nil {
		return ChatTurnResult{}, fmt.Errorf("get provider config for project model: %w", err)
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
			s.recordPostResponseFailure(ctx, input.ProjectID, input.SessionID, "prompt_record_failed", "Failed to marshal prompt record request", err)
		} else {
			responseJSON, err := marshalJSON(map[string]any{
				"providerName": providerConfig.Name,
				"modelName":    model.Model,
				"responses":    responses,
			})
			if err != nil {
				s.recordPostResponseFailure(ctx, input.ProjectID, input.SessionID, "prompt_record_failed", "Failed to marshal prompt record response", err)
			} else {
				promptRecordID, err = s.storePromptRecord(ctx, session, providerConfig.Name, modelVariantFromStore(model), usage, requestJSON, responseJSON, contextSources)
				if err != nil {
					promptRecordID = ""
					s.recordPostResponseFailure(ctx, input.ProjectID, input.SessionID, "prompt_record_failed", "Failed to store prompt record", err)
				}
			}
		}
	}
	if usageMissing {
		if err := s.createActivityEvent(ctx, input.ProjectID, input.SessionID, "usage_missing", "Provider response omitted token usage", "{}"); err != nil {
			s.recordPostResponseFailure(ctx, input.ProjectID, input.SessionID, "usage_missing_failed", "Failed to store missing-usage event", err)
		}
	}

	return ChatTurnResult{
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
		PromptRecordID:   promptRecordID,
	}, nil
}

func (s *Service) RunContinuation(ctx context.Context, input ContinuationInput) (ContinuationResult, error) {
	if input.ApplyMode == "" {
		input.ApplyMode = ApplyModePreview
	}
	if err := validateApplyMode(input.ApplyMode); err != nil {
		return ContinuationResult{}, err
	}
	generated, session, err := s.runQuickActionCompletion(ctx, quickActionCompletionInput{
		ProjectID:        input.ProjectID,
		ContentID:        input.ContentID,
		ModelVariantID:   input.ModelVariantID,
		ApplyMode:        input.ApplyMode,
		ActionKind:       ActionKindContinuation,
		ExpectedRevision: input.ExpectedRevision,
		SelectionStart:   0,
		SelectionEnd:     0,
		Guidance:         input.Guidance,
		Instruction:      continuationInstruction(input),
	})
	if err != nil {
		return ContinuationResult{}, err
	}

	operationKind := "append"
	if input.Insert || input.InsertPosition > 0 {
		operationKind = "insert"
	}
	reason := quickActionReason(ActionKindContinuation, input.Guidance)
	if input.ApplyMode == ApplyModePreview {
		candidate, err := s.createGenerationCandidate(ctx, generationCandidateInput{
			ProjectID:         input.ProjectID,
			SessionID:         session.ID,
			ContentID:         input.ContentID,
			ActionKind:        ActionKindContinuation,
			OperationKind:     operationKind,
			ExpectedRevision:  input.ExpectedRevision,
			InsertPosition:    input.InsertPosition,
			HasInsertPosition: operationKind == "insert",
			GeneratedMarkdown: generated,
			Reason:            reason,
			ModelVariantID:    input.ModelVariantID,
		})
		if err != nil {
			return ContinuationResult{}, err
		}
		if err := s.createActivityEvent(ctx, input.ProjectID, session.ID, "generation_candidate_created", "Created continuation preview", "{}"); err != nil {
			return ContinuationResult{}, err
		}
		return ContinuationResult{Session: session, Candidate: candidate}, nil
	}

	writeInput := project.StructuredWriteInput{
		ProjectID:         input.ProjectID,
		ContentID:         input.ContentID,
		ExpectedRevision:  input.ExpectedRevision,
		GeneratedMarkdown: generated,
		InsertPosition:    input.InsertPosition,
		Reason:            reason,
		AgentSessionID:    session.ID,
		ActionKind:        string(ActionKindContinuation),
		ModelVariantID:    input.ModelVariantID,
	}
	content, err := s.applyCandidateWrite(ctx, operationKind, writeInput)
	if err != nil {
		return ContinuationResult{}, err
	}
	if err := s.createActivityEvent(ctx, input.ProjectID, session.ID, "quick_action_applied", "Applied continuation", "{}"); err != nil {
		return ContinuationResult{}, err
	}
	return ContinuationResult{Session: session, Content: content}, nil
}

func (s *Service) RunRewrite(ctx context.Context, input RewriteInput) (RewriteResult, error) {
	if input.ApplyMode == "" {
		input.ApplyMode = ApplyModePreview
	}
	if err := validateApplyMode(input.ApplyMode); err != nil {
		return RewriteResult{}, err
	}
	chapter, original, before, after, err := s.selectedChapterContext(ctx, input.ProjectID, input.ContentID, input.SelectionStart, input.SelectionEnd)
	if err != nil {
		return RewriteResult{}, err
	}
	generated, session, err := s.runQuickActionCompletion(ctx, quickActionCompletionInput{
		ProjectID:        input.ProjectID,
		ContentID:        input.ContentID,
		ModelVariantID:   input.ModelVariantID,
		ApplyMode:        input.ApplyMode,
		ActionKind:       ActionKindRewrite,
		ExpectedRevision: input.ExpectedRevision,
		SelectionStart:   input.SelectionStart,
		SelectionEnd:     input.SelectionEnd,
		Guidance:         input.Guidance,
		Instruction:      rewriteInstruction(input, original, before, after),
	})
	if err != nil {
		return RewriteResult{}, err
	}
	reasonPayload, err := marshalJSON(map[string]any{
		"action":           "rewrite",
		"contentTitle":     chapter.Title,
		"guidance":         input.Guidance,
		"selectionStart":   input.SelectionStart,
		"selectionEnd":     input.SelectionEnd,
		"before":           before,
		"original":         original,
		"generated":        generated,
		"after":            after,
		"expectedRevision": input.ExpectedRevision,
	})
	if err != nil {
		return RewriteResult{}, err
	}
	if input.ApplyMode == ApplyModePreview {
		candidate, err := s.createGenerationCandidate(ctx, generationCandidateInput{
			ProjectID:         input.ProjectID,
			SessionID:         session.ID,
			ContentID:         input.ContentID,
			ActionKind:        ActionKindRewrite,
			OperationKind:     "replace",
			ExpectedRevision:  input.ExpectedRevision,
			SelectionStart:    input.SelectionStart,
			SelectionEnd:      input.SelectionEnd,
			OriginalMarkdown:  original,
			GeneratedMarkdown: generated,
			Reason:            reasonPayload,
			ModelVariantID:    input.ModelVariantID,
		})
		if err != nil {
			return RewriteResult{}, err
		}
		if err := s.createActivityEvent(ctx, input.ProjectID, session.ID, "generation_candidate_created", "Created rewrite preview", "{}"); err != nil {
			return RewriteResult{}, err
		}
		return RewriteResult{Session: session, Candidate: candidate}, nil
	}

	content, err := s.applyCandidateWrite(ctx, "replace", project.StructuredWriteInput{
		ProjectID:         input.ProjectID,
		ContentID:         input.ContentID,
		ExpectedRevision:  input.ExpectedRevision,
		GeneratedMarkdown: generated,
		SelectionStart:    input.SelectionStart,
		SelectionEnd:      input.SelectionEnd,
		Reason:            reasonPayload,
		AgentSessionID:    session.ID,
		ActionKind:        string(ActionKindRewrite),
		ModelVariantID:    input.ModelVariantID,
	})
	if err != nil {
		return RewriteResult{}, err
	}
	if err := s.createActivityEvent(ctx, input.ProjectID, session.ID, "quick_action_applied", "Applied rewrite", "{}"); err != nil {
		return RewriteResult{}, err
	}
	return RewriteResult{Session: session, Content: content}, nil
}

func (s *Service) AcceptCandidate(ctx context.Context, input AcceptCandidateInput) (AcceptCandidateResult, error) {
	candidate, err := s.queries.GetGenerationCandidate(ctx, store.GetGenerationCandidateParams{
		ProjectID: input.ProjectID,
		ID:        input.CandidateID,
	})
	if err != nil {
		return AcceptCandidateResult{}, fmt.Errorf("get generation candidate: %w", err)
	}
	if candidate.Status != "pending" {
		return AcceptCandidateResult{}, fmt.Errorf("candidate status = %q, want pending: %w", candidate.Status, project.ErrConflict)
	}
	if err := s.transitionCandidateStatus(ctx, input.ProjectID, input.CandidateID, "pending", "applying"); err != nil {
		return AcceptCandidateResult{}, err
	}

	content, err := s.applyCandidateWrite(ctx, candidate.OperationKind, project.StructuredWriteInput{
		ProjectID:         candidate.ProjectID,
		ContentID:         candidate.ContentItemID,
		ExpectedRevision:  candidate.ExpectedRevision,
		GeneratedMarkdown: candidate.GeneratedMarkdown,
		InsertPosition:    valueInt64(candidate.InsertPosition),
		SelectionStart:    valueInt64(candidate.SelectionStart),
		SelectionEnd:      valueInt64(candidate.SelectionEnd),
		Reason:            candidate.Reason,
		AgentSessionID:    candidate.SessionID,
		ActionKind:        candidate.ActionKind,
		ModelVariantID:    valueString(candidate.ModelVariantID),
	})
	if err != nil {
		if errors.Is(err, project.ErrConflict) {
			if updateErr := s.transitionCandidateStatus(ctx, input.ProjectID, input.CandidateID, "applying", "conflict"); updateErr != nil {
				return AcceptCandidateResult{}, updateErr
			}
			if eventErr := s.createActivityEvent(ctx, candidate.ProjectID, candidate.SessionID, "generation_candidate_conflict", "Generation candidate conflicted", "{}"); eventErr != nil {
				return AcceptCandidateResult{}, eventErr
			}
		} else if updateErr := s.transitionCandidateStatus(ctx, input.ProjectID, input.CandidateID, "applying", "pending"); updateErr != nil {
			return AcceptCandidateResult{}, errors.Join(err, updateErr)
		}
		return AcceptCandidateResult{}, err
	}
	if err := s.transitionCandidateStatus(ctx, input.ProjectID, input.CandidateID, "applying", "accepted"); err != nil {
		return AcceptCandidateResult{}, err
	}
	if err := s.createActivityEvent(ctx, candidate.ProjectID, candidate.SessionID, "generation_candidate_accepted", "Accepted generation candidate", "{}"); err != nil {
		return AcceptCandidateResult{}, err
	}
	updated, err := s.queries.GetGenerationCandidate(ctx, store.GetGenerationCandidateParams{
		ProjectID: input.ProjectID,
		ID:        input.CandidateID,
	})
	if err != nil {
		return AcceptCandidateResult{}, fmt.Errorf("get updated generation candidate: %w", err)
	}
	return AcceptCandidateResult{Candidate: generationCandidateFromStore(updated), Content: content}, nil
}

func (s *Service) RejectCandidate(ctx context.Context, input RejectCandidateInput) (GenerationCandidate, error) {
	candidate, err := s.queries.GetGenerationCandidate(ctx, store.GetGenerationCandidateParams{
		ProjectID: input.ProjectID,
		ID:        input.CandidateID,
	})
	if err != nil {
		return GenerationCandidate{}, fmt.Errorf("get generation candidate: %w", err)
	}
	if candidate.Status != "pending" {
		return GenerationCandidate{}, fmt.Errorf("candidate status = %q, want pending: %w", candidate.Status, project.ErrConflict)
	}
	if err := s.transitionCandidateStatus(ctx, input.ProjectID, input.CandidateID, "pending", "rejected"); err != nil {
		return GenerationCandidate{}, err
	}
	if err := s.createActivityEvent(ctx, candidate.ProjectID, candidate.SessionID, "generation_candidate_rejected", "Rejected generation candidate", "{}"); err != nil {
		return GenerationCandidate{}, err
	}
	updated, err := s.queries.GetGenerationCandidate(ctx, store.GetGenerationCandidateParams{
		ProjectID: input.ProjectID,
		ID:        input.CandidateID,
	})
	if err != nil {
		return GenerationCandidate{}, fmt.Errorf("get updated generation candidate: %w", err)
	}
	return generationCandidateFromStore(updated), nil
}

func (s *Service) RunReadAndCheck(ctx context.Context, input ReadAndCheckInput) (ReadAndCheckResult, error) {
	if input.ApplyMode == "" {
		input.ApplyMode = ApplyModePreview
	}
	if err := validateApplyMode(input.ApplyMode); err != nil {
		return ReadAndCheckResult{}, err
	}
	chapter, original, before, after, err := s.selectedChapterContext(ctx, input.ProjectID, input.ContentID, input.SelectionStart, input.SelectionEnd)
	if err != nil {
		return ReadAndCheckResult{}, err
	}
	report, session, err := s.runQuickActionCompletion(ctx, quickActionCompletionInput{
		ProjectID:        input.ProjectID,
		ContentID:        input.ContentID,
		ModelVariantID:   input.ModelVariantID,
		ApplyMode:        input.ApplyMode,
		ActionKind:       ActionKindReadCheck,
		ExpectedRevision: input.ExpectedRevision,
		SelectionStart:   input.SelectionStart,
		SelectionEnd:     input.SelectionEnd,
		Guidance:         input.Guidance,
		Instruction:      readAndCheckInstruction(input, original, before, after),
	})
	if err != nil {
		return ReadAndCheckResult{}, err
	}
	assistantMessage, err := s.AppendMessage(ctx, AppendMessageInput{
		ProjectID:    input.ProjectID,
		SessionID:    session.ID,
		Role:         MessageRoleAssistant,
		BodyMarkdown: report,
		MetadataJSON: `{"actionKind":"read_check"}`,
	})
	if err != nil {
		return ReadAndCheckResult{}, err
	}
	note, err := s.projectService.CreateAttachedNote(ctx, project.CreateAttachedNoteInput{
		ProjectID:      input.ProjectID,
		ContentItemID:  input.ContentID,
		SelectionStart: input.SelectionStart,
		SelectionEnd:   input.SelectionEnd,
		Title:          "Read and check: " + chapter.Title,
		BodyMarkdown:   report,
		Source:         "read_and_check",
	})
	if err != nil {
		return ReadAndCheckResult{}, err
	}
	if err := s.createActivityEvent(ctx, input.ProjectID, session.ID, "read_and_check_completed", "Stored read and check report", "{}"); err != nil {
		return ReadAndCheckResult{}, err
	}
	return ReadAndCheckResult{Session: session, AssistantMessage: assistantMessage, Note: note}, nil
}

func (s *Service) PrunePromptRecords(ctx context.Context, projectID string) (int64, error) {
	profile, err := s.GetPromptProfile(ctx, projectID)
	if err != nil {
		return 0, err
	}
	if profile.PromptRecordRetentionDays == 0 {
		deleted, err := s.queries.DeletePromptRecordsByProject(ctx, projectID)
		if err != nil {
			return 0, fmt.Errorf("delete project prompt records: %w", err)
		}
		return deleted, nil
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

type quickActionCompletionInput struct {
	ProjectID        string
	ContentID        string
	ModelVariantID   string
	ApplyMode        ApplyMode
	ActionKind       ActionKind
	ExpectedRevision int64
	SelectionStart   int64
	SelectionEnd     int64
	Guidance         string
	Instruction      string
}

func (s *Service) runQuickActionCompletion(ctx context.Context, input quickActionCompletionInput) (string, Session, error) {
	if s.provider == nil {
		return "", Session{}, fmt.Errorf("provider is required")
	}
	if input.ModelVariantID == "" {
		return "", Session{}, fmt.Errorf("model variant ID is required")
	}
	model, err := s.GetModelVariantForProject(ctx, input.ProjectID, input.ModelVariantID)
	if err != nil {
		return "", Session{}, err
	}
	if _, err := s.projectService.GetContent(ctx, input.ProjectID, input.ContentID); err != nil {
		return "", Session{}, err
	}
	session, err := s.CreateSession(ctx, CreateSessionInput{
		ProjectID:      input.ProjectID,
		Title:          quickActionTitle(input.ActionKind),
		ActionKind:     input.ActionKind,
		ModelVariantID: input.ModelVariantID,
		ApplyMode:      input.ApplyMode,
	})
	if err != nil {
		return "", Session{}, err
	}
	cursorSummary := quickActionCursorSummary(input)
	bundle, err := s.BuildActionPrompt(ctx, BuildPromptInput{
		ProjectID:       input.ProjectID,
		SessionID:       session.ID,
		ActionKind:      input.ActionKind,
		TargetContentID: input.ContentID,
		CursorSummary:   cursorSummary,
		UserGuidance:    input.Instruction,
	})
	if err != nil {
		return "", Session{}, err
	}
	request := CompletionRequest{
		Model: model.Model,
		Messages: []CompletionMessage{
			{Role: MessageRoleSystem, Content: bundle.SystemMessage},
			{Role: MessageRoleSystem, Content: bundle.DeveloperMessage},
			{Role: MessageRoleUser, Content: bundle.UserMessage},
		},
		Temperature:     &model.Temperature,
		MaxOutputTokens: model.MaxOutputTokens,
	}
	response, err := s.provider.Complete(ctx, request)
	if err != nil {
		return "", Session{}, fmt.Errorf("complete quick action: %w", err)
	}
	generated := response.Message.Content
	if strings.TrimSpace(generated) == "" {
		return "", Session{}, fmt.Errorf("provider did not return generated markdown")
	}
	return generated, session, nil
}

type generationCandidateInput struct {
	ProjectID         string
	SessionID         string
	ContentID         string
	ActionKind        ActionKind
	OperationKind     string
	ExpectedRevision  int64
	SelectionStart    int64
	SelectionEnd      int64
	InsertPosition    int64
	HasInsertPosition bool
	OriginalMarkdown  string
	GeneratedMarkdown string
	Reason            string
	ModelVariantID    string
}

func (s *Service) createGenerationCandidate(ctx context.Context, input generationCandidateInput) (GenerationCandidate, error) {
	now := nowString()
	selectionStart := int64Ptr(input.SelectionStart, input.SelectionEnd > input.SelectionStart)
	selectionEnd := int64Ptr(input.SelectionEnd, input.SelectionEnd > input.SelectionStart)
	insertPosition := int64Ptr(input.InsertPosition, input.HasInsertPosition)
	candidate := GenerationCandidate{
		ID:                newID("candidate"),
		ProjectID:         input.ProjectID,
		SessionID:         input.SessionID,
		ContentItemID:     input.ContentID,
		ActionKind:        input.ActionKind,
		OperationKind:     input.OperationKind,
		ExpectedRevision:  input.ExpectedRevision,
		SelectionStart:    selectionStart,
		SelectionEnd:      selectionEnd,
		InsertPosition:    insertPosition,
		OriginalMarkdown:  input.OriginalMarkdown,
		GeneratedMarkdown: input.GeneratedMarkdown,
		Reason:            input.Reason,
		ModelVariantID:    input.ModelVariantID,
		Status:            "pending",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.queries.CreateGenerationCandidate(ctx, store.CreateGenerationCandidateParams{
		ID:                candidate.ID,
		ProjectID:         candidate.ProjectID,
		SessionID:         candidate.SessionID,
		ContentItemID:     candidate.ContentItemID,
		ActionKind:        string(candidate.ActionKind),
		OperationKind:     candidate.OperationKind,
		ExpectedRevision:  candidate.ExpectedRevision,
		SelectionStart:    nullInt64(input.SelectionStart, selectionStart != nil),
		SelectionEnd:      nullInt64(input.SelectionEnd, selectionEnd != nil),
		InsertPosition:    nullInt64(input.InsertPosition, insertPosition != nil),
		OriginalMarkdown:  candidate.OriginalMarkdown,
		GeneratedMarkdown: candidate.GeneratedMarkdown,
		Reason:            candidate.Reason,
		ModelVariantID:    nullString(candidate.ModelVariantID),
		Status:            candidate.Status,
		CreatedAt:         candidate.CreatedAt,
		UpdatedAt:         candidate.UpdatedAt,
	}); err != nil {
		return GenerationCandidate{}, fmt.Errorf("create generation candidate: %w", err)
	}
	return candidate, nil
}

func (s *Service) transitionCandidateStatus(ctx context.Context, projectID, candidateID, fromStatus, toStatus string) error {
	affected, err := s.queries.UpdateGenerationCandidateStatusIfStatus(ctx, store.UpdateGenerationCandidateStatusIfStatusParams{
		Status:         toStatus,
		UpdatedAt:      nowString(),
		ProjectID:      projectID,
		ID:             candidateID,
		ExpectedStatus: fromStatus,
	})
	if err != nil {
		return fmt.Errorf("update generation candidate status: %w", err)
	}
	if affected == 0 {
		return project.ErrConflict
	}
	return nil
}

func (s *Service) applyCandidateWrite(ctx context.Context, operationKind string, input project.StructuredWriteInput) (project.ContentItem, error) {
	switch operationKind {
	case "append":
		return s.projectService.AppendToContent(ctx, input)
	case "insert":
		return s.projectService.InsertIntoContent(ctx, input)
	case "replace":
		return s.projectService.ReplaceContentRange(ctx, input)
	default:
		return project.ContentItem{}, fmt.Errorf("unknown candidate operation %q", operationKind)
	}
}

func (s *Service) selectedChapterContext(ctx context.Context, projectID, contentID string, start, end int64) (project.ContentItem, string, string, string, error) {
	chapter, err := s.projectService.GetContent(ctx, projectID, contentID)
	if err != nil {
		return project.ContentItem{}, "", "", "", err
	}
	if chapter.Kind != project.KindChapter {
		return project.ContentItem{}, "", "", "", fmt.Errorf("target content must be chapter, got %q", chapter.Kind)
	}
	if start < 0 || end <= start || end > int64(len(chapter.BodyMarkdown)) {
		return project.ContentItem{}, "", "", "", fmt.Errorf("selection range out of range")
	}
	if !validUTF8ByteBoundary(chapter.BodyMarkdown, start) {
		return project.ContentItem{}, "", "", "", fmt.Errorf("selection start byte offset %d is not a UTF-8 rune boundary", start)
	}
	if !validUTF8ByteBoundary(chapter.BodyMarkdown, end) {
		return project.ContentItem{}, "", "", "", fmt.Errorf("selection end byte offset %d is not a UTF-8 rune boundary", end)
	}
	const contextBytes = 200
	beforeStart := start - contextBytes
	if beforeStart < 0 {
		beforeStart = 0
	}
	afterEnd := end + contextBytes
	if afterEnd > int64(len(chapter.BodyMarkdown)) {
		afterEnd = int64(len(chapter.BodyMarkdown))
	}
	return chapter,
		chapter.BodyMarkdown[start:end],
		chapter.BodyMarkdown[beforeStart:start],
		chapter.BodyMarkdown[end:afterEnd],
		nil
}

func continuationInstruction(input ContinuationInput) string {
	unit := emptyDefault(input.ContinuationUnits, "words")
	count := input.ContinuationCount
	if count <= 0 {
		count = 1
	}
	var b strings.Builder
	b.WriteString("Continue the target chapter with polished prose only. ")
	b.WriteString(fmt.Sprintf("Target length: %d %s. ", count, unit))
	if input.Insert || input.InsertPosition > 0 {
		b.WriteString(fmt.Sprintf("Insert at byte position %d. ", input.InsertPosition))
	} else {
		b.WriteString("Append after the current chapter text. ")
	}
	if input.Guidance != "" {
		b.WriteString("Author guidance: ")
		b.WriteString(input.Guidance)
	}
	return b.String()
}

func rewriteInstruction(input RewriteInput, original, before, after string) string {
	var b strings.Builder
	b.WriteString("Rewrite only the selected text. Return replacement Markdown only.\n\n")
	b.WriteString("Before selection:\n")
	b.WriteString(before)
	b.WriteString("\n\nSelected text:\n")
	b.WriteString(original)
	b.WriteString("\n\nAfter selection:\n")
	b.WriteString(after)
	if input.Guidance != "" {
		b.WriteString("\n\nAuthor guidance:\n")
		b.WriteString(input.Guidance)
	}
	return b.String()
}

func readAndCheckInstruction(input ReadAndCheckInput, original, before, after string) string {
	var b strings.Builder
	b.WriteString("Read and check the selected prose. Return a concise assistant report with issues, risks, and suggested fixes. Do not rewrite the prose.\n\n")
	b.WriteString("Before selection:\n")
	b.WriteString(before)
	b.WriteString("\n\nSelected text:\n")
	b.WriteString(original)
	b.WriteString("\n\nAfter selection:\n")
	b.WriteString(after)
	if input.Guidance != "" {
		b.WriteString("\n\nAuthor guidance:\n")
		b.WriteString(input.Guidance)
	}
	return b.String()
}

func quickActionReason(action ActionKind, guidance string) string {
	if guidance == "" {
		return string(action)
	}
	return string(action) + ": " + guidance
}

func quickActionTitle(action ActionKind) string {
	switch action {
	case ActionKindContinuation:
		return "Continuation"
	case ActionKindRewrite:
		return "Rewrite"
	case ActionKindReadCheck:
		return "Read and check"
	default:
		return "Quick action"
	}
}

func quickActionCursorSummary(input quickActionCompletionInput) string {
	switch input.ActionKind {
	case ActionKindContinuation:
		return fmt.Sprintf("Expected revision %d.", input.ExpectedRevision)
	case ActionKindRewrite, ActionKindReadCheck:
		return fmt.Sprintf("Expected revision %d; selection byte range %d-%d.", input.ExpectedRevision, input.SelectionStart, input.SelectionEnd)
	default:
		return ""
	}
}

func chatPromptMessages(profile PromptProfile, transcript []Message) []CompletionMessage {
	messages := []CompletionMessage{
		{Role: MessageRoleSystem, Content: fictionAssistantSystemPrompt},
		{Role: MessageRoleSystem, Content: renderPromptProfile(profile)},
	}
	for _, message := range transcript {
		if message.Role == MessageRoleSystem || message.Role == MessageRoleTool {
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

func (s *Service) recordPostResponseFailure(ctx context.Context, projectID, sessionID, eventType, summary string, cause error) {
	metadataJSON, err := marshalJSON(map[string]any{"error": cause.Error()})
	if err != nil {
		metadataJSON = "{}"
	}
	_ = s.createActivityEvent(ctx, projectID, sessionID, eventType, summary, metadataJSON)
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

func generationCandidateFromStore(candidate store.GenerationCandidate) GenerationCandidate {
	return GenerationCandidate{
		ID:                candidate.ID,
		ProjectID:         candidate.ProjectID,
		SessionID:         candidate.SessionID,
		ContentItemID:     candidate.ContentItemID,
		ActionKind:        ActionKind(candidate.ActionKind),
		OperationKind:     candidate.OperationKind,
		ExpectedRevision:  candidate.ExpectedRevision,
		SelectionStart:    int64Ptr(candidate.SelectionStart.Int64, candidate.SelectionStart.Valid),
		SelectionEnd:      int64Ptr(candidate.SelectionEnd.Int64, candidate.SelectionEnd.Valid),
		InsertPosition:    int64Ptr(candidate.InsertPosition.Int64, candidate.InsertPosition.Valid),
		OriginalMarkdown:  candidate.OriginalMarkdown,
		GeneratedMarkdown: candidate.GeneratedMarkdown,
		Reason:            candidate.Reason,
		ModelVariantID:    valueString(candidate.ModelVariantID),
		Status:            candidate.Status,
		CreatedAt:         candidate.CreatedAt,
		UpdatedAt:         candidate.UpdatedAt,
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

func nullInt64(value int64, valid bool) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: valid}
}

func valueInt64(value sql.NullInt64) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}

func int64Ptr(value int64, valid bool) *int64 {
	if !valid {
		return nil
	}
	return &value
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

func emptyDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func validUTF8ByteBoundary(value string, offset int64) bool {
	if offset < 0 || offset > int64(len(value)) {
		return false
	}
	if offset == 0 || offset == int64(len(value)) {
		return true
	}
	return utf8.RuneStart(value[offset])
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
