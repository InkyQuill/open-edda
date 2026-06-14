package runtime

import "time"

const RuntimeVersion = "skill-script-runtime/v1"

type Status string

const (
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusTimedOut  Status = "timed_out"
	StatusRejected  Status = "rejected"
)

type OutputKind string

const (
	OutputKindReport        OutputKind = "report"
	OutputKindProposal      OutputKind = "proposal"
	OutputKindDraft         OutputKind = "draft"
	OutputKindGeneratedData OutputKind = "generated_data"
)

type Envelope struct {
	RuntimeVersion string         `json:"runtimeVersion"`
	Project        EnvelopeRef    `json:"project"`
	Skill          EnvelopeRef    `json:"skill"`
	Script         ScriptRef      `json:"script"`
	Inputs         EnvelopeInputs `json:"inputs"`
}

type EnvelopeRef struct {
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	Title    string `json:"title,omitempty"`
	Language string `json:"language,omitempty"`
}

type ScriptRef struct {
	FileID       string `json:"fileId"`
	RelativePath string `json:"relativePath"`
}

type EnvelopeInputs struct {
	ContentIDs    []string          `json:"contentIds,omitempty"`
	EntrySections []EntrySectionRef `json:"entrySections,omitempty"`
	Assets        []AssetInput      `json:"assets,omitempty"`
	Arguments     map[string]any    `json:"arguments,omitempty"`
}

type EntrySectionRef struct {
	ContentID string `json:"contentId"`
	Heading   string `json:"heading"`
}

type AssetInput struct {
	Path     string `json:"path"`
	BodyText string `json:"bodyText"`
}

type ScriptOutput struct {
	Kind          OutputKind     `json:"kind"`
	Title         string         `json:"title"`
	Markdown      string         `json:"markdown"`
	Proposals     []Proposal     `json:"proposals,omitempty"`
	GeneratedData map[string]any `json:"generatedData,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

type Proposal struct {
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	Title      string `json:"title"`
	Markdown   string `json:"markdown"`
}

type RunRequest struct {
	Command        string
	Timeout        time.Duration
	MaxStdoutBytes int64
	MaxStderrBytes int64
	Input          Envelope
}

type RunResult struct {
	Status       Status
	Output       ScriptOutput
	StdoutText   string
	StderrText   string
	ExitCode     int
	Duration     time.Duration
	ErrorMessage string
}
