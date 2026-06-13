package agent

import "git.inkyquill.net/inky/writer/project"

type ActionKind string

const (
	ActionKindChat         ActionKind = "chat"
	ActionKindContinuation ActionKind = "continuation"
	ActionKindRewrite      ActionKind = "rewrite"
	ActionKindReadCheck    ActionKind = "read_check"
)

type ApplyMode string

const (
	ApplyModePreview     ApplyMode = "preview"
	ApplyModeDirectApply ApplyMode = "direct_apply"
)

type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleTool      MessageRole = "tool"
	MessageRoleSystem    MessageRole = "system"
)

type Session struct {
	ID             string     `json:"id"`
	ProjectID      string     `json:"projectId"`
	Title          string     `json:"title"`
	ActionKind     ActionKind `json:"actionKind"`
	ModelVariantID string     `json:"modelVariantId"`
	ApplyMode      ApplyMode  `json:"applyMode"`
	CreatedAt      string     `json:"createdAt"`
	UpdatedAt      string     `json:"updatedAt"`
}

type Message struct {
	ID           string      `json:"id"`
	SessionID    string      `json:"sessionId"`
	Role         MessageRole `json:"role"`
	BodyMarkdown string      `json:"bodyMarkdown"`
	MetadataJSON string      `json:"metadataJson"`
	CreatedAt    string      `json:"createdAt"`
}

type CreateSessionInput struct {
	ProjectID      string
	Title          string
	ActionKind     ActionKind
	ModelVariantID string
	ApplyMode      ApplyMode
}

type AppendMessageInput struct {
	ProjectID    string
	SessionID    string
	Role         MessageRole
	BodyMarkdown string
	MetadataJSON string
}

type ChatTurnInput struct {
	ProjectID    string
	SessionID    string
	BodyMarkdown string
}

type ChatTurnResult struct {
	UserMessage      Message `json:"userMessage"`
	AssistantMessage Message `json:"assistantMessage"`
	PromptRecordID   string  `json:"promptRecordId,omitempty"`
}

type ListSessionsInput struct {
	ProjectID string
	Limit     int64
}

