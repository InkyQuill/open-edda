package agent

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
	ID                   string `json:"id"`
	ProjectID            string `json:"projectId"`
	InstructionsMarkdown string `json:"instructionsMarkdown"`
	CreatedAt            string `json:"createdAt"`
	UpdatedAt            string `json:"updatedAt"`
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

type GenerationCandidate struct {
	ID                string     `json:"id"`
	ProjectID         string     `json:"projectId"`
	SessionID         string     `json:"sessionId"`
	ContentItemID     string     `json:"contentItemId"`
	ActionKind        ActionKind `json:"actionKind"`
	OperationKind     string     `json:"operationKind"`
	ExpectedRevision  int64      `json:"expectedRevision"`
	GeneratedMarkdown string     `json:"generatedMarkdown"`
	Status            string     `json:"status"`
	CreatedAt         string     `json:"createdAt"`
	UpdatedAt         string     `json:"updatedAt"`
}
