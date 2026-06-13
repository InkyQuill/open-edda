package markdownio

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"git.inkyquill.net/inky/writer/project"
)

func TestImportElysiumLayout(t *testing.T) {
	items, err := ImportElysiumLayout(filepath.Join("testdata", "elysium"))
	if err != nil {
		t.Fatalf("ImportElysiumLayout() error = %v", err)
	}

	byPath := make(map[string]ImportedItem, len(items))
	for _, item := range items {
		byPath[item.Path] = item
	}

	tests := []struct {
		name string
		path string
		kind string
	}{
		{
			name: "chapter",
			path: filepath.ToSlash(filepath.Join("story", "Chapter 1.md")),
			kind: string(project.KindChapter),
		},
		{
			name: "character",
			path: filepath.ToSlash(filepath.Join("characters", "Hugh.md")),
			kind: string(project.KindStoryBibleEntry),
		},
		{
			name: "worldbuilding",
			path: filepath.ToSlash(filepath.Join("worldbuilding", "Dwarves.md")),
			kind: string(project.KindStoryBibleEntry),
		},
		{
			name: "genre",
			path: "genre.md",
			kind: string(project.KindWritingBrief),
		},
		{
			name: "synopsis",
			path: "synopsis.md",
			kind: string(project.KindWritingBrief),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, ok := byPath[tt.path]
			if !ok {
				t.Fatalf("missing imported item %q", tt.path)
			}
			if item.Kind != tt.kind {
				t.Fatalf("kind = %q, want %q", item.Kind, tt.kind)
			}
		})
	}

	character := byPath[filepath.ToSlash(filepath.Join("characters", "Hugh.md"))]
	assertMetadataString(t, character.MetadataJSON, "type", "character")
	if len(character.Relations) != 1 {
		t.Fatalf("character relations len = %d, want 1", len(character.Relations))
	}
	if character.Relations[0] != (ImportedRelation{TargetTitle: "Dwarves", RelationType: "related"}) {
		t.Fatalf("character relation = %#v, want Dwarves related", character.Relations[0])
	}

	dwarves := byPath[filepath.ToSlash(filepath.Join("worldbuilding", "Dwarves.md"))]
	assertMetadataString(t, dwarves.MetadataJSON, "type", "worldbuilding")
	assertMetadataString(t, dwarves.MetadataJSON, "status", "canon")
	if len(dwarves.Sections) != 2 {
		t.Fatalf("dwarves sections len = %d, want 2", len(dwarves.Sections))
	}
	if dwarves.Sections[0].Heading != "Player Perspective" || dwarves.Sections[0].SortOrder != 0 {
		t.Fatalf("first dwarves section = %#v, want Player Perspective at 0", dwarves.Sections[0])
	}
	if dwarves.Sections[1].Heading != "NPC Perspective" || dwarves.Sections[1].SortOrder != 1 {
		t.Fatalf("second dwarves section = %#v, want NPC Perspective at 1", dwarves.Sections[1])
	}
	if dwarves.Sections[0].BodyMarkdown != "Dwarves are compact, broad-shouldered, grounded, and built for endurance." {
		t.Fatalf("first dwarves section body = %q", dwarves.Sections[0].BodyMarkdown)
	}
	if len(dwarves.Relations) != 1 {
		t.Fatalf("dwarves relations len = %d, want 1", len(dwarves.Relations))
	}
	if dwarves.Relations[0] != (ImportedRelation{TargetTitle: "Species", RelationType: "related"}) {
		t.Fatalf("dwarves relation = %#v, want Species related", dwarves.Relations[0])
	}
}

func TestImportElysiumLayoutParserEdges(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "story/No H1.md", "Opening line.\r\n\r\nSecond line.\r\n")
	writeFile(t, root, "characters/Empty Scalar.md", "# Empty Scalar\n---\npronouns:\nrelated:\n  - Hugh\n---   \nCharacter body.\n")
	writeFile(t, root, "worldbuilding/Dashes.md", "# Dashes\n---\nstatus: canon\nupdated: 2026-06-13\n---   \n--- not frontmatter\nBody after frontmatter.\n")
	writeFile(t, root, "unknown/Ignored.md", "# Ignored\n\nThis should not import.\n")

	items, err := ImportElysiumLayout(root)
	if err != nil {
		t.Fatalf("ImportElysiumLayout() error = %v", err)
	}

	byPath := make(map[string]ImportedItem, len(items))
	for _, item := range items {
		byPath[item.Path] = item
	}
	if len(byPath) != 3 {
		t.Fatalf("imported items = %d, want 3: %#v", len(byPath), byPath)
	}
	if _, ok := byPath["unknown/Ignored.md"]; ok {
		t.Fatal("unknown markdown path was imported")
	}

	noH1 := byPath["story/No H1.md"]
	if noH1.Title != "No H1" {
		t.Fatalf("fallback title = %q, want No H1", noH1.Title)
	}
	if noH1.BodyMarkdown != "Opening line.\n\nSecond line." {
		t.Fatalf("CRLF-normalized body = %q", noH1.BodyMarkdown)
	}

	emptyScalar := byPath["characters/Empty Scalar.md"]
	var metadata map[string]any
	if err := json.Unmarshal([]byte(emptyScalar.MetadataJSON), &metadata); err != nil {
		t.Fatalf("metadata json = %q: %v", emptyScalar.MetadataJSON, err)
	}
	if metadata["pronouns"] != "" {
		t.Fatalf("empty scalar metadata = %#v, want empty string", metadata["pronouns"])
	}
	if len(emptyScalar.Relations) != 1 || emptyScalar.Relations[0].TargetTitle != "Hugh" {
		t.Fatalf("relations = %#v, want Hugh related", emptyScalar.Relations)
	}
	if emptyScalar.BodyMarkdown != "Character body." {
		t.Fatalf("body with delimiter trailing spaces = %q", emptyScalar.BodyMarkdown)
	}

	dashes := byPath["worldbuilding/Dashes.md"]
	assertMetadataString(t, dashes.MetadataJSON, "updated", "2026-06-13")
	if dashes.BodyMarkdown != "--- not frontmatter\nBody after frontmatter." {
		t.Fatalf("line-exact delimiter body = %q", dashes.BodyMarkdown)
	}
}

func TestImportElysiumLayoutRejectsMalformedFrontmatter(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "characters/Broken.md", "# Broken\n---\ninvalid\n---\nBody.\n")

	if _, err := ImportElysiumLayout(root); err == nil {
		t.Fatal("ImportElysiumLayout() error = nil, want malformed frontmatter error")
	}
}

func writeFile(t *testing.T, root string, rel string, content string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir fixture dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture %q: %v", rel, err)
	}
}

func assertMetadataString(t *testing.T, metadataJSON, key, want string) {
	t.Helper()

	var metadata map[string]any
	if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
		t.Fatalf("metadata json = %q: %v", metadataJSON, err)
	}

	got, ok := metadata[key].(string)
	if !ok {
		t.Fatalf("metadata[%q] = %#v, want string", key, metadata[key])
	}
	if got != want {
		t.Fatalf("metadata[%q] = %q, want %q", key, got, want)
	}
}
