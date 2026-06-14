package skill

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	maxSkillArchiveBytes int64 = 5 << 20
	maxSkillFileBytes    int64 = 512 << 10
	maxSkillArchiveFiles       = 256
)

type entry struct {
	path string
	file *zip.File
}

type skillFrontmatter struct {
	Name        string
	Description string
	Route       routingFrontmatter
	Routing     routingFrontmatter
}

type routingFrontmatter struct {
	ActionKinds  []string
	Actions      []string
	ContentKinds []string
	Content      []string
	Tags         []string
	Priority     int64
}

func ParseSkillArchive(reader io.ReaderAt, size int64, sourceLabel string) (ImportedSkill, error) {
	if size > maxSkillArchiveBytes {
		return ImportedSkill{}, fmt.Errorf("skill archive exceeds %d bytes", maxSkillArchiveBytes)
	}

	zr, err := zip.NewReader(reader, size)
	if err != nil {
		return ImportedSkill{}, fmt.Errorf("open skill archive: %w", err)
	}

	entries := make([]entry, 0, len(zr.File))
	var declaredBytes int64
	for _, file := range zr.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if len(entries)+1 > maxSkillArchiveFiles {
			return ImportedSkill{}, fmt.Errorf("skill archive contains more than %d files", maxSkillArchiveFiles)
		}
		if file.UncompressedSize64 > uint64(maxSkillFileBytes) {
			return ImportedSkill{}, fmt.Errorf("file %s exceeds %d bytes", file.Name, maxSkillFileBytes)
		}
		if file.UncompressedSize64 > uint64(maxSkillArchiveBytes) || declaredBytes+int64(file.UncompressedSize64) > maxSkillArchiveBytes {
			return ImportedSkill{}, fmt.Errorf("skill archive contents exceed %d bytes", maxSkillArchiveBytes)
		}
		relativePath, err := safeRelativePath(file.Name)
		if err != nil {
			return ImportedSkill{}, err
		}
		declaredBytes += int64(file.UncompressedSize64)
		entries = append(entries, entry{path: relativePath, file: file})
	}

	root := commonSkillRoot(entries)
	files := make([]ImportedSkillFile, 0, len(entries))
	var instructions string
	var frontmatter skillFrontmatter
	var hasSkillFile bool
	var scriptCount int64
	var actualBytes int64

	for _, item := range entries {
		relativePath := stripCommonRoot(item.path, root)
		body, err := readZipText(item.file)
		if err != nil {
			return ImportedSkill{}, fmt.Errorf("read %s: %w", item.path, err)
		}
		actualBytes += int64(len(body))
		if actualBytes > maxSkillArchiveBytes {
			return ImportedSkill{}, fmt.Errorf("skill archive contents exceed %d bytes", maxSkillArchiveBytes)
		}

		purpose, scriptDisabled := classifySkillFile(relativePath)
		if purpose == FilePurposeScript {
			scriptCount++
		}
		if relativePath == "SKILL.md" {
			hasSkillFile = true
			frontmatter, instructions, err = parseSkillMarkdown(body)
			if err != nil {
				return ImportedSkill{}, err
			}
		}

		files = append(files, ImportedSkillFile{
			RelativePath:   relativePath,
			Purpose:        purpose,
			MediaType:      mediaTypeForPath(relativePath),
			BodyText:       body,
			Bytes:          int64(len(body)),
			ScriptDisabled: scriptDisabled,
		})
	}

	if !hasSkillFile {
		return ImportedSkill{}, fmt.Errorf("skill archive missing SKILL.md")
	}
	name, err := normalizeSkillName(frontmatter.Name)
	if err != nil {
		return ImportedSkill{}, err
	}

	return ImportedSkill{
		Name:                 name,
		DisplayName:          name,
		Description:          frontmatter.Description,
		InstructionsMarkdown: instructions,
		MetadataJSON:         frontmatterMetadataJSON(frontmatter),
		SourceLabel:          sourceLabel,
		ScriptCount:          scriptCount,
		ScriptsDisabled:      true,
		Files:                files,
		RoutingHints:         frontmatter.routingHints(),
	}, nil
}

func safeRelativePath(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("unsafe archive path: empty")
	}
	if strings.Contains(raw, "\\") {
		return "", fmt.Errorf("unsafe archive path %q: backslashes are not allowed", raw)
	}
	if path.IsAbs(raw) || isWindowsDrivePath(raw) {
		return "", fmt.Errorf("unsafe archive path %q: absolute paths are not allowed", raw)
	}
	for _, part := range strings.Split(raw, "/") {
		if part == "" || part == ".." {
			return "", fmt.Errorf("unsafe archive path %q", raw)
		}
	}
	clean := path.Clean(raw)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("unsafe archive path %q", raw)
	}
	return clean, nil
}

