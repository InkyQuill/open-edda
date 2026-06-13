package skill

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
