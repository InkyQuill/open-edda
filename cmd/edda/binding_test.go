package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportedFolderTracksRemoteFromNestedDirectory(t *testing.T) {
	server, id, _, _ := importTestServer(t, nil)
	root := t.TempDir()
	nested := filepath.Join(root, "chapters", "one")
	if err := os.MkdirAll(nested, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "local.txt"), []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	runSyncTest(t, "import", root, "--server", server, "--project", id, "--exclude", "local.txt")
	state, err := readCheckout(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Excludes) != 1 || state.Excludes[0] != "local.txt" {
		t.Fatal("lost import exclusions")
	}
	t.Chdir(nested)
	if output := runSyncTest(t, "status"); !strings.Contains(output, id) {
		t.Fatal(output)
	}
	if err := os.WriteFile("new.md", []byte("chapter"), 0600); err != nil {
		t.Fatal(err)
	}
	runSyncTest(t, "send")
	runSyncTest(t, "take")
	updated, err := readCheckout(root)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Base.ID == state.Base.ID {
		t.Fatal("nested send did not update root binding")
	}
	if _, err := os.Stat(filepath.Join(nested, ".edda")); !os.IsNotExist(err) {
		t.Fatal("nested checkout created")
	}
	moved := root + "-moved"
	t.Chdir(filepath.Dir(root))
	if err := os.Rename(root, moved); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(moved) })
	t.Chdir(filepath.Join(moved, "chapters", "one"))
	if output := runSyncTest(t, "status"); !strings.Contains(output, id) {
		t.Fatal("moved folder lost binding", output)
	}
}

func TestDiscoveryStopsAtNestedOrInvalidMetadata(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".edda"), 0700); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "nested")
	if err := os.MkdirAll(filepath.Join(nested, ".edda"), 0700); err != nil {
		t.Fatal(err)
	}
	found, err := findProjectRoot(nested)
	if err != nil || found != nested {
		t.Fatalf("crossed nested boundary: %s %v", found, err)
	}
	broken := filepath.Join(root, "broken")
	if err := os.Mkdir(broken, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, ".edda"), filepath.Join(broken, ".edda")); err != nil {
		t.Fatal(err)
	}
	if _, err := findProjectRoot(broken); err == nil {
		t.Fatal("followed metadata symlink")
	}
}

func TestImportRefusesToReplaceBindingBeforeNetworkMutation(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".edda"), 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".edda", "checkout.json")
	if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	err := run([]string{"import", root, "--server", "https://unreachable.invalid", "--project", "another"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "never replaces") {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "original" {
		t.Fatal("binding replaced")
	}
}