func isWindowsDrivePath(raw string) bool {
	if len(raw) < 2 || raw[1] != ':' {
		return false
	}
	c := raw[0]
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func readZipText(file *zip.File) (string, error) {
	if file.UncompressedSize64 > uint64(maxSkillFileBytes) {
		return "", fmt.Errorf("file exceeds %d bytes", maxSkillFileBytes)
	}

	rc, err := file.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, maxSkillFileBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(data)) > maxSkillFileBytes {
		return "", fmt.Errorf("file exceeds %d bytes", maxSkillFileBytes)
	}
	if !utf8.Valid(data) {
		return "", fmt.Errorf("file is not valid UTF-8")
	}
	return string(data), nil
}

func commonSkillRoot(entries []entry) string {
	if len(entries) == 0 {
		return ""
	}
	firstRoot, ok := splitTopLevel(entries[0].path)
	if !ok {
		return ""
	}
	for _, item := range entries[1:] {
		root, ok := splitTopLevel(item.path)
		if !ok || root != firstRoot {
			return ""
		}
	}
	return firstRoot
}

func splitTopLevel(relativePath string) (string, bool) {
	before, after, ok := strings.Cut(relativePath, "/")
	return before, ok && after != ""
}

func stripCommonRoot(relativePath, root string) string {
	if root == "" {
		return relativePath
	}
	return strings.TrimPrefix(relativePath, root+"/")
}

func parseSkillMarkdown(body string) (skillFrontmatter, string, error) {
	if !strings.HasPrefix(body, "---") {
		return skillFrontmatter{}, body, fmt.Errorf("skill frontmatter missing")
	}

	lines := strings.SplitAfter(body, "\n")
	if strings.TrimSpace(lines[0]) != "---" {
		return skillFrontmatter{}, body, fmt.Errorf("skill frontmatter missing")
	}

	var frontmatter strings.Builder
	bodyStart := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			bodyStart = i + 1
			break
		}
		frontmatter.WriteString(lines[i])
	}
	if bodyStart == -1 {
		return skillFrontmatter{}, body, fmt.Errorf("skill frontmatter is not closed")
	}

	parsed, err := parseSkillFrontmatter(frontmatter.String())
	if err != nil {
		return skillFrontmatter{}, body, err
	}
	return parsed, strings.Join(lines[bodyStart:], ""), nil
}

func parseSkillFrontmatter(body string) (skillFrontmatter, error) {
	var parsed skillFrontmatter
	section := ""

	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.Contains(trimmed, ":") {
			return skillFrontmatter{}, fmt.Errorf("invalid frontmatter line %q", line)
		}

		key, value, _ := strings.Cut(trimmed, ":")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "name":
			parsed.Name = trimScalar(value)
			section = ""
		case "description":
			parsed.Description = trimScalar(value)
			section = ""
		case "route":
			section = "route"
		case "routing":
			section = "routing"
		default:
			if section != "route" && section != "routing" {
				continue
			}
			if err := setRoutingField(routingForSection(&parsed, section), key, value); err != nil {
				return skillFrontmatter{}, err
			}
		}
	}

	if strings.TrimSpace(parsed.Name) == "" {
		return skillFrontmatter{}, fmt.Errorf("skill frontmatter requires name")
	}
	return parsed, nil
}

func routingForSection(frontmatter *skillFrontmatter, section string) *routingFrontmatter {
	if section == "routing" {
		return &frontmatter.Routing
	}
	return &frontmatter.Route
}

func setRoutingField(route *routingFrontmatter, key, value string) error {
	switch key {
	case "actionKinds":
		route.ActionKinds = parseInlineStringList(value)
	case "actions":
		route.Actions = parseInlineStringList(value)
	case "contentKinds":
		route.ContentKinds = parseInlineStringList(value)
	case "content":
		route.Content = parseInlineStringList(value)
	case "tags":
		route.Tags = parseInlineStringList(value)
	case "priority":
		priority, err := strconv.ParseInt(trimScalar(value), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid routing priority %q: %w", value, err)
		}
		route.Priority = priority
	}
	return nil
}

func parseInlineStringList(value string) []string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		item := trimScalar(part)
		if item != "" {
			values = append(values, item)
		}
	}
	return values
}

func trimScalar(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	return value
}

func normalizeSkillName(name string) (string, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return "", fmt.Errorf("skill name is required")
	}
	for _, r := range name {
		if ('a' <= r && r <= 'z') || ('0' <= r && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return "", fmt.Errorf("invalid skill name %q", name)
	}
	return name, nil
}

func classifySkillFile(relative string) (FilePurpose, bool) {
	switch {
	case relative == "SKILL.md":
		return FilePurposeInstruction, false
	case strings.HasPrefix(relative, "templates/"):
		return FilePurposeTemplate, false
	case strings.HasPrefix(relative, "references/"), strings.HasPrefix(relative, "reference/"):
		return FilePurposeReference, false
	case strings.HasPrefix(relative, "data/"):
		return FilePurposeData, false
	case strings.HasPrefix(relative, "scripts/"), hasScriptExtension(relative):
		return FilePurposeScript, true
	default:
		return FilePurposeOther, false
	}
}

func hasScriptExtension(relative string) bool {
	switch strings.ToLower(path.Ext(relative)) {
	case ".sh", ".bash", ".zsh", ".ps1", ".bat", ".cmd", ".py", ".js", ".ts":
		return true
	default:
		return false
	}
}

