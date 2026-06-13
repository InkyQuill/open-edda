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

func TestExportElysiumLayoutWritesMappedPathsAndSections(t *testing.T) {
	root := t.TempDir()
	items := []ImportedItem{
		{
			Kind:         string(project.KindChapter),
			Title:        "Chapter 1",
			BodyMarkdown: "Chapter body.\n\nWith a second paragraph.\n\n",
			MetadataJSON: `{"status":"draft"}`,
		},
		{
			Kind:         string(project.KindStoryBibleEntry),
			Title:        "Hugh",
			BodyMarkdown: "Character body.",
			MetadataJSON: `{"type":"character","related":["Dwarves"]}`,
		},
		{
			Kind:         string(project.KindStoryBibleEntry),
			Title:        "Dwarves",
			BodyMarkdown: "Worldbuilding body.",
			MetadataJSON: `{"type":"worldbuilding","status":"canon"}`,
			Sections: []ImportedSection{
				{
					Heading:      "NPC Perspective",
					BodyMarkdown: "NPC-facing details.",
					SortOrder:    1,
				},
				{
					Heading:      "Player Perspective",
					BodyMarkdown: "Player-facing details.",
					SortOrder:    0,
				},
			},
		},
		{
			Kind:         string(project.KindProjectNote),
			Title:        "Loose Ideas",
			BodyMarkdown: "Note body.",
			MetadataJSON: `{"mood":"exploratory"}`,
		},
	}

	if err := ExportElysiumLayout(root, items); err != nil {
		t.Fatalf("ExportElysiumLayout() error = %v", err)
	}

	assertFile(t, root, "story/Chapter 1.md", "# Chapter 1\n---\nstatus: \"draft\"\n---\nChapter body.\n\nWith a second paragraph.\n")
	assertFile(t, root, "characters/Hugh.md", "# Hugh\n---\nrelated:\n  - \"Dwarves\"\ntype: \"character\"\n---\nCharacter body.\n")
	assertFile(t, root, "worldbuilding/Dwarves.md", "# Dwarves\n---\nstatus: \"canon\"\ntype: \"worldbuilding\"\n---\nWorldbuilding body.\n\n## Player Perspective\nPlayer-facing details.\n\n## NPC Perspective\nNPC-facing details.\n")
	assertFile(t, root, "braindump/Loose Ideas.md", "# Loose Ideas\n---\nmood: \"exploratory\"\n---\nNote body.\n")
}

func TestExportElysiumLayoutWritesBriefPathsAndMetadataRoundTrips(t *testing.T) {
	root := t.TempDir()
	items := []ImportedItem{
		{
			Kind:         string(project.KindWritingBrief),
			Title:        "Genre",
			BodyMarkdown: "Fantasy adventure.",
			MetadataJSON: `{"audience":"adult","type":"genre"}`,
		},
		{
			Kind:         string(project.KindWritingBrief),
			Title:        "Synopsis",
			BodyMarkdown: "A compact plot summary.",
			MetadataJSON: `{"type":"synopsis"}`,
		},
	}

	if err := ExportElysiumLayout(root, items); err != nil {
		t.Fatalf("ExportElysiumLayout() error = %v", err)
	}

	items, err := ImportElysiumLayout(root)
	if err != nil {
		t.Fatalf("ImportElysiumLayout() exported files error = %v", err)
	}

	byPath := make(map[string]ImportedItem, len(items))
	for _, item := range items {
		byPath[item.Path] = item
	}
	if _, ok := byPath["genre.md"]; !ok {
		t.Fatal("missing exported genre.md")
	}
	if _, ok := byPath["synopsis.md"]; !ok {
		t.Fatal("missing exported synopsis.md")
	}
	assertMetadataString(t, byPath["genre.md"].MetadataJSON, "audience", "adult")
	assertMetadataString(t, byPath["genre.md"].MetadataJSON, "type", "genre")
	assertMetadataString(t, byPath["synopsis.md"].MetadataJSON, "type", "synopsis")
}

func TestExportElysiumLayoutPreservesBodyLineEndings(t *testing.T) {
	root := t.TempDir()
	if err := ExportElysiumLayout(root, []ImportedItem{
		{
			Kind:         string(project.KindChapter),
			Title:        "Line Endings",
			BodyMarkdown: "First line\r\nSecond line\r\n",
			Sections: []ImportedSection{
				{
					Heading:      "Notes",
					BodyMarkdown: "Section first\r\nSection second\r\n",
				},
			},
		},
	}); err != nil {
		t.Fatalf("ExportElysiumLayout() error = %v", err)
	}

	assertFile(t, root, "story/Line Endings.md", "# Line Endings\n---\n---\nFirst line\r\nSecond line\n\n## Notes\nSection first\r\nSection second\n")
}

func TestExportElysiumLayoutRejectsUnsafeTitles(t *testing.T) {
	tests := []string{
		"../Outside",
		"..",
		"Nested/Title",
		`Nested\Title`,
	}

	for _, title := range tests {
		t.Run(title, func(t *testing.T) {
			root := t.TempDir()
			err := ExportElysiumLayout(root, []ImportedItem{
				{
					Kind:         string(project.KindChapter),
					Title:        title,
					BodyMarkdown: "Body.",
				},
			})
			if err == nil {
				t.Fatal("ExportElysiumLayout() error = nil, want unsafe title error")
			}
		})
	}
}

