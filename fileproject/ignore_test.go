package fileproject

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestIgnoreInventoryAndTrackedFiles(t *testing.T) {
	root := t.TempDir()
	write := func(name, body string) {
		t.Helper()
		full := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("draft.md", "text")
	write(".remember/log.md", "log")
	write(".creative-writing/transactions/log", "log")
	write(".agents/rules.md", "guidance")
	if err := os.MkdirAll(filepath.Join(root, ".agents/skills"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/outside", filepath.Join(root, ".agents/skills/link")); err != nil {
		t.Fatal(err)
	}
	base, err := ScanInventory(context.Background(), root, nil)
	if err != nil || len(base.Problems) != 0 {
		t.Fatalf("%+v %v", base, err)
	}
	write(".eddaignore", "# local extras\n*.log\n!keep.log\n/cache/\nassets/**/*.tmp\ndraft.md\n")
	write("debug.log", "log")
	write("keep.log", "keep")
	write("cache/file.md", "cache")
	write("assets/deep/a.tmp", "tmp")
	inv, err := ScanInventory(context.Background(), root, nil, base.Entries)
	if err != nil {
		t.Fatal(err)
	}
	paths := map[string]bool{}
	for _, e := range inv.Entries {
		paths[e.Path] = true
	}
	for _, name := range []string{"draft.md", "keep.log", ".eddaignore", ".agents/rules.md"} {
		if !paths[name] {
			t.Fatalf("missing %s", name)
		}
	}
	for _, name := range []string{"debug.log", "cache", "assets/deep/a.tmp", ".remember", ".creative-writing", ".agents/skills"} {
		if paths[name] {
			t.Fatalf("included %s", name)
		}
	}
	if err := StageInventory(context.Background(), inv, nil, t.TempDir()); err != nil {
		t.Fatal(err)
	}
}

func TestIgnorePatterns(t *testing.T) {
	for _, tc := range []struct {
		pattern, name string
		want          bool
	}{
		{"**/*.log", "a.log", true}, {"**/*.log", "a/b.log", true}, {"assets/**/*.tmp", "assets/a.tmp", true}, {"assets/**/*.tmp", "assets/a/b.tmp", true}, {"*.log", "a/b.log", false}, {"a/?[12].md", "a/x2.md", true},
	} {
		if got := ignoreGlob(tc.pattern, tc.name); got != tc.want {
			t.Errorf("%s %s: %v", tc.pattern, tc.name, got)
		}
	}
}

func TestIgnoreRejectsSymlinkAndInvalidRules(t *testing.T) {
	for _, value := range []string{"../outside", "[", "!", "/"} {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, ".eddaignore"), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := ScanInventory(context.Background(), root, nil); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
	root := t.TempDir()
	if err := os.Symlink("missing", filepath.Join(root, ".eddaignore")); err != nil {
		t.Fatal(err)
	}
	if _, err := ScanInventory(context.Background(), root, nil); err == nil {
		t.Fatal("accepted symlink ignore file")
	}
}

func TestExternalLinksAreIgnoredAnywhereWithoutReadingTargets(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret"), []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "nested"), 0700); err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(filepath.Join(root, "nested"), outside)
	if err != nil {
		t.Fatal(err)
	}
	for name, target := range map[string]string{".eddaignore": filepath.Join(outside, "secret"), "absolute": outside, "dangling": filepath.Join(outside, "missing"), "nested/relative": relative, "chain": "absolute/secret"} {
		if err := os.Symlink(target, filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
	inventory, err := ScanInventory(context.Background(), root, nil)
	if err != nil || len(inventory.Problems) != 0 || len(inventory.Excluded) != 5 || len(inventory.Entries) != 1 {
		t.Fatalf("%+v %v", inventory, err)
	}
}
