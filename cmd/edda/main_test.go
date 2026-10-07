package main

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/InkyQuill/open-edda/fileproject"
)

func TestStatusReportsUninitializedLayout(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "alchemist-lite"))

	var stdout bytes.Buffer
	if err := run([]string{"status", root}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("status error = %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"Project: uninitialized Edda folder",
		"Stable IDs: missing",
		"story: 2",
		"character: 2",
		"worldbuilding: 3",
		"skill: 1",
		"missing_metadata",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("status output missing %q:\n%s", want, output)
		}
	}
}

func TestInitCreatesMetadataAndStatusReadsIt(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))

	var initOut bytes.Buffer
	if err := run([]string{"init", root, "--title", "Alchemy Draft", "--id", "project-1"}, &initOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("init error = %v", err)
	}
	if !strings.Contains(initOut.String(), "Initialized Edda project: Alchemy Draft (project-1)") {
		t.Fatalf("init output = %s", initOut.String())
	}

	var statusOut bytes.Buffer
	if err := run([]string{"status", root}, &statusOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("status after init error = %v", err)
	}
	if !strings.Contains(statusOut.String(), "Project: Alchemy Draft (project-1)") {
		t.Fatalf("status output = %s", statusOut.String())
	}
	if strings.Contains(statusOut.String(), "missing_metadata") {
		t.Fatalf("status still reports missing metadata:\n%s", statusOut.String())
	}
}

func TestIDSyncCreatesIDMap(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))

	var stdout bytes.Buffer
	if err := run([]string{"ids", "sync", root}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("ids sync error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Updated .edda/ids.json for 1 files.") {
		t.Fatalf("ids sync output = %s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".edda", "ids.json")); err != nil {
		t.Fatalf("ids.json not created: %v", err)
	}

	var statusOut bytes.Buffer
	if err := run([]string{"status", root}, &statusOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("status error = %v", err)
	}
	if !strings.Contains(statusOut.String(), "Stable IDs: present") {
		t.Fatalf("status output = %s", statusOut.String())
	}
}

func TestSavePromotesDraftToCanonicalFile(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))
	stable := prepareStableCLIFile(t, root)
	if _, err := fileproject.WriteDraft(root, fileproject.WriteDraftInput{
		FileID:       stable.ID,
		BasePath:     stable.Path,
		BaseSHA256:   stable.SHA256,
		BodyMarkdown: "# Chapter 1\n\nPromoted from CLI.\n",
	}); err != nil {
		t.Fatalf("WriteDraft error = %v", err)
	}

	var stdout bytes.Buffer
	if err := run([]string{"save", root, "--id", stable.ID, "--from-draft"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("save --from-draft error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Saved story/chapter-01.md") {
		t.Fatalf("save output = %s", stdout.String())
	}
	data, err := os.ReadFile(filepath.Join(root, "story", "chapter-01.md"))
	if err != nil {
		t.Fatalf("read canonical file: %v", err)
	}
	if string(data) != "# Chapter 1\n\nPromoted from CLI.\n" {
		t.Fatalf("canonical body = %q", string(data))
	}
}

func TestSaveBodyFileRejectsStaleExpectedHash(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))
	stable := prepareStableCLIFile(t, root)
	bodyFile := filepath.Join(t.TempDir(), "body.md")
	if err := os.WriteFile(bodyFile, []byte("stale overwrite"), 0o644); err != nil {
		t.Fatalf("write body file: %v", err)
	}

	err := run(
		[]string{"save", root, "--id", stable.ID, "--body-file", bodyFile, "--expected-sha256", strings.Repeat("0", 64)},
		&bytes.Buffer{},
		&bytes.Buffer{},
	)
	if !errors.Is(err, fileproject.ErrFileConflict) {
		t.Fatalf("save stale error = %v, want ErrFileConflict", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "story", "chapter-01.md"))
	if err != nil {
		t.Fatalf("read canonical file: %v", err)
	}
	if strings.Contains(string(data), "stale overwrite") {
		t.Fatalf("stale save changed canonical file:\n%s", string(data))
	}
}

