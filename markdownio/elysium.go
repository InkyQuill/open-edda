package markdownio

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type ImportedItem struct {
	Path         string
	Kind         string
	Title        string
	BodyMarkdown string
	MetadataJSON string
	Sections     []ImportedSection
	Relations    []ImportedRelation
}

type ImportedSection struct {
	Heading      string
	BodyMarkdown string
	SortOrder    int
}

type ImportedRelation struct {
	TargetTitle  string
	RelationType string
}

func ImportElysiumLayout(root string) ([]ImportedItem, error) {
	var items []ImportedItem

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || strings.ToLower(filepath.Ext(path)) != ".md" {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("relative path for %q: %w", path, err)
		}
		rel = filepath.ToSlash(rel)

		kind, metadataType, ok := elysiumKind(rel)
		if !ok {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %q: %w", path, err)
		}

		item, err := parseElysiumItem(rel, string(content), kind, metadataType)
		if err != nil {
			return fmt.Errorf("parse %q: %w", path, err)
		}
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

func ExportElysiumLayout(root string, items []ImportedItem) error {
	for _, item := range items {
		metadata, err := exportMetadata(item)
		if err != nil {
			return fmt.Errorf("metadata for %q: %w", item.Title, err)
		}

		rel, err := exportElysiumPath(item, metadata)
		if err != nil {
			return err
		}

		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("create directory for %q: %w", rel, err)
		}

		content, err := renderElysiumItem(item, metadata)
		if err != nil {
			return fmt.Errorf("render %q: %w", item.Title, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write %q: %w", rel, err)
		}
	}

	return nil
}

func elysiumKind(path string) (kind string, metadataType string, ok bool) {
	switch {
	case path == "genre.md":
		return "writing_brief", "genre", true
	case path == "synopsis.md":
		return "writing_brief", "synopsis", true
	case strings.HasPrefix(path, "story/"):
		return "chapter", "", true
	case strings.HasPrefix(path, "characters/"):
		return "story_bible_entry", "character", true
	case strings.HasPrefix(path, "worldbuilding/"):
		return "story_bible_entry", "worldbuilding", true
	case strings.HasPrefix(path, "braindump/"):
		return "project_note", "project_note", true
	default:
		return "", "", false
	}
}

func exportElysiumPath(item ImportedItem, metadata map[string]any) (string, error) {
	title := strings.TrimSpace(item.Title)
	if title == "" {
		return "", fmt.Errorf("export item has empty title")
	}

	filename, err := safeMarkdownFilename(title)
	if err != nil {
		return "", err
	}
	metadataType, _ := metadata["type"].(string)
	switch item.Kind {
	case "chapter":
		return filepath.ToSlash(filepath.Join("story", filename)), nil
	case "story_bible_entry":
		switch metadataType {
		case "character":
			return filepath.ToSlash(filepath.Join("characters", filename)), nil
		case "worldbuilding":
			return filepath.ToSlash(filepath.Join("worldbuilding", filename)), nil
		default:
			return filepath.ToSlash(filepath.Join("worldbuilding", filename)), nil
		}
	case "project_note":
		return filepath.ToSlash(filepath.Join("braindump", filename)), nil
	case "writing_brief":
		switch metadataType {
		case "genre":
			return "genre.md", nil
		case "synopsis":
			return "synopsis.md", nil
		default:
			return filename, nil
		}
	default:
		return "", fmt.Errorf("unsupported export item kind %q", item.Kind)
	}
}

func exportMetadata(item ImportedItem) (map[string]any, error) {
	metadata := make(map[string]any)
	if strings.TrimSpace(item.MetadataJSON) != "" {
		if err := json.Unmarshal([]byte(item.MetadataJSON), &metadata); err != nil {
			return nil, fmt.Errorf("parse metadata JSON: %w", err)
		}
	}

	if _, ok := metadata["related"]; !ok {
		related := relatedTargets(item.Relations)
		if len(related) > 0 {
			metadata["related"] = related
		}
	}

	return metadata, nil
}

func relatedTargets(relations []ImportedRelation) []string {
	var targets []string
	for _, relation := range relations {
		if relation.RelationType != "related" || strings.TrimSpace(relation.TargetTitle) == "" {
			continue
		}
		targets = append(targets, relation.TargetTitle)
	}
	return targets
}

func renderElysiumItem(item ImportedItem, metadata map[string]any) (string, error) {
	var builder strings.Builder
	builder.WriteString("# ")
	builder.WriteString(strings.TrimSpace(item.Title))
	builder.WriteString("\n---\n")
	frontmatter, err := renderFrontmatter(metadata)
	if err != nil {
		return "", err
	}
	builder.WriteString(frontmatter)
	builder.WriteString("---\n")

	body := normalizeMarkdownBlock(item.BodyMarkdown)
	if body != "" {
		builder.WriteString(body)
	}

	sections := sortedSections(item.Sections)
	for _, section := range sections {
		if builder.Len() > 0 && !strings.HasSuffix(builder.String(), "\n\n") {
			if strings.HasSuffix(builder.String(), "\n") {
				builder.WriteString("\n")
			} else {
				builder.WriteString("\n\n")
			}
		}
		builder.WriteString("## ")
		builder.WriteString(strings.TrimSpace(section.Heading))
		builder.WriteString("\n")
		builder.WriteString(normalizeMarkdownBlock(section.BodyMarkdown))
	}

	return strings.TrimRight(builder.String(), "\n") + "\n", nil
}

func renderFrontmatter(metadata map[string]any) (string, error) {
	keys := make([]string, 0, len(metadata))
	for key := range metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	for _, key := range keys {
		if strings.TrimSpace(key) == "" {
			return "", fmt.Errorf("metadata contains empty key")
		}
		if err := renderFrontmatterValue(&builder, key, metadata[key]); err != nil {
			return "", err
		}
	}
	return builder.String(), nil
}

func renderFrontmatterValue(builder *strings.Builder, key string, value any) error {
	switch typed := value.(type) {
	case []any:
		if len(typed) == 0 {
			builder.WriteString(key)
			builder.WriteString(": []\n")
			return nil
		}
		builder.WriteString(key)
		builder.WriteString(":\n")
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return fmt.Errorf("metadata list %q contains unsupported non-string value", key)
			}
			builder.WriteString("  - ")
			builder.WriteString(frontmatterScalar(text))
			builder.WriteString("\n")
		}
	case []string:
		if len(typed) == 0 {
			builder.WriteString(key)
			builder.WriteString(": []\n")
			return nil
		}
		builder.WriteString(key)
		builder.WriteString(":\n")
		for _, item := range typed {
			builder.WriteString("  - ")
			builder.WriteString(frontmatterScalar(item))
			builder.WriteString("\n")
		}
	default:
		if !supportedFrontmatterScalar(typed) {
			return fmt.Errorf("metadata key %q contains unsupported value", key)
		}
		builder.WriteString(key)
		builder.WriteString(": ")
		builder.WriteString(frontmatterScalar(typed))
		builder.WriteString("\n")
	}
	return nil
}

