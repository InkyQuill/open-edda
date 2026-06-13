package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"git.inkyquill.net/inky/writer/project"
	"git.inkyquill.net/inky/writer/store"
)

const (
	maxToolVisibleLines = 2000
	maxToolVisibleBytes = 50 * 1024
)

func ContextToolDefinitions() []CompletionTool {
	return []CompletionTool{
		contextTool("project_map", "Read a concise map of project content.", objectSchema(map[string]any{})),
		contextTool("search_content", "Search project content by text, kind, metadata, and tags.", objectSchema(map[string]any{
			"query": map[string]any{"type": "string"},
			"kind": map[string]any{
				"type": "string",
				"enum": []string{"chapter", "story_bible_entry", "writing_brief", "project_note"},
			},
			"metadataFilters": map[string]any{
				"type":                 "object",
				"additionalProperties": map[string]any{"type": "string"},
			},
			"tags":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 100},
		})),
		contextTool("read_content", "Read one project content item.", contentIDSchema()),
		contextTool("read_chapter", "Read one chapter.", contentIDSchema()),
		contextTool("read_story_bible_entry", "Read one story bible entry.", contentIDSchema()),
		contextTool("read_entry_section", "Read one story bible entry section.", objectSchema(map[string]any{
			"contentId": map[string]any{"type": "string"},
			"heading":   map[string]any{"type": "string"},
		}, "contentId", "heading")),
		contextTool("list_revisions", "List revisions for one content item.", contentIDSchema()),
	}
}

