package markdownio

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"git.inkyquill.net/inky/writer/project"
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

func elysiumKind(path string) (kind string, metadataType string, ok bool) {
	switch {
	case path == "genre.md":
		return string(project.KindWritingBrief), "genre", true
	case path == "synopsis.md":
		return string(project.KindWritingBrief), "synopsis", true
	case strings.HasPrefix(path, "story/"):
		return string(project.KindChapter), "", true
	case strings.HasPrefix(path, "characters/"):
		return string(project.KindStoryBibleEntry), "character", true
	case strings.HasPrefix(path, "worldbuilding/"):
		return string(project.KindStoryBibleEntry), "worldbuilding", true
	default:
		return "", "", false
	}
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
			list.append(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
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

		metadata[key] = value
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
