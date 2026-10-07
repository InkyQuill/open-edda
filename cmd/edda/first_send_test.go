package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFirstSendCreatesAndConnectsProject(t *testing.T) {
	server, _, versions, _ := importTestServer(t, nil)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "chapter.md"), []byte("Chapter\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "private.txt"), []byte("Local only\n"), 0600); err != nil {
		t.Fatal(err)
	}
	output := runSyncTest(t, "send", root, "--server", server, "--title", "Book", "--exclude", "private.txt")
	if !strings.Contains(output, "Created Book") {
		t.Fatal(output)
	}
	state, err := readCheckout(root)
	if err != nil {
		t.Fatal(err)
	}
	head, err := versions.Version(context.Background(), "author", state.Base.ProjectID, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(head.Entries) != 1 || head.Entries[0].Path != "chapter.md" {
		t.Fatalf("unexpected uploaded entries: %+v", head.Entries)
	}
	if len(state.Excludes) != 1 || state.Excludes[0] != "private.txt" {
		t.Fatal("exclusions not persisted")
	}
	if err := os.WriteFile(filepath.Join(root, "chapter.md"), []byte("Revised\n"), 0600); err != nil {
		t.Fatal(err)
	}
	runSyncTest(t, "send", root)
	updated, err := readCheckout(root)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Base.ID == head.ID {
		t.Fatal("second send did not publish changes")
	}
	if err := run([]string{"send", root, "--title", "Another"}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("accepted first-send options on attached folder")
	}
}

func TestFirstSendDoesNotCreateProjectForUnsupportedFolder(t *testing.T) {
	server, _, _, db := importTestServer(t, nil)
	for _, kind := range []string{"symlink-entry", "prototype-metadata", "symlink-root"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			switch kind {
			case "symlink-entry":
				if err := os.Symlink("missing", filepath.Join(root, "link")); err != nil {
					t.Fatal(err)
				}
			case "prototype-metadata":
				if err := os.Mkdir(filepath.Join(root, ".edda"), 0700); err != nil {
					t.Fatal(err)
				}
			case "symlink-root":
				link := filepath.Join(t.TempDir(), "root")
				if err := os.Symlink(root, link); err != nil {
					t.Fatal(err)
				}
				root = link
			}
			var before, after int
			if err := db.QueryRow("SELECT COUNT(*) FROM story_projects").Scan(&before); err != nil {
				t.Fatal(err)
			}
			if err := run([]string{"send", root, "--server", server, "--title", "Must not create"}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
				t.Fatal("accepted unsupported folder")
			}
			if err := db.QueryRow("SELECT COUNT(*) FROM story_projects").Scan(&after); err != nil {
				t.Fatal(err)
			}
			if before != after {
				t.Fatal("created an orphan project before validating the folder")
			}
		})
	}
}