func (s *Service) ExecuteTool(ctx context.Context, input ToolCallInput) (ToolResult, error) {
	if input.ProjectID == "" {
		return ToolResult{}, fmt.Errorf("project ID is required")
	}
	if input.SessionID != "" {
		if _, err := s.GetSession(ctx, input.ProjectID, input.SessionID); err != nil {
			return ToolResult{}, err
		}
	} else if _, err := s.queries.GetStoryProjectByID(ctx, input.ProjectID); err != nil {
		return ToolResult{}, fmt.Errorf("get story project: %w", err)
	}
	if input.ToolCallID == "" {
		return ToolResult{}, fmt.Errorf("tool call ID is required")
	}

	payload, directRead, err := s.executeContextTool(ctx, input)
	if err != nil {
		return ToolResult{}, err
	}

	fullJSONBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return ToolResult{}, fmt.Errorf("marshal tool result: %w", err)
	}
	fullJSON := string(fullJSONBytes)
	modelVisible, truncated := boundedToolMarkdown(input.ToolName, payload, directRead)
	result := ToolResult{
		ToolCallID:           input.ToolCallID,
		ToolName:             input.ToolName,
		FullResultJSON:       fullJSON,
		ModelVisibleMarkdown: modelVisible,
		Truncated:            truncated,
	}

	now := nowString()
	metadataJSON, err := json.Marshal(map[string]any{
		"toolName":    input.ToolName,
		"toolCallId":  input.ToolCallID,
		"truncated":   truncated,
		"resultBytes": len(fullJSONBytes),
	})
	if err != nil {
		return ToolResult{}, fmt.Errorf("marshal activity metadata: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ToolResult{}, fmt.Errorf("begin tool result transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	queries := s.queries.WithTx(tx)
	if err := queries.CreateActivityEvent(ctx, store.CreateActivityEventParams{
		ID:           newID("event"),
		ProjectID:    input.ProjectID,
		SessionID:    sql.NullString{String: input.SessionID, Valid: input.SessionID != ""},
		EventType:    "agent_tool_executed",
		Summary:      fmt.Sprintf("Executed %s", input.ToolName),
		MetadataJson: string(metadataJSON),
		CreatedAt:    now,
	}); err != nil {
		return ToolResult{}, fmt.Errorf("create activity event: %w", err)
	}
	if err := queries.CreateToolResultArtifact(ctx, store.CreateToolResultArtifactParams{
		ID:                   newID("artifact"),
		ProjectID:            input.ProjectID,
		SessionID:            sql.NullString{String: input.SessionID, Valid: input.SessionID != ""},
		ToolCallID:           input.ToolCallID,
		ToolName:             input.ToolName,
		FullResultJson:       fullJSON,
		ModelVisibleMarkdown: modelVisible,
		Truncated:            boolInt(truncated),
		FullResultBytes:      int64(len(fullJSONBytes)),
		CreatedAt:            now,
	}); err != nil {
		return ToolResult{}, fmt.Errorf("create tool result artifact: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return ToolResult{}, fmt.Errorf("commit tool result transaction: %w", err)
	}
	committed = true

	return result, nil
}

func (s *Service) executeContextTool(ctx context.Context, input ToolCallInput) (any, bool, error) {
	switch input.ToolName {
	case "project_map":
		var args struct{}
		if err := decodeToolArgs(input.ArgumentsJSON, &args); err != nil {
			return nil, false, err
		}
		projectMap, err := s.projectService.ProjectMap(ctx, input.ProjectID)
		return projectMap, false, err
	case "search_content":
		var args struct {
			Query           string              `json:"query"`
			Kind            project.ContentKind `json:"kind"`
			MetadataFilters map[string]string   `json:"metadataFilters"`
			Tags            []string            `json:"tags"`
			Limit           *int64              `json:"limit"`
		}
		if err := decodeToolArgs(input.ArgumentsJSON, &args); err != nil {
			return nil, false, err
		}
		if args.Kind != "" && !validToolContentKind(args.Kind) {
			return nil, false, fmt.Errorf("invalid content kind %q", args.Kind)
		}
		var limit int64
		if args.Limit != nil {
			if *args.Limit <= 0 || *args.Limit > 100 {
				return nil, false, fmt.Errorf("limit must be between 1 and 100")
			}
			limit = *args.Limit
		}
		items, err := s.projectService.SearchContent(ctx, project.SearchContentInput{
			ProjectID:       input.ProjectID,
			Query:           args.Query,
			Kind:            args.Kind,
			MetadataFilters: args.MetadataFilters,
			Tags:            args.Tags,
			Limit:           limit,
		})
		return items, false, err
	case "read_content":
		item, err := s.readContentTool(ctx, input.ProjectID, input.ArgumentsJSON, "")
		return item, true, err
	case "read_chapter":
		item, err := s.readContentTool(ctx, input.ProjectID, input.ArgumentsJSON, project.KindChapter)
		return item, true, err
	case "read_story_bible_entry":
		item, err := s.readContentTool(ctx, input.ProjectID, input.ArgumentsJSON, project.KindStoryBibleEntry)
		return item, true, err
	case "read_entry_section":
		var args struct {
			ContentID string `json:"contentId"`
			Heading   string `json:"heading"`
		}
		if err := decodeToolArgs(input.ArgumentsJSON, &args); err != nil {
			return nil, false, err
		}
		if strings.TrimSpace(args.ContentID) == "" {
			return nil, false, fmt.Errorf("contentId is required")
		}
		if strings.TrimSpace(args.Heading) == "" {
			return nil, false, fmt.Errorf("heading is required")
		}
		item, err := s.projectService.GetContent(ctx, input.ProjectID, args.ContentID)
		if err != nil {
			return nil, false, err
		}
		if item.Kind != project.KindStoryBibleEntry {
			return nil, false, fmt.Errorf("content %q kind = %q, want %q", args.ContentID, item.Kind, project.KindStoryBibleEntry)
		}
		sections, err := s.projectService.ListEntrySections(ctx, input.ProjectID, args.ContentID)
		if err != nil {
			return nil, false, err
		}
		for _, section := range sections {
			if section.Heading == args.Heading {
				return section, true, nil
			}
		}
		return nil, false, fmt.Errorf("entry section %q not found", args.Heading)
	case "list_revisions":
		var args struct {
			ContentID string `json:"contentId"`
		}
		if err := decodeToolArgs(input.ArgumentsJSON, &args); err != nil {
			return nil, false, err
		}
		revisions, err := s.projectService.ListRevisions(ctx, input.ProjectID, args.ContentID)
		return revisions, false, err
	default:
		return nil, false, fmt.Errorf("unknown tool %q", input.ToolName)
	}
}

func (s *Service) readContentTool(ctx context.Context, projectID, argsJSON string, wantKind project.ContentKind) (project.ContentItem, error) {
	var args struct {
		ContentID string `json:"contentId"`
	}
	if err := decodeToolArgs(argsJSON, &args); err != nil {
		return project.ContentItem{}, err
	}
	if strings.TrimSpace(args.ContentID) == "" {
		return project.ContentItem{}, fmt.Errorf("contentId is required")
	}
	item, err := s.projectService.GetContent(ctx, projectID, args.ContentID)
	if err != nil {
		return project.ContentItem{}, err
	}
	if wantKind != "" && item.Kind != wantKind {
		return project.ContentItem{}, fmt.Errorf("content %q kind = %q, want %q", args.ContentID, item.Kind, wantKind)
	}
	return item, nil
}

func decodeToolArgs(argumentsJSON string, out any) error {
	if strings.TrimSpace(argumentsJSON) == "" {
		argumentsJSON = "{}"
	}
	decoder := json.NewDecoder(strings.NewReader(argumentsJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("decode tool arguments: %w", err)
	}
	return nil
}

func boundedToolMarkdown(toolName string, payload any, directRead bool) (string, bool) {
	rendered := renderToolMarkdown(toolName, payload)
	bounded, truncated := boundModelVisible(rendered, directRead)
	if truncated {
		bounded += "\n\n_Result truncated; full JSON is stored in the tool result artifact._"
	}
	return bounded, truncated
}

func renderToolMarkdown(toolName string, payload any) string {
	var b strings.Builder
	b.WriteString("## ")
	b.WriteString(toolName)
	b.WriteString("\n\n")

	switch value := payload.(type) {
	case project.ContentItem:
		b.WriteString("# ")
		b.WriteString(value.Title)
		b.WriteString("\n\n")
		b.WriteString(value.BodyMarkdown)
	case project.EntrySection:
		b.WriteString("# ")
		b.WriteString(value.Heading)
		b.WriteString("\n\n")
		b.WriteString(value.BodyMarkdown)
	case []project.ContentItem:
		for _, item := range value {
			fmt.Fprintf(&b, "- %s (%s, revision %d)\n", item.Title, item.Kind, item.CurrentRevision)
			if item.BodyMarkdown != "" {
				b.WriteString("  ")
				b.WriteString(firstLine(item.BodyMarkdown))
				b.WriteByte('\n')
			}
		}
	case []project.Revision:
		for _, revision := range value {
			fmt.Fprintf(&b, "- revision %d: %s by %s at %s\n", revision.RevisionNumber, revision.Reason, revision.CreatedBy, revision.CreatedAt)
		}
	case project.ProjectMap:
		fmt.Fprintf(&b, "# %s\n\n", value.Project.Title)
		for _, item := range value.Content {
			fmt.Fprintf(&b, "- %s (%s, revision %d)\n", item.Title, item.Kind, item.CurrentRevision)
			for _, section := range item.Sections {
				fmt.Fprintf(&b, "  - section: %s\n", section.Heading)
			}
			for _, relation := range item.Relations {
				fmt.Fprintf(&b, "  - relation: %s %s\n", relation.RelationType, relation.TargetTitle)
			}
		}
	default:
		bytes, err := json.MarshalIndent(payload, "", "  ")
		if err == nil {
			b.Write(bytes)
		}
	}
	return b.String()
}

func boundModelVisible(markdown string, directRead bool) (string, bool) {
	lines := strings.Split(markdown, "\n")
	tooManyLines := len(lines) > maxToolVisibleLines
	tooManyBytes := len([]byte(markdown)) > maxToolVisibleBytes
	if !tooManyLines && !tooManyBytes {
		return markdown, false
	}

	if directRead {
		return boundDirectRead(lines), true
	}
	return boundListResult(lines), true
}

func boundDirectRead(lines []string) string {
	text := strings.Join(lines, "\n")
	if len(lines) <= 1 || longestLineBytes(lines) > maxToolVisibleBytes-256 {
		return boundTextHeadTail(text, maxToolVisibleBytes-256, 3, 4)
	}
	keepLines := maxToolVisibleLines - 8
	if keepLines < 2 {
		keepLines = 2
	}
	headCount := (keepLines * 3) / 4
	tailCount := keepLines - headCount
	if headCount > len(lines) {
		headCount = len(lines)
	}
	if tailCount > len(lines)-headCount {
		tailCount = len(lines) - headCount
	}
	if tailCount == 0 && len(lines) > 1 {
		tailCount = min(32, len(lines)/4)
		if tailCount < 1 {
			tailCount = 1
		}
		headCount = len(lines) - tailCount
	}

	for headCount > 1 {
		candidate := strings.Join(lines[:headCount], "\n")
		if tailCount > 0 {
			candidate += "\n\n...\n\n" + strings.Join(lines[len(lines)-tailCount:], "\n")
		}
		if len([]byte(candidate)) <= maxToolVisibleBytes-256 {
			return candidate
		}
		headCount--
		if headCount%3 == 0 && tailCount > 1 {
			tailCount--
		}
	}
	return boundTextHeadTail(text, maxToolVisibleBytes-256, 3, 4)
}

func boundListResult(lines []string) string {
	if len(lines) <= 1 || longestLineBytes(lines) > maxToolVisibleBytes-256 {
		lines = compactLongLines(lines, 4096)
	}
	text := strings.Join(lines, "\n")
	if len([]byte(text)) <= maxToolVisibleBytes-256 && len(lines) <= maxToolVisibleLines {
		return text
	}
	keepLines := maxToolVisibleLines - 8
	if keepLines < 2 {
		keepLines = 2
	}
	headCount := keepLines / 2
	tailCount := keepLines - headCount
	if headCount > len(lines) {
		headCount = len(lines)
	}
	if tailCount > len(lines)-headCount {
		tailCount = len(lines) - headCount
	}

	for headCount > 1 && tailCount > 1 {
		candidate := strings.Join(lines[:headCount], "\n") + "\n\n...\n\n" + strings.Join(lines[len(lines)-tailCount:], "\n")
		if len([]byte(candidate)) <= maxToolVisibleBytes-256 {
			return candidate
		}
		headCount--
		tailCount--
	}
	return boundTextHeadTail(text, maxToolVisibleBytes-256, 1, 2)
}

func compactLongLines(lines []string, maxBytes int) []string {
	compacted := make([]string, len(lines))
	for i, line := range lines {
		if len([]byte(line)) > maxBytes {
			compacted[i] = boundTextHeadTail(line, maxBytes, 1, 2)
		} else {
			compacted[i] = line
		}
	}
	return compacted
}

func boundTextHeadTail(value string, maxBytes int, headWeight int, totalWeight int) string {
	if len([]byte(value)) <= maxBytes {
		return value
	}
	if maxBytes <= 16 {
		return truncateUTF8(value, maxBytes)
	}
	separator := "\n\n...\n\n"
	available := maxBytes - len([]byte(separator))
	if available <= 0 {
		return truncateUTF8(value, maxBytes)
	}
	headBytes := (available * headWeight) / totalWeight
	tailBytes := available - headBytes
	head := truncateUTF8(value, headBytes)
	tail := utf8Tail(value, tailBytes)
	return head + separator + tail
}

func utf8Tail(value string, maxBytes int) string {
	if len([]byte(value)) <= maxBytes {
		return value
	}
	if maxBytes <= 0 {
		return ""
	}
	start := len(value) - maxBytes
	for start < len(value) && !utf8.ValidString(value[start:]) {
		start++
	}
	return value[start:]
}

func longestLineBytes(lines []string) int {
	longest := 0
	for _, line := range lines {
		if size := len([]byte(line)); size > longest {
			longest = size
		}
	}
	return longest
}

func truncateUTF8(value string, maxBytes int) string {
	if len([]byte(value)) <= maxBytes {
		return value
	}
	if maxBytes <= 0 {
		return ""
	}
	cut := value[:maxBytes]
	for !utf8.ValidString(cut) {
		cut = cut[:len(cut)-1]
	}
	return cut
}

func contextTool(name, description string, parameters map[string]any) CompletionTool {
	return CompletionTool{
		Type: "function",
		Function: CompletionToolFunction{
			Name:        name,
			Description: description,
			Parameters:  parameters,
		},
	}
}

func objectSchema(properties map[string]any, required ...string) map[string]any {
	schema := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func contentIDSchema() map[string]any {
	return objectSchema(map[string]any{
		"contentId": map[string]any{"type": "string"},
	}, "contentId")
}

func firstLine(value string) string {
	line, _, _ := strings.Cut(value, "\n")
	return line
}

func boolInt(value bool) int64 {
	if value {
		return 1
	}
	return 0
}

func validToolContentKind(kind project.ContentKind) bool {
	switch kind {
	case project.KindChapter, project.KindStoryBibleEntry, project.KindWritingBrief, project.KindProjectNote:
		return true
	default:
		return false
	}
}
