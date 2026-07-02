package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.inkyquill.net/inky/writer/fileproject"
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

func TestStatusWriteIDsCreatesIDMap(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))

	var stdout bytes.Buffer
	if err := run([]string{"status", root, "--write-ids"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("status --write-ids error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Updated .edda/ids.json for 1 files.") {
		t.Fatalf("status --write-ids output = %s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".edda", "ids.json")); err != nil {
		t.Fatalf("ids.json not created: %v", err)
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

func TestGetInitializesConnectedProjectAndSyncState(t *testing.T) {
	root := t.TempDir()

	var stdout bytes.Buffer
	if err := run(
		[]string{"get", "--title", "Alchemy Draft", "--id", "project-1", "https://edda.example/projects/project-1", root},
		&stdout,
		&bytes.Buffer{},
	); err != nil {
		t.Fatalf("get error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Connected Alchemy Draft (project-1)") {
		t.Fatalf("get output = %s", stdout.String())
	}
	metadata, err := fileproject.ReadMetadata(root)
	if err != nil {
		t.Fatalf("ReadMetadata error = %v", err)
	}
	if metadata.ServerURL != "https://edda.example/projects/project-1" {
		t.Fatalf("server URL = %q", metadata.ServerURL)
	}
	if _, err := fileproject.ReadSyncState(root); err != nil {
		t.Fatalf("ReadSyncState error = %v", err)
	}
}

func TestSaveCheckpointSendAndTakeWorkflow(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))
	if _, err := fileproject.InitMetadata(root, fileproject.InitMetadataInput{
		ID:        "project-1",
		Title:     "Alchemy Draft",
		ServerURL: "https://edda.example/projects/project-1",
	}); err != nil {
		t.Fatalf("InitMetadata error = %v", err)
	}

	var saveOut bytes.Buffer
	if err := run([]string{"save", root, "Chapter", "polish"}, &saveOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("save checkpoint error = %v", err)
	}
	if !strings.Contains(saveOut.String(), "upload pending") {
		t.Fatalf("save output = %s", saveOut.String())
	}
	state, err := fileproject.ReadSyncState(root)
	if err != nil {
		t.Fatalf("ReadSyncState after save error = %v", err)
	}
	if state.PendingUpload == nil {
		t.Fatalf("pending upload is nil")
	}
	checkpointID := state.PendingUpload.CheckpointID

	var sendOut bytes.Buffer
	if err := run([]string{"send", root}, &sendOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("send error = %v", err)
	}
	if !strings.Contains(sendOut.String(), checkpointID) {
		t.Fatalf("send output = %s", sendOut.String())
	}
	state, err = fileproject.ReadSyncState(root)
	if err != nil {
		t.Fatalf("ReadSyncState after send error = %v", err)
	}
	if state.PendingUpload != nil || state.LastSentCheckpointID != checkpointID {
		t.Fatalf("sync state after send = %#v", state)
	}

	var takeOut bytes.Buffer
	if err := run([]string{"take", root}, &takeOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("take error = %v", err)
	}
	if !strings.Contains(takeOut.String(), "Checked https://edda.example/projects/project-1") {
		t.Fatalf("take output = %s", takeOut.String())
	}
	state, err = fileproject.ReadSyncState(root)
	if err != nil {
		t.Fatalf("ReadSyncState after take error = %v", err)
	}
	if state.LastTakeAt == nil {
		t.Fatalf("last take cursor not recorded")
	}
}

func TestSendRequiresServerURLAndRecordsRetryFailure(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))
	if _, err := fileproject.InitMetadata(root, fileproject.InitMetadataInput{
		ID:    "project-1",
		Title: "Alchemy Draft",
	}); err != nil {
		t.Fatalf("InitMetadata error = %v", err)
	}
	if err := run([]string{"save", root, "Offline checkpoint"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatalf("save checkpoint error = %v", err)
	}

	err := run([]string{"send", root}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "server URL is not configured") {
		t.Fatalf("send without server error = %v", err)
	}
	state, readErr := fileproject.ReadSyncState(root)
	if readErr != nil {
		t.Fatalf("ReadSyncState error = %v", readErr)
	}
	if state.PendingUpload == nil || state.PendingUpload.Attempts != 1 || state.PendingUpload.LastError == "" {
		t.Fatalf("pending upload after failed send = %#v", state.PendingUpload)
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