func frontmatterScalar(value any) string {
	switch typed := value.(type) {
	case string:
		return strconv.Quote(typed)
	case nil:
		return "null"
	default:
		content, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}
		return string(content)
	}
}

func supportedFrontmatterScalar(value any) bool {
	switch value.(type) {
	case string, bool, nil, float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
}

func parseFrontmatterScalar(value string) any {
	if strings.HasPrefix(value, `"`) {
		var quoted string
		if err := json.Unmarshal([]byte(value), &quoted); err == nil {
			return quoted
		}
	}
	switch value {
	case "true":
		return true
	case "false":
		return false
	case "null":
		return nil
	case "[]":
		return []string{}
	}
	if integer, err := strconv.ParseInt(value, 10, 64); err == nil {
		return integer
	}
	if number, err := strconv.ParseFloat(value, 64); err == nil {
		return number
	}
	return value
}

func safeMarkdownFilename(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", fmt.Errorf("export item has empty title")
	}
	if filepath.IsAbs(title) || strings.ContainsAny(title, `/\`) {
		return "", fmt.Errorf("export item title %q is not a safe filename", title)
	}
	clean := filepath.Clean(title)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean != title {
		return "", fmt.Errorf("export item title %q is not a safe filename", title)
	}
	return title + ".md", nil
}

func sortedSections(sections []ImportedSection) []ImportedSection {
	sorted := append([]ImportedSection(nil), sections...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].SortOrder < sorted[j].SortOrder
	})
	return sorted
}

func normalizeMarkdownBlock(content string) string {
	return strings.TrimRight(content, "\r\n")
}

func parseElysiumItem(path, content, kind, metadataType string) (ImportedItem, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	title, rest := splitTitle(path, content)
	frontmatter, body := splitFrontmatter(rest)

	metadata, err := parseFrontmatter(frontmatter)
	if err != nil {
		return ImportedItem{}, err
	}
	if metadataType != "" {
		if _, ok := metadata["type"]; !ok {
			metadata["type"] = metadataType
		}
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return ImportedItem{}, fmt.Errorf("marshal metadata: %w", err)
	}

	bodyMarkdown, sections := parseSections(body)

	return ImportedItem{
		Path:         path,
		Kind:         kind,
		Title:        title,
		BodyMarkdown: bodyMarkdown,
		MetadataJSON: string(metadataJSON),
		Sections:     sections,
		Relations:    importedRelations(metadata),
	}, nil
}

func splitTitle(path, content string) (title string, rest string) {
	line, remaining, hasLine := strings.Cut(content, "\n")
	if hasLine && strings.HasPrefix(line, "# ") {
		return strings.TrimSpace(strings.TrimPrefix(line, "# ")), remaining
	}
	if !hasLine && strings.HasPrefix(line, "# ") {
		return strings.TrimSpace(strings.TrimPrefix(line, "# ")), ""
	}
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), content
}

func splitFrontmatter(content string) (frontmatter string, body string) {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return "", strings.Trim(content, "\n")
	}

	for index := 1; index < len(lines); index++ {
		if strings.TrimSpace(lines[index]) != "---" {
			continue
		}

		frontmatter = strings.Join(lines[1:index], "\n")
		body = strings.Join(lines[index+1:], "\n")
		return frontmatter, strings.Trim(body, "\n")
	}

	return "", strings.Trim(content, "\n")
}

type pendingFrontmatterList struct {
	key   string
	items []string
}

func (list *pendingFrontmatterList) flush(metadata map[string]any) {
	if list.key == "" {
		return
	}
	if len(list.items) > 0 {
		metadata[list.key] = list.items
	} else {
		metadata[list.key] = ""
	}
	list.key = ""
	list.items = nil
}

func (list *pendingFrontmatterList) start(key string) {
	list.key = key
	list.items = nil
}

func (list *pendingFrontmatterList) append(item string) {
	list.items = append(list.items, item)
}

func (list *pendingFrontmatterList) active() bool {
	return list.key != ""
}

func parseFrontmatter(frontmatter string) (map[string]any, error) {
	metadata := make(map[string]any)
	if strings.TrimSpace(frontmatter) == "" {
		return metadata, nil
	}

	var list pendingFrontmatterList
	for _, line := range strings.Split(frontmatter, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if list.active() && strings.HasPrefix(trimmed, "- ") {
			item := parseFrontmatterScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
			text, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("frontmatter list %q contains non-string item", list.key)
			}
			list.append(text)
			continue
		}

		list.flush(metadata)

		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			return nil, fmt.Errorf("invalid frontmatter line %q", line)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return nil, fmt.Errorf("invalid empty frontmatter key in line %q", line)
		}
		if value == "" {
			list.start(key)
			continue
		}

		metadata[key] = parseFrontmatterScalar(value)
	}
	list.flush(metadata)

	return metadata, nil
}

func parseSections(body string) (string, []ImportedSection) {
	lines := strings.Split(body, "\n")
	var bodyLines []string
	var sections []ImportedSection
	var current *ImportedSection
	var currentLines []string

	flushCurrent := func() {
		if current == nil {
			return
		}
		current.BodyMarkdown = strings.Trim(strings.Join(currentLines, "\n"), "\n")
		sections = append(sections, *current)
		current = nil
		currentLines = nil
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			flushCurrent()
			current = &ImportedSection{
				Heading:   strings.TrimSpace(strings.TrimPrefix(line, "## ")),
				SortOrder: len(sections),
			}
			continue
		}

		if current != nil {
			currentLines = append(currentLines, line)
			continue
		}
		bodyLines = append(bodyLines, line)
	}
	flushCurrent()

	return strings.Trim(strings.Join(bodyLines, "\n"), "\n"), sections
}

func importedRelations(metadata map[string]any) []ImportedRelation {
	related, ok := metadata["related"].([]string)
	if !ok {
		return nil
	}

	relations := make([]ImportedRelation, 0, len(related))
	for _, target := range related {
		if target == "" {
			continue
		}
		relations = append(relations, ImportedRelation{
			TargetTitle:  target,
			RelationType: "related",
		})
	}
	return relations
}