func frontmatterMetadataJSON(frontmatter skillFrontmatter) string {
	metadata := map[string]string{
		"name":        frontmatter.Name,
		"description": frontmatter.Description,
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func (frontmatter skillFrontmatter) routingHints() []RoutingHint {
	hints := frontmatter.Route.routingHints()
	hints = append(hints, frontmatter.Routing.routingHints()...)
	return hints
}

func (routing routingFrontmatter) routingHints() []RoutingHint {
	actions := append([]string{}, routing.ActionKinds...)
	actions = append(actions, routing.Actions...)
	contents := append([]string{}, routing.ContentKinds...)
	contents = append(contents, routing.Content...)

	var hints []RoutingHint
	for _, action := range actions {
		for _, content := range contents {
			hints = append(hints, RoutingHint{
				ActionKind:  action,
				ContentKind: content,
				Priority:    routing.Priority,
			})
		}
		for _, tag := range routing.Tags {
			hints = append(hints, RoutingHint{
				ActionKind: action,
				Tag:        tag,
				Priority:   routing.Priority,
			})
		}
	}
	if len(actions) == 0 {
		for _, content := range contents {
			hints = append(hints, RoutingHint{
				ContentKind: content,
				Priority:    routing.Priority,
			})
		}
		for _, tag := range routing.Tags {
			hints = append(hints, RoutingHint{
				Tag:      tag,
				Priority: routing.Priority,
			})
		}
	}
	return hints
}

func mediaTypeForPath(relative string) string {
	if relative == "SKILL.md" || strings.HasSuffix(strings.ToLower(relative), ".md") {
		return "text/markdown; charset=utf-8"
	}
	if mediaType := mime.TypeByExtension(path.Ext(relative)); mediaType != "" {
		return mediaType
	}
	return "text/plain; charset=utf-8"
}

func ParseSkillDirectory(dirPath string) (ImportedSkill, error) {
	info, err := os.Stat(dirPath)
	if err != nil {
		return ImportedSkill{}, fmt.Errorf("stat directory: %w", err)
	}
	if !info.IsDir() {
		return ImportedSkill{}, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	var entries []struct {
		rel  string
		body string
	}
	err = filepath.WalkDir(dirPath, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dirPath, filePath)
		if err != nil {
			return err
		}
		if safe, safeErr := safeRelativePath(rel); safeErr != nil {
			return safeErr
		} else {
			rel = safe
		}
		raw, readErr := os.ReadFile(filePath)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", rel, readErr)
		}
		if !utf8.Valid(raw) {
			return fmt.Errorf("file %s is not valid UTF-8", rel)
		}
		if int64(len(raw)) > maxSkillFileBytes {
			return fmt.Errorf("file %s exceeds max size", rel)
		}
		entries = append(entries, struct {
			rel  string
			body string
		}{rel: rel, body: string(raw)})
		return nil
	})
	if err != nil {
		return ImportedSkill{}, fmt.Errorf("walk directory: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].rel < entries[j].rel })

	root := ""
	for _, entry := range entries {
		parts := strings.SplitN(entry.rel, "/", 2)
		if root == "" {
			root = parts[0]
		} else if root != parts[0] {
			root = ""
			break
		}
	}
	if root != "" && root != "SKILL.md" {
		for i := range entries {
			entries[i].rel = strings.TrimPrefix(entries[i].rel, root+"/")
		}
	}

	var skillFile *struct {
		rel  string
		body string
	}
	for i := range entries {
		if entries[i].rel == "SKILL.md" {
			skillFile = &entries[i]
			break
		}
	}
	if skillFile == nil {
		return ImportedSkill{}, fmt.Errorf("skill directory must contain SKILL.md")
	}

	frontmatter, instructions, err := parseSkillMarkdown(skillFile.body)
	if err != nil {
		return ImportedSkill{}, err
	}
	name, err := normalizeSkillName(frontmatter.Name)
	if err != nil {
		return ImportedSkill{}, fmt.Errorf("skill frontmatter name: %w", err)
	}

	imported := ImportedSkill{
		Name:                 name,
		DisplayName:          frontmatter.Name,
		Description:          frontmatter.Description,
		InstructionsMarkdown: strings.TrimSpace(instructions),
		MetadataJSON:         frontmatterMetadataJSON(frontmatter),
		SourceLabel:          dirPath,
		ScriptsDisabled:      true,
		RoutingHints:         frontmatter.routingHints(),
	}
	for _, entry := range entries {
		purpose, disabled := classifySkillFile(entry.rel)
		file := ImportedSkillFile{
			RelativePath:   entry.rel,
			Purpose:        purpose,
			MediaType:      mediaTypeForPath(entry.rel),
			BodyText:       entry.body,
			Bytes:          int64(len([]byte(entry.body))),
			ScriptDisabled: disabled,
		}
		if purpose == FilePurposeScript {
			imported.ScriptCount++
		}
		imported.Files = append(imported.Files, file)
	}
	return imported, nil
}
