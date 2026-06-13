package project

type ContentKind string

const (
	KindChapter         ContentKind = "chapter"
	KindStoryBibleEntry ContentKind = "story_bible_entry"
	KindWritingBrief    ContentKind = "writing_brief"
	KindProjectNote     ContentKind = "project_note"
)

type StoryProject struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Slug     string `json:"slug"`
	Language string `json:"language"`
}

type ContentItem struct {
	ID              string      `json:"id"`
	ProjectID       string      `json:"projectId"`
	Kind            ContentKind `json:"kind"`
	Title           string      `json:"title"`
	Slug            string      `json:"slug"`
	BodyMarkdown    string      `json:"bodyMarkdown"`
	MetadataJSON    string      `json:"metadataJson"`
	SortOrder       int64       `json:"sortOrder"`
	CurrentRevision int64       `json:"currentRevision"`
}

type ProjectMap struct {
	Project StoryProject            `json:"project"`
	Content []ProjectMapContentItem `json:"content"`
}

type ProjectMapContentItem struct {
	ContentItem
	Sections  []EntrySection  `json:"sections,omitempty"`
	Relations []EntryRelation `json:"relations,omitempty"`
}

type EntrySection struct {
	Heading      string `json:"heading"`
	BodyMarkdown string `json:"bodyMarkdown"`
	SortOrder    int64  `json:"sortOrder"`
}

type EntryRelation struct {
	TargetTitle  string `json:"targetTitle"`
	RelationType string `json:"relationType"`
}

type Revision struct {
	ID             string `json:"id"`
	ContentItemID  string `json:"contentItemId"`
	RevisionNumber int64  `json:"revisionNumber"`
	BodyMarkdown   string `json:"bodyMarkdown"`
	MetadataJSON   string `json:"metadataJson"`
	Reason         string `json:"reason"`
	CreatedBy      string `json:"createdBy"`
	CreatedAt      string `json:"createdAt"`
}

type AttachedNote struct {
	ID             string `json:"id"`
	ProjectID      string `json:"projectId"`
	ContentItemID  string `json:"contentItemId,omitempty"`
	SelectionStart int64  `json:"selectionStart,omitempty"`
	SelectionEnd   int64  `json:"selectionEnd,omitempty"`
	Title          string `json:"title"`
	BodyMarkdown   string `json:"bodyMarkdown"`
	Source         string `json:"source"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

type CreateProjectInput struct {
	AuthorID string
	Title    string
	Language string
}

type CreateContentInput struct {
	ProjectID    string
	Kind         ContentKind
	Title        string
	BodyMarkdown string
	MetadataJSON string
	SortOrder    int64
	Reason       string
	CreatedBy    string
}

type CreateEntrySectionInput struct {
	ProjectID     string
	ContentItemID string
	Heading       string
	BodyMarkdown  string
	SortOrder     int64
}

type CreateEntryRelationInput struct {
	ProjectID    string
	SourceItemID string
	TargetTitle  string
	RelationType string
}

type SearchContentInput struct {
	ProjectID       string
	Query           string
	Kind            ContentKind
	MetadataFilters map[string]string
	Tags            []string
	Limit           int64
}

type CreateAttachedNoteInput struct {
	ProjectID      string
	ContentItemID  string
	SelectionStart int64
	SelectionEnd   int64
	Title          string
	BodyMarkdown   string
	Source         string
}

type UpdateContentInput struct {
	ProjectID        string
	ContentID        string
	ExpectedRevision int64
	BodyMarkdown     string
	MetadataJSON     string
	Reason           string
	CreatedBy        string
}