func TestCheckpointHistoryDiffAndRestoreWorkflow(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))
	var checkpointOut bytes.Buffer
	if err := run([]string{"checkpoint", root, "--message", "base"}, &checkpointOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("checkpoint error = %v", err)
	}
	fields := strings.Fields(checkpointOut.String())
	if len(fields) < 2 || fields[0] != "Checkpoint" {
		t.Fatalf("checkpoint output = %s", checkpointOut.String())
	}
	checkpointID := fields[1]

	var historyOut bytes.Buffer
	if err := run([]string{"history", root}, &historyOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("history error = %v", err)
	}
	if !strings.Contains(historyOut.String(), checkpointID) || !strings.Contains(historyOut.String(), "base") {
		t.Fatalf("history output = %s", historyOut.String())
	}

	if err := os.WriteFile(filepath.Join(root, "story", "chapter-01.md"), []byte("# Chapter 1\n\nChanged for diff.\n"), 0o644); err != nil {
		t.Fatalf("write changed file: %v", err)
	}
	var diffOut bytes.Buffer
	if err := run([]string{"diff", root, "--from", checkpointID}, &diffOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("diff error = %v", err)
	}
	if !strings.Contains(diffOut.String(), "modified story/chapter-01.md") {
		t.Fatalf("diff output = %s", diffOut.String())
	}

	var restoreOut bytes.Buffer
	if err := run([]string{"restore", root, "--checkpoint", checkpointID}, &restoreOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("restore error = %v", err)
	}
	if !strings.Contains(restoreOut.String(), checkpointID) {
		t.Fatalf("restore output = %s", restoreOut.String())
	}
	body, err := os.ReadFile(filepath.Join(root, "story", "chapter-01.md"))
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if strings.Contains(string(body), "Changed for diff") {
		t.Fatalf("restore did not roll back file:\n%s", string(body))
	}
}

func TestNetworkCommandsPreserveLocalWork(t *testing.T) {
	for _, serverURL := range []string{"", "http://127.0.0.1:1"} {
		t.Run("server="+serverURL, func(t *testing.T) {
			root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))
			if _, err := fileproject.InitMetadata(root, fileproject.InitMetadataInput{ID: "project-1", Title: "Existing draft", ServerURL: serverURL}); err != nil {
				t.Fatal(err)
			}
			for _, note := range []string{"First local version", "Second local version"} {
				if err := run([]string{"save", root, note}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
					t.Fatal(err)
				}
			}
			state, err := fileproject.ReadSyncState(root)
			if err != nil {
				t.Fatal(err)
			}
			if len(state.PendingUploads) != 2 {
				t.Fatalf("pending uploads = %d, want 2", len(state.PendingUploads))
			}
			before := snapshotFiles(t, root)
			for _, args := range [][]string{
				{"get", root, "--project", "project-2", "--server", "http://127.0.0.1:1"},
				{"send", root}, {"take", root}, {"send", root},
			} {
				var stdout bytes.Buffer
				err := run(args, &stdout, &bytes.Buffer{})
				if err == nil {
					t.Fatalf("%s error = %v", args[0], err)
				}
				if stdout.Len() != 0 {
					t.Fatalf("%s printed success output: %s", args[0], stdout.String())
				}
				if after := snapshotFiles(t, root); !reflect.DeepEqual(before, after) {
					t.Fatalf("%s changed project files or sync state", args[0])
				}
			}
		})
	}
}

func TestNetworkCommandsDoNotInitializeFolders(t *testing.T) {
	for _, command := range []string{"get", "send", "take"} {
		t.Run(command, func(t *testing.T) {
			parent := t.TempDir()
			root := filepath.Join(parent, "new-project")
			args := []string{command, root}
			if command == "get" {
				args = []string{command, root, "--server", "http://127.0.0.1:1", "--project", "project-1"}
			}
			if err := run(args, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
				t.Fatalf("error = %v", err)
			}
			if _, err := os.Stat(root); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("command created destination: %v", err)
			}
			if len(snapshotFiles(t, parent)) != 0 {
				t.Fatal("command wrote files")
			}
		})
	}
}

