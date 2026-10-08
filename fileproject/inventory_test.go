package fileproject

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
)

func TestInventoryPreservesFlexibleTreeAndStagesExactBytes(t *testing.T) {
	root := t.TempDir()
	files := map[string][]byte{"project.md": []byte("---\ncustom: 値\n---\r\n"), "kb/персонаж.md": []byte("Персонаж"), "work/draft.txt": {}, "sources/一.txt": []byte("原文"), "translations/01.md": []byte("Перевод"), "timeline.yaml": []byte("version: 1\n"), ".pocket-editor.json": []byte("{}"), ".agents/rules.md": []byte("guidance"), "assets/image.bin": {0, 255, 4}, "a.txt": []byte("flat")}
	for path, body := range files {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"empty", "a", ".git", ".edda", "node_modules"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	inventory, err := ScanInventory(context.Background(), root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(inventory.Problems) != 0 || len(inventory.Excluded) != 4 {
		t.Fatalf("inventory: %+v", inventory)
	}
	staged := t.TempDir()
	if err := StageInventory(context.Background(), inventory, nil, staged); err != nil {
		t.Fatal(err)
	}
	for _, entry := range inventory.Entries {
		if entry.Kind != "file" {
			continue
		}
		got, err := os.ReadFile(filepath.Join(staged, entry.ID))
		if err != nil || !reflect.DeepEqual(got, files[entry.Path]) {
			t.Fatalf("%s differs: %v", entry.Path, err)
		}
	}
	again, err := ScanInventory(context.Background(), root, nil)
	if err != nil || !reflect.DeepEqual(inventory, again) {
		t.Fatal("source inventory changed")
	}
	op, err := ImportOperationID(inventory.Entries)
	if err != nil || !strings.HasPrefix(op, "import_") {
		t.Fatal(op, err)
	}
}

func TestInventoryRejectsLinksSpecialFilesAndCollisions(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "skills")); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(root, "pipe"), 0600); err != nil {
		t.Fatal(err)
	}
	inventory, err := ScanInventory(context.Background(), root, nil)
	if err != nil || len(inventory.Problems) != 1 || len(inventory.Excluded) != 1 {
		t.Fatalf("link/special: %+v %v", inventory, err)
	}
	inventory, err = ScanInventory(context.Background(), root, []string{"skills", "pipe"})
	if err != nil || len(inventory.Problems) != 0 || len(inventory.Excluded) != 2 {
		t.Fatalf("explicit exclude: %+v %v", inventory, err)
	}
	for _, name := range []string{"Chapter.md", "chapter.md"} {
		if err := os.WriteFile(filepath.Join(root, name), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	inventory, err = ScanInventory(context.Background(), root, []string{"skills", "pipe"})
	if err != nil || len(inventory.Problems) != 1 {
		t.Fatalf("collision: %+v %v", inventory, err)
	}
	if _, err := ScanInventory(context.Background(), root, []string{"../outside"}); err == nil {
		t.Fatal("traversal exclusion accepted")
	}
}

func TestStageRejectsChangedAddedRemovedAndReplacedFiles(t *testing.T) {
	for _, change := range []string{"changed", "added", "removed", "link"} {
		t.Run(change, func(t *testing.T) {
			root := t.TempDir()
			name := filepath.Join(root, "draft.md")
			if err := os.WriteFile(name, []byte("before"), 0600); err != nil {
				t.Fatal(err)
			}
			inventory, err := ScanInventory(context.Background(), root, nil)
			if err != nil {
				t.Fatal(err)
			}
			switch change {
			case "changed":
				err = os.WriteFile(name, []byte("after!"), 0600)
			case "added":
				err = os.WriteFile(filepath.Join(root, "new.md"), nil, 0600)
			case "removed":
				err = os.Remove(name)
			case "link":
				if err = os.Remove(name); err == nil {
					err = os.Symlink(filepath.Join(t.TempDir(), "missing"), name)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := StageInventory(context.Background(), inventory, nil, t.TempDir()); err == nil {
				t.Fatal("stale inventory staged")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ScanInventory(ctx, t.TempDir(), nil); err == nil {
		t.Fatal("cancellation ignored")
	}
}
