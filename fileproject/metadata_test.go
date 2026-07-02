package fileproject

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadMetadataRejectsUnsupportedVersions(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".edda"), 0o755); err != nil {
		t.Fatalf("create .edda: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".edda", "project.json"), []byte(`{"schemaVersion":999,"layoutVersion":1,"id":"project-1","title":"Draft"}`), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	if _, err := ReadMetadata(root); err == nil {
		t.Fatalf("ReadMetadata accepted unsupported schema version")
	}
}

func TestInitMetadataRefusesOverwrite(t *testing.T) {
	root := t.TempDir()
	if _, err := InitMetadata(root, InitMetadataInput{ID: "project-1", Title: "Draft"}); err != nil {
		t.Fatalf("first InitMetadata error = %v", err)
	}
	if _, err := InitMetadata(root, InitMetadataInput{ID: "project-2", Title: "Other"}); err == nil {
		t.Fatalf("second InitMetadata succeeded")
	}
	metadata, err := ReadMetadata(root)
	if err != nil {
		t.Fatalf("ReadMetadata error = %v", err)
	}
	if metadata.ID != "project-1" || metadata.Title != "Draft" {
		t.Fatalf("metadata overwritten: %#v", metadata)
	}
}