func snapshotFiles(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		result[relative] = string(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestSaveCheckpointTreatsBareTextAsMessage(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir root: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	var stdout bytes.Buffer
	if err := run([]string{"save", "Chapter", "polish"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("save checkpoint error = %v", err)
	}
	if !strings.Contains(stdout.String(), "separate from network checkout sends") {
		t.Fatalf("save output = %s", stdout.String())
	}
	checkpoints, err := fileproject.ListCheckpoints(root)
	if err != nil {
		t.Fatalf("ListCheckpoints error = %v", err)
	}
	if len(checkpoints) != 1 || checkpoints[0].Message != "Chapter polish" {
		t.Fatalf("checkpoints = %#v", checkpoints)
	}
}

func TestSaveCheckpointTreatsSingleBareArgAsMessage(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir root: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	var stdout bytes.Buffer
	if err := run([]string{"save", "Quick note"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("save checkpoint error = %v", err)
	}
	checkpoints, err := fileproject.ListCheckpoints(root)
	if err != nil {
		t.Fatalf("ListCheckpoints error = %v", err)
	}
	if len(checkpoints) != 1 || checkpoints[0].Message != "Quick note" {
		t.Fatalf("checkpoints = %#v", checkpoints)
	}
}

func TestConflictsAndResolveWorkflow(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))
	stable := prepareStableCLIFile(t, root)
	if _, err := fileproject.PreserveConflict(root, fileproject.PreserveConflictInput{
		FileID:         stable.ID,
		Path:           stable.Path,
		BaseMarkdown:   "base",
		LocalMarkdown:  "# Chapter 1\n\nLocal.\n",
		ServerMarkdown: "# Chapter 1\n\nServer.\n",
	}); err != nil {
		t.Fatalf("PreserveConflict error = %v", err)
	}

	var conflictsOut bytes.Buffer
	if err := run([]string{"conflicts", root}, &conflictsOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("conflicts error = %v", err)
	}
	if !strings.Contains(conflictsOut.String(), stable.ID) || !strings.Contains(conflictsOut.String(), stable.Path) {
		t.Fatalf("conflicts output = %s", conflictsOut.String())
	}

	var resolveOut bytes.Buffer
	if err := run([]string{"resolve", root, "--id", stable.ID, "--use", "server"}, &resolveOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("resolve error = %v", err)
	}
	if !strings.Contains(resolveOut.String(), "Resolved "+stable.ID) {
		t.Fatalf("resolve output = %s", resolveOut.String())
	}
	body, err := os.ReadFile(filepath.Join(root, "story", "chapter-01.md"))
	if err != nil {
		t.Fatalf("read resolved file: %v", err)
	}
	if string(body) != "# Chapter 1\n\nServer.\n" {
		t.Fatalf("resolved body = %q", string(body))
	}
}

func TestResolveRequiresOneResolutionSource(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))
	stable := prepareStableCLIFile(t, root)

	err := run([]string{"resolve", root, "--id", stable.ID}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "exactly one of --use or --body-file") {
		t.Fatalf("resolve without source error = %v", err)
	}

	err = run([]string{"resolve", root, "--id", stable.ID, "--use", "local", "--body-file", filepath.Join(root, "story", "chapter-01.md")}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "exactly one of --use or --body-file") {
		t.Fatalf("resolve with two sources error = %v", err)
	}
}

