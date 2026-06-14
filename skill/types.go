package skill

import "git.inkyquill.net/inky/writer/skill/runtime"

type FilePurpose string

const (
	FilePurposeInstruction FilePurpose = "instruction"
	FilePurposeTemplate    FilePurpose = "template"
	FilePurposeReference   FilePurpose = "reference"
	FilePurposeData        FilePurpose = "data"
	FilePurposeScript      FilePurpose = "script"
	FilePurposeOther       FilePurpose = "other"
)

type SourceType string

const (
	SourceTypeUpload         SourceType = "upload"
	SourceTypeLocalDirectory SourceType = "local_directory"
)

type Skill struct {
	ID                   string        `json:"id"`
	ProjectID            string        `json:"projectId"`
	Name                 string        `json:"name"`
	DisplayName          string        `json:"displayName"`
	Description          string        `json:"description"`
	InstructionsMarkdown string        `json:"instructionsMarkdown,omitempty"`
	SourceType           SourceType    `json:"sourceType"`
	SourceLabel          string        `json:"sourceLabel"`
	ScriptCount          int64         `json:"scriptCount"`
	ScriptsDisabled      bool          `json:"scriptsDisabled"`
	MetadataJSON         string        `json:"metadataJson"`
	InstalledAt          string        `json:"installedAt"`
	UpdatedAt            string        `json:"updatedAt"`
	Files                []SkillFile   `json:"files,omitempty"`
	RoutingHints         []RoutingHint `json:"routingHints,omitempty"`
}

type SkillFile struct {
	ID             string      `json:"id"`
	SkillID        string      `json:"skillId"`
	RelativePath   string      `json:"relativePath"`
	Purpose        FilePurpose `json:"purpose"`
	MediaType      string      `json:"mediaType"`
	BodyText       string      `json:"bodyText,omitempty"`
	Bytes          int64       `json:"bytes"`
	ScriptDisabled bool        `json:"scriptDisabled"`
	CreatedAt      string      `json:"createdAt"`
}

type RoutingHint struct {
	ID          string `json:"id,omitempty"`
	SkillID     string `json:"skillId,omitempty"`
	ActionKind  string `json:"actionKind"`
	ContentKind string `json:"contentKind"`
	Tag         string `json:"tag"`
	Priority    int64  `json:"priority"`
	CreatedAt   string `json:"createdAt,omitempty"`
}

type ImportedSkill struct {
	Name                 string
	DisplayName          string
	Description          string
	InstructionsMarkdown string
	MetadataJSON         string
	SourceLabel          string
	ScriptCount          int64
	ScriptsDisabled      bool
	Files                []ImportedSkillFile
	RoutingHints         []RoutingHint
}

type ImportedSkillFile struct {
	RelativePath   string
	Purpose        FilePurpose
	MediaType      string
	BodyText       string
	Bytes          int64
	ScriptDisabled bool
}

type InstallInput struct {
	ProjectID   string
	SourceType  SourceType
	SourceLabel string
	Imported    ImportedSkill
}

type SelectSessionSkillsInput struct {
	ProjectID string
	SessionID string
	SkillIDs  []string
}

type RenderSkillInput struct {
	ProjectID string
	SkillID   string
}

type ScriptRecommendation string

const (
	ScriptRecommendationDisabled          ScriptRecommendation = "disabled"
	ScriptRecommendationApproveWithLimits ScriptRecommendation = "approve_with_limits"
	ScriptRecommendationDefer             ScriptRecommendation = "defer"
)

type ScriptRunStatus string

const (
	ScriptRunStatusSucceeded ScriptRunStatus = "succeeded"
	ScriptRunStatusFailed    ScriptRunStatus = "failed"
	ScriptRunStatusTimedOut  ScriptRunStatus = "timed_out"
	ScriptRunStatusRejected  ScriptRunStatus = "rejected"
)

type ScriptAudit struct {
	ID                    string               `json:"id"`
	ProjectID             string               `json:"projectId"`
	SkillID               string               `json:"skillId"`
	SkillFileID           string               `json:"skillFileId"`
	RelativePath          string               `json:"relativePath"`
	Runtime               string               `json:"runtime"`
	DestructiveOperations bool                 `json:"destructiveOperations"`
	FilesystemAccess      string               `json:"filesystemAccess"`
	NetworkAccess         bool                 `json:"networkAccess"`
	ExternalDependencies  string               `json:"externalDependencies"`
	ExpectedInputsJSON    string               `json:"expectedInputsJson"`
	ExpectedOutputsJSON   string               `json:"expectedOutputsJson"`
	RiskNotes             string               `json:"riskNotes"`
	Recommendation        ScriptRecommendation `json:"recommendation"`
	AuditedAt             string               `json:"auditedAt"`
	Approval              *ScriptApproval      `json:"approval,omitempty"`
}

type ScriptApproval struct {
	ID                string `json:"id"`
	ProjectID         string `json:"projectId"`
	SkillID           string `json:"skillId"`
	SkillFileID       string `json:"skillFileId"`
	AuditID           string `json:"auditId"`
	Enabled           bool   `json:"enabled"`
	RuntimeCommand    string `json:"runtimeCommand"`
	TimeoutMS         int64  `json:"timeoutMs"`
	MaxStdoutBytes    int64  `json:"maxStdoutBytes"`
	MaxStderrBytes    int64  `json:"maxStderrBytes"`
	AllowNetwork      bool   `json:"allowNetwork"`
	AllowProjectFiles bool   `json:"allowProjectFiles"`
	ApprovedBy        string `json:"approvedBy"`
	ApprovedAt        string `json:"approvedAt"`
	UpdatedAt         string `json:"updatedAt"`
}

type UpdateScriptApprovalInput struct {
	ProjectID         string
	SkillID           string
	SkillFileID       string
	Enabled           bool
	RuntimeCommand    string
	TimeoutMS         int64
	MaxStdoutBytes    int64
	MaxStderrBytes    int64
	AllowNetwork      bool
	AllowProjectFiles bool
	ApprovedBy        string
}

type RunScriptInput struct {
	ProjectID     string
	SessionID     string
	SkillID       string
	SkillFileID   string
	ScriptPath    string
	ToolCallID    string
	ContentIDs    []string
	EntrySections []ScriptEntrySectionInput
	AssetPaths    []string
	Arguments     map[string]any
}

type ScriptEntrySectionInput struct {
	ContentID string
	Heading   string
}

type ScriptRun struct {
	ID           string               `json:"id"`
	ProjectID    string               `json:"projectId"`
	SessionID    string               `json:"sessionId,omitempty"`
	SkillID      string               `json:"skillId"`
	SkillFileID  string               `json:"skillFileId"`
	ApprovalID   string               `json:"approvalId"`
	ToolCallID   string               `json:"toolCallId"`
	Status       ScriptRunStatus      `json:"status"`
	OutputKind   string               `json:"outputKind"`
	InputJSON    string               `json:"inputJson"`
	OutputJSON   string               `json:"outputJson"`
	Output       runtime.ScriptOutput `json:"output,omitempty"`
	StdoutText   string               `json:"stdoutText"`
	StderrText   string               `json:"stderrText"`
	ExitCode     int64                `json:"exitCode"`
	DurationMS   int64                `json:"durationMs"`
	ErrorMessage string               `json:"errorMessage"`
	CreatedAt    string               `json:"createdAt"`
}
