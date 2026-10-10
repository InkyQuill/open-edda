package project

import (
	"bytes"
	"context"
	"io"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// Metadata replaces only the source prefix; body bytes and unresolved records survive.
func replacePocketMetadata(source []byte, d *pocketDocument, input string) ([]byte, error) {
	normalized := strings.ReplaceAll(strings.ReplaceAll(input, "\r\n", "\n"), "\r", "\n")
	for _, line := range strings.Split(normalized, "\n") {
		if strings.TrimRight(line, " \t") == "---" || strings.TrimRight(line, " \t") == "..." {
			return nil, pocketInvalid("YAML delimiters are not allowed inside metadata")
		}
	}
	if err := validatePocketMetadata(normalized); err != nil {
		return nil, err
	}
	bom := []byte{}
	offset := 0
	if bytes.HasPrefix(source, []byte{0xef, 0xbb, 0xbf}) {
		bom = source[:3]
		offset = 3
	}
	eol := "\n"
	if bytes.Contains(source, []byte("\r\n")) {
		eol = "\r\n"
	}
	end := offset
	line, after := metadataLine(source, offset)
	if strings.TrimRight(line, " \t") == "---" && after > offset+len(line) {
		eol = string(source[offset+len(line) : after])
		end = -1
		for pos := after; pos < len(source); {
			line, next := metadataLine(source, pos)
			if strings.TrimRight(line, " \t") == "---" || strings.TrimRight(line, " \t") == "..." {
				if err := validatePocketMetadata(string(source[after:pos])); err != nil {
					return nil, pocketInvalid("existing block is not valid YAML mapping metadata; repair the source first")
				}
				end = next
				break
			}
			pos = next
		}
		if end < 0 {
			return nil, pocketInvalid("metadata has no closing delimiter; repair the source first")
		}
	}
	prefix := append(append([]byte{}, bom...), []byte("---"+eol+strings.ReplaceAll(normalized, "\n", eol))...)
	if normalized != "" && !strings.HasSuffix(normalized, "\n") {
		prefix = append(prefix, []byte(eol)...)
	}
	prefix = append(prefix, []byte("---"+eol)...)
	// Minimize replacement so anchors borrowing an unchanged delimiter stay intact.
	from := 0
	for from < end && from < len(prefix) && source[from] == prefix[from] {
		from++
	}
	for from > 0 && from < len(source) && !utf8.RuneStart(source[from]) {
		from--
	}
	suffix := 0
	for suffix < end-from && suffix < len(prefix)-from && source[end-suffix-1] == prefix[len(prefix)-suffix-1] {
		suffix++
	}
	for suffix > 0 && end-suffix < len(source) && !utf8.RuneStart(source[end-suffix]) {
		suffix--
	}
	to := end - suffix
	replacement := prefix[from : len(prefix)-suffix]
	next := append(append(append([]byte{}, source[:from]...), replacement...), source[to:]...)
	shift := len(replacement) - (to - from)
	for _, group := range []struct {
		records []PocketRecord
		signal  bool
	}{{d.Edits, false}, {d.Signals, true}} {
		for i := range group.records {
			record := &group.records[i]
			r := record.Resolution.Range
			if r == nil {
				continue
			}
			if r.From < to && r.To > from {
				return nil, pocketInvalid("a review overlaps changed metadata; resolve that record first")
			}
			moved := *r
			if r.From >= to {
				moved.From += shift
				moved.To += shift
			}
			if err := movePocketRecord(record, next, moved, group.signal); err != nil {
				return nil, err
			}
		}
	}
	for _, group := range []struct {
		key     string
		records []PocketRecord
	}{{"edits", d.Edits}, {"signals", d.Signals}} {
		if _, exists := d.raw[group.key]; !exists {
			continue
		}
		rows := []pocketJSON{}
		for _, r := range group.records {
			rows = append(rows, r.raw)
		}
		if err := pocketSet(d.raw, group.key, rows); err != nil {
			return nil, err
		}
	}
	return next, nil
}
func metadataLine(source []byte, start int) (string, int) {
	end := start
	for end < len(source) && source[end] != '\r' && source[end] != '\n' {
		end++
	}
	next := end
	if next < len(source) {
		next++
		if source[end] == '\r' && next < len(source) && source[next] == '\n' {
			next++
		}
	}
	return string(source[start:end]), next
}

// Plain Markdown can carry metadata without enrollment in a review/book workflow.
func (s *VersionStore) plainMetadata(ctx context.Context, author, project string, input PocketAction) (ProjectVersion, error) {
	v, err := s.Version(ctx, author, project, input.ExpectedVersion)
	if err != nil {
		return ProjectVersion{}, err
	}
	var chapter TreeEntry
	for _, e := range v.Entries {
		if e.ID == input.EntryID {
			chapter = e
		}
	}
	if chapter.Kind != "file" || !strings.HasSuffix(chapter.Path, ".md") {
		return ProjectVersion{}, pocketInvalid("select a Markdown file")
	}
	for _, e := range v.Entries {
		if e.Path == strings.TrimSuffix(chapter.Path, ".md")+".review.json" || e.Path == chapter.Path+".review.json" {
			return ProjectVersion{}, pocketInvalid("unbound review must be linked before metadata changes")
		}
	}
	source, err := s.pocketBytes(ctx, author, project, v, chapter)
	if err != nil {
		return ProjectVersion{}, err
	}
	next, err := replacePocketMetadata(source, &pocketDocument{raw: pocketJSON{}}, input.Metadata)
	if err != nil {
		return ProjectVersion{}, err
	}
	if len(next) > 1<<20 {
		return ProjectVersion{}, pocketInvalid("result exceeds browser limit")
	}
	hash := pocketHash(next)
	if err = s.UploadObject(ctx, author, project, hash, int64(len(next)), bytes.NewReader(next)); err != nil {
		return ProjectVersion{}, err
	}
	entries := append([]TreeEntry{}, v.Entries...)
	for i := range entries {
		if entries[i].ID == chapter.ID {
			entries[i].SHA256 = hash
			entries[i].Bytes = int64(len(next))
		}
	}
	return s.Publish(ctx, PublishVersionInput{AuthorID: author, ProjectID: project, ExpectedVersion: input.ExpectedVersion, OperationID: input.OperationID, Entries: entries})
}

func validatePocketMetadata(text string) error {
	var node yaml.Node
	decoder := yaml.NewDecoder(strings.NewReader(text))
	if err := decoder.Decode(&node); err != nil && err != io.EOF {
		return pocketInvalid("invalid YAML metadata")
	}
	if len(node.Content) > 0 && node.Content[0].Kind != yaml.MappingNode {
		return pocketInvalid("metadata must be a YAML mapping")
	}
	// Decode mappings too, to reject duplicate keys and invalid alias expansions.
	var value map[string]interface{}
	if err := yaml.Unmarshal([]byte(text), &value); err != nil {
		return pocketInvalid("invalid YAML metadata mapping")
	}
	return nil
}