func TestFilesAndFilteredHistoryExposeStableHashes(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))
	stable := prepareStableCLIFile(t, root)
	if _, err := fileproject.CreateCheckpoint(root, fileproject.CreateCheckpointInput{Message: "base"}); err != nil {
		t.Fatalf("CreateCheckpoint error = %v", err)
	}

	var filesOut bytes.Buffer
	if err := run([]string{"files", root}, &filesOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("files error = %v", err)
	}
	if !strings.Contains(filesOut.String(), stable.ID) || !strings.Contains(filesOut.String(), stable.SHA256) {
		t.Fatalf("files output = %s", filesOut.String())
	}

	var historyOut bytes.Buffer
	if err := run([]string{"history", root, "--id", stable.ID}, &historyOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("history --id error = %v", err)
	}
	if !strings.Contains(historyOut.String(), stable.ID) {
		t.Fatalf("history --id output missing file id context = %s", historyOut.String())
	}
	if !strings.Contains(historyOut.String(), stable.SHA256) || !strings.Contains(historyOut.String(), "base") {
		t.Fatalf("history --id output = %s", historyOut.String())
	}
}

func TestInitRequiresTitle(t *testing.T) {
	var stderr bytes.Buffer
	err := run([]string{"init", t.TempDir()}, &bytes.Buffer{}, &stderr)
	if err == nil || !strings.Contains(err.Error(), "project title is required") {
		t.Fatalf("init without title error = %v", err)
	}
}

func prepareStableCLIFile(t *testing.T, root string) fileproject.StableFile {
	t.Helper()
	layout, err := fileproject.Scan(root)
	if err != nil {
		t.Fatalf("Scan error = %v", err)
	}
	idMap, files, err := fileproject.AssignStableIDs(root, layout)
	if err != nil {
		t.Fatalf("AssignStableIDs error = %v", err)
	}
	if err := fileproject.WriteIDMap(root, idMap); err != nil {
		t.Fatalf("WriteIDMap error = %v", err)
	}
	for _, file := range files {
		if file.Path == "story/chapter-01.md" {
			return file
		}
	}
	t.Fatalf("stable story file not found")
	return fileproject.StableFile{}
}

func copyFixture(t *testing.T, source string) string {
	t.Helper()
	root := t.TempDir()
	if err := filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(root, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	}); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	return root
}

func TestInitNewDirectoryBeforeAndAfterFlags(t *testing.T) {
	for _, before := range []bool{true, false} {
		root := filepath.Join(t.TempDir(), "new-project")
		args := []string{"init", "--title", "New book", "--id", "new-id", "--server-url", "https://example.invalid"}
		if before {
			args = append([]string{"init", root}, args[1:]...)
		} else {
			args = append(args, root)
		}
		if err := run(args, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		metadata, err := fileproject.ReadMetadata(root)
		if err != nil || metadata.ID != "new-id" || metadata.Title != "New book" || metadata.ServerURL != "https://example.invalid" {
			t.Fatalf("metadata=%#v err=%v", metadata, err)
		}
		var out bytes.Buffer
		if err := run([]string{"status", root}, &out, &bytes.Buffer{}); err != nil || strings.Contains(out.String(), "missing_metadata") {
			t.Fatalf("status=%s err=%v", out.String(), err)
		}
	}
}

func TestSaveReportsCanonicalSuccessWhenDraftCleanupFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))
	_, files, err := fileproject.SyncStableIDs(root)
	if err != nil {
		t.Fatal(err)
	}
	file := files[0]
	_, err = fileproject.WriteDraft(root, fileproject.WriteDraftInput{FileID: file.ID, BasePath: file.Path, BaseSHA256: file.SHA256, BodyMarkdown: "Saved body\n"})
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, ".edda", "drafts")
	if err := os.Chmod(dir, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0700) })
	var out bytes.Buffer
	err = run([]string{"save", root, "--id", file.ID, "--from-draft"}, &out, &bytes.Buffer{})
	if !errors.Is(err, fileproject.ErrDraftCleanup) || !strings.Contains(out.String(), "Saved "+file.Path) {
		t.Fatalf("output=%s err=%v", out.String(), err)
	}
	body, err := os.ReadFile(filepath.Join(root, file.Path))
	if err != nil || string(body) != "Saved body\n" {
		t.Fatalf("body=%q err=%v", body, err)
	}
}