type ProviderConfig struct {
	ID        string `json:"id"`
	AuthorID  string `json:"authorId"`
	Name      string `json:"name"`
	BaseURL   string `json:"baseUrl"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type ModelVariant struct {
	ID                        string  `json:"id"`
	ProviderConfigID          string  `json:"providerConfigId"`
	Name                      string  `json:"name"`
	Model                     string  `json:"model"`
	Temperature               float64 `json:"temperature"`
	MaxOutputTokens           int64   `json:"maxOutputTokens"`
	ContextWindowTokens       int64   `json:"contextWindowTokens"`
	InputPricePerMillion      float64 `json:"inputPricePerMillion"`
	OutputPricePerMillion     float64 `json:"outputPricePerMillion"`
	CacheReadPricePerMillion  float64 `json:"cacheReadPricePerMillion"`
	CacheWritePricePerMillion float64 `json:"cacheWritePricePerMillion"`
	RequestTokenField         string  `json:"requestTokenField"`
	ReasoningFormat           string  `json:"reasoningFormat"`
	CompatibilityJSON         string  `json:"compatibilityJson"`
	CreatedAt                 string  `json:"createdAt"`
	UpdatedAt                 string  `json:"updatedAt"`
}

type CreateProviderConfigInput struct {
	AuthorID string
	Name     string
	BaseURL  string
	APIKey   string
}

type CreateModelVariantInput struct {
	AuthorID                  string
	ProviderConfigID          string
	Name                      string
	Model                     string
	Temperature               float64
	MaxOutputTokens           int64
	ContextWindowTokens       int64
	InputPricePerMillion      float64
	OutputPricePerMillion     float64
	CacheReadPricePerMillion  float64
	CacheWritePricePerMillion float64
	RequestTokenField         string
	ReasoningFormat           string
	CompatibilityJSON         string
}

type PromptProfile struct {
	ID                        string `json:"id"`
	ProjectID                 string `json:"projectId"`
	Genre                     string `json:"genre"`
	Tense                     string `json:"tense"`
	POV                       string `json:"pov"`
	Voice                     string `json:"voice"`
	InstructionsMarkdown      string `json:"instructionsMarkdown"`
	PromptRecordRetentionDays int64  `json:"promptRecordRetentionDays"`
	CreatedAt                 string `json:"createdAt"`
	UpdatedAt                 string `json:"updatedAt"`
}

type UpsertPromptProfileInput struct {
	ProjectID                 string
	Genre                     string
	Tense                     string
	POV                       string
	Voice                     string
	InstructionsMarkdown      string
	PromptRecordRetentionDays int64
}

type BuildPromptInput struct {
	ProjectID       string
	SessionID       string
	ActionKind      ActionKind
	TargetContentID string
	CursorSummary   string
	UserGuidance    string
}

type PromptBundle struct {
	SystemMessage    string                  `json:"systemMessage"`
	DeveloperMessage string                  `json:"developerMessage"`
	UserMessage      string                  `json:"userMessage"`
	Tools            []CompletionTool        `json:"tools"`
	MetadataJSON     string                  `json:"metadataJson"`
	ContextSources   []ContextSourceSnapshot `json:"contextSources"`
}

type ContextSourceSnapshot struct {
	SourceKey        string `json:"sourceKey"`
	SourceVersion    string `json:"sourceVersion,omitempty"`
	RenderedMarkdown string `json:"renderedMarkdown"`
	ValueJSON        string `json:"valueJson"`
}

type QuickAction struct {
	ID         string     `json:"id"`
	Label      string     `json:"label"`
	ActionKind ActionKind `json:"actionKind"`
}

type StructuredWrite struct {
	ContentItemID     string `json:"contentItemId"`
	ExpectedRevision  int64  `json:"expectedRevision"`
	GeneratedMarkdown string `json:"generatedMarkdown"`
}

type ContinuationInput struct {
	ProjectID         string
	ContentID         string
	ModelVariantID    string
	ApplyMode         ApplyMode
	Guidance          string
	ExpectedRevision  int64
	InsertPosition    int64
	Insert            bool
	ContinuationUnits string
	ContinuationCount int64
}

type ContinuationResult struct {
	Session   Session             `json:"session"`
	Candidate GenerationCandidate `json:"candidate,omitempty"`
	Content   project.ContentItem `json:"content,omitempty"`
}

type RewriteInput struct {
	ProjectID        string
	ContentID        string
	ModelVariantID   string
	ApplyMode        ApplyMode
	Guidance         string
	ExpectedRevision int64
	SelectionStart   int64
	SelectionEnd     int64
}

type RewriteResult struct {
	Session   Session             `json:"session"`
	Candidate GenerationCandidate `json:"candidate,omitempty"`
	Content   project.ContentItem `json:"content,omitempty"`
}

type ReadAndCheckInput struct {
	ProjectID        string
	ContentID        string
	ModelVariantID   string
	ApplyMode        ApplyMode
	Guidance         string
	ExpectedRevision int64
	SelectionStart   int64
	SelectionEnd     int64
}

type ReadAndCheckResult struct {
	Session          Session              `json:"session"`
	AssistantMessage Message              `json:"assistantMessage"`
	Note             project.AttachedNote `json:"note"`
}

type AcceptCandidateInput struct {
	ProjectID   string
	CandidateID string
}

type AcceptCandidateResult struct {
	Candidate GenerationCandidate `json:"candidate"`
	Content   project.ContentItem `json:"content"`
}

type RejectCandidateInput struct {
	ProjectID   string
	CandidateID string
}

type ActivityEvent struct {
	ID           string `json:"id"`
	ProjectID    string `json:"projectId"`
	SessionID    string `json:"sessionId"`
	EventType    string `json:"eventType"`
	Summary      string `json:"summary"`
	MetadataJSON string `json:"metadataJson"`
	CreatedAt    string `json:"createdAt"`
}

type PromptRecord struct {
	ID           string     `json:"id"`
	ProjectID    string     `json:"projectId"`
	SessionID    string     `json:"sessionId"`
	ProviderName string     `json:"providerName"`
	ModelName    string     `json:"modelName"`
	ActionKind   ActionKind `json:"actionKind"`
	RequestJSON  string     `json:"requestJson"`
	ResponseJSON string     `json:"responseJson"`
	Usage        Usage      `json:"usage"`
	CreatedAt    string     `json:"createdAt"`
}

type Usage struct {
	InputTokens      int64   `json:"inputTokens"`
	OutputTokens     int64   `json:"outputTokens"`
	CacheReadTokens  int64   `json:"cacheReadTokens"`
	CacheWriteTokens int64   `json:"cacheWriteTokens"`
	TotalTokens      int64   `json:"totalTokens"`
	InputCost        float64 `json:"inputCost"`
	OutputCost       float64 `json:"outputCost"`
	CacheReadCost    float64 `json:"cacheReadCost"`
	CacheWriteCost   float64 `json:"cacheWriteCost"`
	TotalCost        float64 `json:"totalCost"`
}

type ContextSnapshot struct {
	ID               string `json:"id"`
	PromptRecordID   string `json:"promptRecordId"`
	SourceKey        string `json:"sourceKey"`
	SourceVersion    string `json:"sourceVersion"`
	RenderedMarkdown string `json:"renderedMarkdown"`
	ValueJSON        string `json:"valueJson"`
	CreatedAt        string `json:"createdAt"`
}

type ToolResultArtifact struct {
	ID                   string `json:"id"`
	ProjectID            string `json:"projectId"`
	SessionID            string `json:"sessionId"`
	ToolCallID           string `json:"toolCallId"`
	ToolName             string `json:"toolName"`
	FullResultJSON       string `json:"fullResultJson"`
	ModelVisibleMarkdown string `json:"modelVisibleMarkdown"`
	Truncated            bool   `json:"truncated"`
	CreatedAt            string `json:"createdAt"`
}

type ToolCallInput struct {
	ProjectID     string
	SessionID     string
	ToolCallID    string
	ToolName      string
	ArgumentsJSON string
}

type ToolResult struct {
	ToolCallID           string `json:"toolCallId"`
	ToolName             string `json:"toolName"`
	FullResultJSON       string `json:"fullResultJson"`
	ModelVisibleMarkdown string `json:"modelVisibleMarkdown"`
	Truncated            bool   `json:"truncated"`
}

type GenerationCandidate struct {
	ID                string     `json:"id"`
	ProjectID         string     `json:"projectId"`
	SessionID         string     `json:"sessionId"`
	ContentItemID     string     `json:"contentItemId"`
	ActionKind        ActionKind `json:"actionKind"`
	OperationKind     string     `json:"operationKind"`
	ExpectedRevision  int64      `json:"expectedRevision"`
	SelectionStart    *int64     `json:"selectionStart,omitempty"`
	SelectionEnd      *int64     `json:"selectionEnd,omitempty"`
	InsertPosition    *int64     `json:"insertPosition,omitempty"`
	OriginalMarkdown  string     `json:"originalMarkdown,omitempty"`
	GeneratedMarkdown string     `json:"generatedMarkdown"`
	Reason            string     `json:"reason,omitempty"`
	ModelVariantID    string     `json:"modelVariantId,omitempty"`
	Status            string     `json:"status"`
	CreatedAt         string     `json:"createdAt"`
	UpdatedAt         string     `json:"updatedAt"`
}