func TestExportElysiumLayoutRoundTripsTypedMetadata(t *testing.T) {
	root := t.TempDir()
	if err := ExportElysiumLayout(root, []ImportedItem{
		{
			Kind:         string(project.KindWritingBrief),
			Title:        "Genre",
			BodyMarkdown: "Fantasy.",
			MetadataJSON: `{"count":2,"enabled":true,"ratio":1.5,"type":"genre","unset":null}`,
		},
	}); err != nil {
		t.Fatalf("ExportElysiumLayout() error = %v", err)
	}

	items, err := ImportElysiumLayout(root)
	if err != nil {
		t.Fatalf("ImportElysiumLayout() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("imported items = %d, want 1", len(items))
	}

	var metadata map[string]any
	if err := json.Unmarshal([]byte(items[0].MetadataJSON), &metadata); err != nil {
		t.Fatalf("metadata json = %q: %v", items[0].MetadataJSON, err)
	}
	if metadata["enabled"] != true {
		t.Fatalf("enabled metadata = %#v, want true", metadata["enabled"])
	}
	if metadata["count"] != float64(2) {
		t.Fatalf("count metadata = %#v, want 2", metadata["count"])
	}
	if metadata["ratio"] != 1.5 {
		t.Fatalf("ratio metadata = %#v, want 1.5", metadata["ratio"])
	}
	if metadata["unset"] != nil {
		t.Fatalf("unset metadata = %#v, want nil", metadata["unset"])
	}
}

func TestExportElysiumLayoutRoundTripsAmbiguousStringMetadata(t *testing.T) {
	root := t.TempDir()
	if err := ExportElysiumLayout(root, []ImportedItem{
		{
			Kind:         string(project.KindWritingBrief),
			Title:        "Genre",
			BodyMarkdown: "Fantasy.",
			MetadataJSON: `{"count":"2","enabled":"true","type":"genre","unset":"null"}`,
		},
	}); err != nil {
		t.Fatalf("ExportElysiumLayout() error = %v", err)
	}

	items, err := ImportElysiumLayout(root)
	if err != nil {
		t.Fatalf("ImportElysiumLayout() error = %v", err)
	}

	var metadata map[string]any
	if err := json.Unmarshal([]byte(items[0].MetadataJSON), &metadata); err != nil {
		t.Fatalf("metadata json = %q: %v", items[0].MetadataJSON, err)
	}
	for key, want := range map[string]string{
		"count":   "2",
		"enabled": "true",
		"unset":   "null",
	} {
		got, ok := metadata[key].(string)
		if !ok || got != want {
			t.Fatalf("metadata[%q] = %#v, want string %q", key, metadata[key], want)
		}
	}
}

func TestExportElysiumLayoutRoundTripsAmbiguousRelationTargets(t *testing.T) {
	root := t.TempDir()
	if err := ExportElysiumLayout(root, []ImportedItem{
		{
			Kind:         string(project.KindStoryBibleEntry),
			Title:        "Flag",
			BodyMarkdown: "Body.",
			MetadataJSON: `{"type":"worldbuilding"}`,
			Relations: []ImportedRelation{
				{TargetTitle: "true", RelationType: "related"},
			},
		},
	}); err != nil {
		t.Fatalf("ExportElysiumLayout() error = %v", err)
	}

	items, err := ImportElysiumLayout(root)
	if err != nil {
		t.Fatalf("ImportElysiumLayout() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("imported items = %d, want 1", len(items))
	}
	if len(items[0].Relations) != 1 || items[0].Relations[0].TargetTitle != "true" {
		t.Fatalf("relations = %#v, want related target true", items[0].Relations)
	}
}

func TestExportElysiumLayoutRoundTripsEmptyStringLists(t *testing.T) {
	root := t.TempDir()
	if err := ExportElysiumLayout(root, []ImportedItem{
		{
			Kind:         string(project.KindWritingBrief),
			Title:        "Genre",
			BodyMarkdown: "Fantasy.",
			MetadataJSON: `{"tags":[],"type":"genre"}`,
		},
	}); err != nil {
		t.Fatalf("ExportElysiumLayout() error = %v", err)
	}

	items, err := ImportElysiumLayout(root)
	if err != nil {
		t.Fatalf("ImportElysiumLayout() error = %v", err)
	}

	var metadata map[string]any
	if err := json.Unmarshal([]byte(items[0].MetadataJSON), &metadata); err != nil {
		t.Fatalf("metadata json = %q: %v", items[0].MetadataJSON, err)
	}
	tags, ok := metadata["tags"].([]any)
	if !ok {
		t.Fatalf("tags metadata = %#v, want empty list", metadata["tags"])
	}
	if len(tags) != 0 {
		t.Fatalf("tags len = %d, want 0", len(tags))
	}
}

func TestExportElysiumLayoutRejectsUnsupportedMetadataValues(t *testing.T) {
	tests := []string{
		`{"nested":{"value":true},"type":"genre"}`,
		`{"values":[true,2,null],"type":"genre"}`,
	}

	for _, metadataJSON := range tests {
		t.Run(metadataJSON, func(t *testing.T) {
			err := ExportElysiumLayout(t.TempDir(), []ImportedItem{
				{
					Kind:         string(project.KindWritingBrief),
					Title:        "Genre",
					BodyMarkdown: "Fantasy.",
					MetadataJSON: metadataJSON,
				},
			})
			if err == nil {
				t.Fatal("ExportElysiumLayout() error = nil, want unsupported metadata error")
			}
		})
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

func assertFile(t *testing.T, root, rel, want string) {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read exported %q: %v", rel, err)
	}
	if string(content) != want {
		t.Fatalf("exported %q =\n%s\nwant:\n%s", rel, string(content), want)
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
