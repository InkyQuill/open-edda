package project

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/InkyQuill/open-edda/store"
	"github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

func newTestVersions(t *testing.T) (*VersionStore, *sql.DB, string) {
	t.Helper()
	db := openMigratedTestDB(t)
	dir := t.TempDir()
	versions, err := NewVersionStore(db, dir, VersionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { versions.Close() })
	return versions, db, dir
}

func objectDigest(body []byte) string { sum := sha256.Sum256(body); return hex.EncodeToString(sum[:]) }
func uploadTestEntry(t *testing.T, s *VersionStore, id, path string, body []byte) TreeEntry {
	t.Helper()
	hash := objectDigest(body)
	if err := s.UploadObject(context.Background(), "author-1", "project-1", hash, int64(len(body)), bytes.NewReader(body)); err != nil {
		t.Fatal(err)
	}
	return TreeEntry{ID: id, Path: path, Kind: "file", SHA256: hash, Bytes: int64(len(body))}
}
func publishTestVersion(t *testing.T, s *VersionStore, base, op string, entries []TreeEntry) ProjectVersion {
	t.Helper()
	v, err := s.Publish(context.Background(), PublishVersionInput{AuthorID: "author-1", ProjectID: "project-1", ExpectedVersion: base, OperationID: op, Entries: entries})
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func versionCount(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT count(*) FROM project_versions WHERE project_id='project-1'").Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
func readTestFile(t *testing.T, s *VersionStore, v, id string) []byte {
	t.Helper()
	r, err := s.OpenVersionFile(context.Background(), "author-1", "project-1", v, id)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestVersionTreeRoundTripRenameDeleteRestoreAndReopen(t *testing.T) {
	s, db, dir := newTestVersions(t)
	ctx := context.Background()
	raw := []byte("---\nunknown: сохраняется\n---\nТекст 日本語\r\n")
	entries := []TreeEntry{
		uploadTestEntry(t, s, "guidance", "project.md", raw),
		{ID: "kb", Path: "kb", Kind: "directory"},
		uploadTestEntry(t, s, "canon", "kb/канон.md", []byte("канон")),
		{ID: "empty", Path: "пустая папка", Kind: "directory"},
		{ID: "sources", Path: "sources", Kind: "directory"},
		uploadTestEntry(t, s, "binary", "sources/original.bin", []byte{0, 255, 1, 2}),
		uploadTestEntry(t, s, "manifest", ".pocket-editor.json", []byte(`{"schema_version":1}`)),
		uploadTestEntry(t, s, "timeline", "timeline.yaml", []byte("unknown: preserved\n")),
	}
	first := publishTestVersion(t, s, "", "first", entries)
	got, err := s.Version(ctx, "author-1", "project-1", "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, first) {
		t.Fatalf("manifest differs: %#v", got)
	}
	if !bytes.Equal(readTestFile(t, s, first.ID, "guidance"), raw) {
		t.Fatal("source bytes changed")
	}
	changed := slices.Clone(first.Entries)
	for i := range changed {
		if changed[i].ID == "canon" {
			changed[i].Path = "kb/renamed.md"
		}
	}
	changed = slices.DeleteFunc(changed, func(e TreeEntry) bool { return e.ID == "binary" })
	second := publishTestVersion(t, s, first.ID, "second", changed)
	if second.ParentID != first.ID {
		t.Fatal("parent not retained")
	}
	if _, err := s.OpenVersionFile(ctx, "author-1", "project-1", second.ID, "binary"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("deleted file = %v", err)
	}
	if !bytes.Equal(readTestFile(t, s, first.ID, "binary"), []byte{0, 255, 1, 2}) {
		t.Fatal("old version lost")
	}
	restored, err := s.Restore(ctx, PublishVersionInput{AuthorID: "author-1", ProjectID: "project-1", ExpectedVersion: second.ID, OperationID: "restore"}, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if restored.ID == first.ID || restored.ParentID != second.ID || !reflect.DeepEqual(restored.Entries, first.Entries) {
		t.Fatal("restore replaced history")
	}
	// Reopen database and object handles: do not rely on in-memory state.
	var seq int
	var name, dbPath string
	if err := db.QueryRow("PRAGMA database_list").Scan(&seq, &name, &dbPath); err != nil {
		t.Fatal(err)
	}
	reopened, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	other, err := NewVersionStore(reopened, dir, VersionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if !bytes.Equal(readTestFile(t, other, restored.ID, "guidance"), raw) || versionCount(t, reopened) != 3 {
		t.Fatal("reopened data differs")
	}
}

func TestVersionIdempotencyAndStaleBase(t *testing.T) {
	s, db, _ := newTestVersions(t)
	ctx := context.Background()
	entries := []TreeEntry{uploadTestEntry(t, s, "a", "a.md", []byte("a")), uploadTestEntry(t, s, "b", "b.md", []byte("b"))}
	first := publishTestVersion(t, s, "", "first", entries)
	second := publishTestVersion(t, s, first.ID, "second", nil)
	slices.Reverse(entries)
	retry := publishTestVersion(t, s, "", "first", entries)
	if retry.ID != first.ID || versionCount(t, db) != 2 {
		t.Fatal("retry duplicated or returned wrong version")
	}
	input := PublishVersionInput{AuthorID: "author-1", ProjectID: "project-1", OperationID: "first", Entries: entries, Message: "different request"}
	if _, err := s.Publish(ctx, input); !errors.Is(err, ErrOperationConflict) {
		t.Fatalf("operation reuse: %v", err)
	}
	input.OperationID = "stale"
	input.Message = ""
	if _, err := s.Publish(ctx, input); !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("stale request: %v", err)
	}
	head, err := s.Version(ctx, "author-1", "project-1", "")
	if err != nil || head.ID != second.ID {
		t.Fatalf("head changed: %v", err)
	}
}

func TestVersionAuthorizationAndObjectGrants(t *testing.T) {
	s, db, _ := newTestVersions(t)
	ctx := context.Background()
	entry := uploadTestEntry(t, s, "private", "private.md", []byte("private"))
	v := publishTestVersion(t, s, "", "first", []TreeEntry{entry})
	if _, err := db.Exec(`INSERT INTO authors VALUES ('other','other@example.test','hash','now');
 INSERT INTO story_projects(id,author_id,title,slug,language,created_at,updated_at) VALUES ('other-project','other','Other','other','en','now','now')`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Version(ctx, "other", "project-1", v.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unauthorized manifest: %v", err)
	}
	if _, err := s.OpenVersionFile(ctx, "other", "project-1", v.ID, entry.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unauthorized bytes: %v", err)
	}
	if err := s.UploadObject(ctx, "other", "project-1", entry.SHA256, entry.Bytes, strings.NewReader("private")); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unauthorized upload: %v", err)
	}
	if _, err := s.Publish(ctx, PublishVersionInput{AuthorID: "other", ProjectID: "other-project", OperationID: "stolen-hash", Entries: []TreeEntry{entry}}); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("known foreign digest granted access: %v", err)
	}
	if versionCount(t, db) != 1 {
		t.Fatal("unauthorized mutation")
	}
}

func TestVersionRejectsInvalidTreesAndPreservesHead(t *testing.T) {
	s, db, _ := newTestVersions(t)
	ctx := context.Background()
	file := uploadTestEntry(t, s, "file", "a.md", []byte("a"))
	base := publishTestVersion(t, s, "", "base", []TreeEntry{file})
	tests := map[string][]TreeEntry{
		"traversal":            {{ID: "bad", Path: "../escape", Kind: "directory"}},
		"absolute":             {{ID: "bad", Path: "/escape", Kind: "directory"}},
		"backslash":            {{ID: "bad", Path: `a\b`, Kind: "directory"}},
		"nul":                  {{ID: "bad", Path: "a\x00b", Kind: "directory"}},
		"empty":                {{ID: "bad", Path: "", Kind: "directory"}},
		"missing-parent":       {{ID: "bad", Path: "kb/text", Kind: "directory"}},
		"file-as-parent":       {file, {ID: "bad", Path: "a.md/child", Kind: "directory"}},
		"symlink":              {{ID: "bad", Path: "link", Kind: "symlink"}},
		"unicode-collision":    {{ID: "a", Path: "é", Kind: "directory"}, {ID: "b", Path: "e\u0301", Kind: "directory"}},
		"duplicate-id":         {{ID: "same", Path: "a", Kind: "directory"}, {ID: "same", Path: "b", Kind: "directory"}},
		"case-collision":       {{ID: "a", Path: "Book", Kind: "directory"}, {ID: "b", Path: "book", Kind: "directory"}},
		"directory-bytes":      {{ID: "bad", Path: "dir", Kind: "directory", Bytes: 1}},
		"identity-kind-change": {{ID: "file", Path: "dir", Kind: "directory"}},
		"windows-device":       {{ID: "bad", Path: "CON.txt", Kind: "directory"}},
	}
	for name, entries := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := s.Publish(ctx, PublishVersionInput{AuthorID: "author-1", ProjectID: "project-1", ExpectedVersion: base.ID, OperationID: name, Entries: entries})
			if !errors.Is(err, ErrInvalidTree) {
				t.Fatalf("error = %v", err)
			}
			if versionCount(t, db) != 1 {
				t.Fatal("invalid manifest committed")
			}
		})
	}
	head, err := s.Version(ctx, "author-1", "project-1", "")
	if err != nil || head.ID != base.ID {
		t.Fatalf("head changed: %v", err)
	}
}

type failedUploadReader struct{}

func (failedUploadReader) Read([]byte) (int, error) { return 0, errors.New("interrupted upload") }

func TestVersionObjectFailuresAndLimits(t *testing.T) {
	s, db, dir := newTestVersions(t)
	ctx := context.Background()
	hash := objectDigest([]byte("abc"))
	for name, reader := range map[string]io.Reader{"short": strings.NewReader("ab"), "long": strings.NewReader("abcd"), "wrong-hash": strings.NewReader("xyz"), "interrupted": failedUploadReader{}} {
		t.Run(name, func(t *testing.T) {
			if err := s.UploadObject(ctx, "author-1", "project-1", hash, 3, reader); err == nil {
				t.Fatal("bad upload accepted")
			}
		})
	}
	files, err := os.ReadDir(filepath.Join(dir, "objects"))
	if err != nil || len(files) != 0 {
		t.Fatalf("failed uploads left files: %v %v", files, err)
	}
	limited, err := NewVersionStore(db, dir, VersionLimits{MaxObjectBytes: 2, MaxProjectBytes: 3, MaxEntries: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer limited.Close()
	if err := limited.UploadObject(ctx, "author-1", "project-1", hash, 3, strings.NewReader("abc")); !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("object limit: %v", err)
	}
	small := uploadTestEntry(t, s, "a", "a", []byte("ab"))
	another := small
	another.ID = "b"
	another.Path = "b"
	if _, err := limited.Publish(ctx, PublishVersionInput{AuthorID: "author-1", ProjectID: "project-1", OperationID: "limit", Entries: []TreeEntry{small, another}}); !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("entry limit: %v", err)
	}
	entry := uploadTestEntry(t, s, "good", "good.bin", []byte("abc"))
	v := publishTestVersion(t, s, "", "base", []TreeEntry{entry})
	if err := os.WriteFile(filepath.Join(dir, "objects", hash), []byte("bad"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.OpenVersionFile(ctx, "author-1", "project-1", v.ID, entry.ID); !errors.Is(err, ErrObjectIntegrity) {
		t.Fatalf("corrupt read: %v", err)
	}
	if _, err := s.Publish(ctx, PublishVersionInput{AuthorID: "author-1", ProjectID: "project-1", ExpectedVersion: v.ID, OperationID: "corrupt", Entries: []TreeEntry{entry}}); !errors.Is(err, ErrObjectIntegrity) {
		t.Fatalf("corrupt publication: %v", err)
	}
	if err := s.UploadObject(ctx, "author-1", "project-1", hash, 3, strings.NewReader("abc")); !errors.Is(err, ErrObjectIntegrity) {
		t.Fatalf("corrupt dedup silently replaced: %v", err)
	}
	if err := os.Remove(filepath.Join(dir, "objects", hash)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.OpenVersionFile(ctx, "author-1", "project-1", v.ID, entry.ID); !errors.Is(err, ErrObjectIntegrity) {
		t.Fatalf("missing read: %v", err)
	}
	if versionCount(t, db) != 1 {
		t.Fatal("failed object changed versions")
	}
}

func TestVersionTransactionRollsBackPartialManifest(t *testing.T) {
	s, db, _ := newTestVersions(t)
	ctx := context.Background()
	base := publishTestVersion(t, s, "", "base", nil)
	if _, err := db.Exec(`CREATE TRIGGER reject_tree_entry BEFORE INSERT ON project_version_entries WHEN NEW.path='z' BEGIN SELECT RAISE(ABORT,'injected write failure'); END`); err != nil {
		t.Fatal(err)
	}
	_, err := s.Publish(ctx, PublishVersionInput{AuthorID: "author-1", ProjectID: "project-1", ExpectedVersion: base.ID, OperationID: "failure", Entries: []TreeEntry{{ID: "a", Path: "a", Kind: "directory"}, {ID: "z", Path: "z", Kind: "directory"}}})
	if err == nil {
		t.Fatal("injected error ignored")
	}
	var ids int
	if err := db.QueryRow("SELECT count(*) FROM project_tree_ids").Scan(&ids); err != nil {
		t.Fatal(err)
	}
	head, err := s.Version(ctx, "author-1", "project-1", "")
	if err != nil || head.ID != base.ID || versionCount(t, db) != 1 || ids != 0 {
		t.Fatalf("partial transaction survived: ids=%d err=%v", ids, err)
	}
}

func TestConcurrentVersionPublicationAndRetry(t *testing.T) {
	for _, same := range []bool{false, true} {
		t.Run(fmt.Sprint(same), func(t *testing.T) {
			s, db, dir := newTestVersions(t)
			base := publishTestVersion(t, s, "", "base", nil)
			var seq int
			var name, dbPath string
			if err := db.QueryRow("PRAGMA database_list").Scan(&seq, &name, &dbPath); err != nil {
				t.Fatal(err)
			}
			otherDB, err := store.Open(dbPath)
			if err != nil {
				t.Fatal(err)
			}
			defer otherDB.Close()
			other, err := NewVersionStore(otherDB, dir, VersionLimits{})
			if err != nil {
				t.Fatal(err)
			}
			defer other.Close()
			start := make(chan struct{})
			results := make(chan error, 2)
			var wg sync.WaitGroup
			for i, target := range []*VersionStore{s, other} {
				wg.Add(1)
				go func(i int, target *VersionStore) {
					defer wg.Done()
					<-start
					op := fmt.Sprintf("op-%d", i)
					if same {
						op = "same"
					}
					_, err := target.Publish(context.Background(), PublishVersionInput{AuthorID: "author-1", ProjectID: "project-1", ExpectedVersion: base.ID, OperationID: op, Entries: []TreeEntry{{ID: "dir", Path: "dir", Kind: "directory"}}})
					results <- err
				}(i, target)
			}
			close(start)
			wg.Wait()
			close(results)
			success, conflict := 0, 0
			for err := range results {
				if err == nil {
					success++
				} else if errors.Is(err, ErrVersionConflict) {
					conflict++
				} else {
					t.Fatal(err)
				}
			}
			if same && success != 2 || !same && (success != 1 || conflict != 1) || versionCount(t, db) != 2 {
				t.Fatalf("success=%d conflicts=%d versions=%d", success, conflict, versionCount(t, db))
			}
		})
	}
}

// A separate process exits inside a SQL trigger (before publication) or just
// after Publish returns. This exercises actual OS/SQLite recovery, not mocks.
func TestVersionCrashHelper(t *testing.T) {
	phase := os.Getenv("EDDA_VERSION_CRASH_PHASE")
	if phase == "" {
		t.Skip("subprocess helper")
	}
	sql.Register("edda-crash", &sqlite3.SQLiteDriver{ConnectHook: func(conn *sqlite3.SQLiteConn) error {
		return conn.RegisterFunc("crash_process", func() int { os.Exit(23); return 0 }, false)
	}})
	db, err := sql.Open("edda-crash", os.Getenv("EDDA_VERSION_CRASH_DB")+"?_foreign_keys=on&_journal_mode=WAL&_synchronous=FULL&_busy_timeout=5000")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	s, err := NewVersionStore(db, os.Getenv("EDDA_VERSION_CRASH_ROOT"), VersionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	entry := uploadTestEntry(t, s, "crash-file", "crash.md", []byte("durable bytes"))
	if phase == "before" {
		if _, err := db.Exec(`CREATE TEMP TRIGGER crash_publish BEFORE UPDATE OF version_id ON project_version_heads BEGIN SELECT crash_process(); END`); err != nil {
			t.Fatal(err)
		}
	}
	publishTestVersion(t, s, os.Getenv("EDDA_VERSION_CRASH_BASE"), "crash-"+phase, []TreeEntry{entry})
	os.Exit(23)
}

func TestVersionCrashRecovery(t *testing.T) {
	for _, phase := range []string{"before", "after"} {
		t.Run(phase, func(t *testing.T) {
			s, db, dir := newTestVersions(t)
			base := publishTestVersion(t, s, "", "base", nil)
			var seq int
			var name, dbPath string
			if err := db.QueryRow("PRAGMA database_list").Scan(&seq, &name, &dbPath); err != nil {
				t.Fatal(err)
			}
			child := exec.Command(os.Args[0], "-test.run=^TestVersionCrashHelper$")
			child.Env = append(os.Environ(), "EDDA_VERSION_CRASH_PHASE="+phase, "EDDA_VERSION_CRASH_DB="+dbPath, "EDDA_VERSION_CRASH_ROOT="+dir, "EDDA_VERSION_CRASH_BASE="+base.ID)
			output, err := child.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 23 {
				t.Fatalf("child: %v %s", err, output)
			}
			head, err := s.Version(context.Background(), "author-1", "project-1", "")
			if err != nil {
				t.Fatal(err)
			}
			if phase == "before" {
				if head.ID != base.ID || versionCount(t, db) != 1 {
					t.Fatal("uncommitted tree exposed")
				}
			} else {
				if head.ID == base.ID || versionCount(t, db) != 2 {
					t.Fatal("committed version lost")
				}
			}
			entry := TreeEntry{ID: "crash-file", Path: "crash.md", Kind: "file", SHA256: objectDigest([]byte("durable bytes")), Bytes: 13}
			retry := publishTestVersion(t, s, base.ID, "crash-"+phase, []TreeEntry{entry})
			if phase == "after" && retry.ID != head.ID {
				t.Fatal("retry duplicated committed version")
			}
			if string(readTestFile(t, s, retry.ID, "crash-file")) != "durable bytes" || versionCount(t, db) != 2 {
				t.Fatal("recovery lost bytes or duplicated history")
			}
		})
	}
}

func TestVersionRejectsUnsafeObjectDirectoryAndLinks(t *testing.T) {
	db := openMigratedTestDB(t)
	root := t.TempDir()
	external := t.TempDir()
	if err := os.Symlink(external, filepath.Join(root, "objects")); err != nil {
		t.Fatal(err)
	}
	if s, err := NewVersionStore(db, root, VersionLimits{}); err == nil {
		s.Close()
		t.Fatal("symlink object directory accepted")
	}
	s, _, dir := newTestVersions(t)
	hash := objectDigest([]byte("outside"))
	outside := filepath.Join(external, "content")
	if err := os.WriteFile(outside, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "objects", hash)); err != nil {
		t.Fatal(err)
	}
	if err := s.UploadObject(context.Background(), "author-1", "project-1", hash, 7, strings.NewReader("outside")); !errors.Is(err, ErrObjectIntegrity) {
		t.Fatalf("symlink object accepted: %v", err)
	}
}

func TestVersionCancellationAndStorageFailurePreserveHead(t *testing.T) {
	s, db, dir := newTestVersions(t)
	base := publishTestVersion(t, s, "", "base", nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.Publish(ctx, PublishVersionInput{AuthorID: "author-1", ProjectID: "project-1", ExpectedVersion: base.ID, OperationID: "cancel"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled publish: %v", err)
	}
	// Removing the empty open object directory makes staging fail deterministically
	// without depending on whether this test runs as root or on filesystem quotas.
	if err := os.Remove(filepath.Join(dir, "objects")); err != nil {
		t.Fatal(err)
	}
	if err := s.UploadObject(context.Background(), "author-1", "project-1", objectDigest([]byte("x")), 1, strings.NewReader("x")); err == nil {
		t.Fatal("unavailable storage accepted upload")
	}
	if versionCount(t, db) != 1 {
		t.Fatal("failed storage changed history")
	}
}

func TestVersionProjectLimitAndZeroByteFiles(t *testing.T) {
	s, db, dir := newTestVersions(t)
	empty := uploadTestEntry(t, s, "empty", "empty.md", nil)
	v := publishTestVersion(t, s, "", "zero", []TreeEntry{empty})
	if len(readTestFile(t, s, v.ID, "empty")) != 0 {
		t.Fatal("zero byte file changed")
	}
	limited, err := NewVersionStore(db, dir, VersionLimits{MaxObjectBytes: 2, MaxProjectBytes: 3, MaxEntries: 10})
	if err != nil {
		t.Fatal(err)
	}
	defer limited.Close()
	file := uploadTestEntry(t, s, "a", "a", []byte("ab"))
	other := file
	other.ID = "b"
	other.Path = "b"
	_, err = limited.Publish(context.Background(), PublishVersionInput{AuthorID: "author-1", ProjectID: "project-1", ExpectedVersion: v.ID, OperationID: "too-many-bytes", Entries: []TreeEntry{file, other}})
	if !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("project bytes limit: %v", err)
	}
}

func TestVersionReadersSeeWholePinnedTrees(t *testing.T) {
	s, _, _ := newTestVersions(t)
	head := publishTestVersion(t, s, "", "base", nil)
	done := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		for {
			select {
			case <-done:
				result <- nil
				return
			default:
			}
			v, err := s.Version(context.Background(), "author-1", "project-1", "")
			if err != nil {
				result <- err
				return
			}
			want := 0
			switch v.OperationID {
			case "base":
			default:
				if _, err := fmt.Sscanf(v.Message, "entries=%d", &want); err != nil {
					result <- err
					return
				}
			}
			if len(v.Entries) != want {
				result <- fmt.Errorf("partial manifest for %s: %d of %d", v.ID, len(v.Entries), want)
				return
			}
		}
	}()
	defer func() {
		close(done)
		if err := <-result; err != nil {
			t.Error(err)
		}
	}()
	for n := 1; n <= 12; n++ {
		entries := make([]TreeEntry, n)
		for i := range entries {
			entries[i] = TreeEntry{ID: fmt.Sprintf("dir-%d", i), Path: fmt.Sprintf("dir-%d", i), Kind: "directory"}
		}
		var err error
		head, err = s.Publish(context.Background(), PublishVersionInput{AuthorID: "author-1", ProjectID: "project-1", ExpectedVersion: head.ID, OperationID: fmt.Sprintf("op-%d", n), Message: fmt.Sprintf("entries=%d", n), Entries: entries})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestVersionRestoreReceiptDistinguishesPublication(t *testing.T) {
	s, db, _ := newTestVersions(t)
	ctx := context.Background()
	first := publishTestVersion(t, s, "", "first", nil)
	second := publishTestVersion(t, s, first.ID, "second", []TreeEntry{{ID: "d", Path: "d", Kind: "directory"}})
	input := PublishVersionInput{AuthorID: "author-1", ProjectID: "project-1", ExpectedVersion: second.ID, OperationID: "restore"}
	restored, err := s.Restore(ctx, input, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := s.Restore(ctx, input, first.ID)
	if err != nil || retry.ID != restored.ID || versionCount(t, db) != 3 {
		t.Fatalf("restore retry: %v", err)
	}
	if _, err := s.Publish(ctx, input); !errors.Is(err, ErrOperationConflict) {
		t.Fatalf("publish reused restore operation: %v", err)
	}
}

func TestVersionRefusesWeakDatabaseDurability(t *testing.T) {
	s, db, _ := newTestVersions(t)
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		t.Fatal(err)
	}
	_, err := s.Publish(context.Background(), PublishVersionInput{AuthorID: "author-1", ProjectID: "project-1", OperationID: "weak"})
	if err == nil {
		t.Fatal("weak durability acknowledged")
	}
	if versionCount(t, db) != 0 {
		t.Fatal("weak publication survived")
	}
}

func TestVersionConcurrentObjectDeduplication(t *testing.T) {
	s, db, dir := newTestVersions(t)
	start := make(chan struct{})
	results := make(chan error, 4)
	body := []byte("same immutable content")
	hash := objectDigest(body)
	for range 4 {
		go func() {
			<-start
			results <- s.UploadObject(context.Background(), "author-1", "project-1", hash, int64(len(body)), bytes.NewReader(body))
		}()
	}
	close(start)
	for range 4 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	files, err := os.ReadDir(filepath.Join(dir, "objects"))
	if err != nil || len(files) != 1 || files[0].Name() != hash {
		t.Fatalf("objects = %v err=%v", files, err)
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM project_objects").Scan(&count); err != nil || count != 1 {
		t.Fatalf("refs=%d err=%v", count, err)
	}
}

func TestVersionMigrationPreservesLegacyContent(t *testing.T) {
	db := openMigratedTestDB(t)
	legacy := NewService(db)
	content, err := legacy.CreateContent(context.Background(), CreateContentInput{ProjectID: "project-1", Kind: KindChapter, Title: "Legacy", BodyMarkdown: "before", MetadataJSON: "{}", CreatedBy: "author"})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := legacy.UpdateContent(context.Background(), UpdateContentInput{ProjectID: "project-1", ContentID: content.ID, ExpectedRevision: content.CurrentRevision, BodyMarkdown: "retained", MetadataJSON: "{}", CreatedBy: "author"})
	if err != nil {
		t.Fatal(err)
	}
	if err := goose.DownTo(db, filepath.Join("..", "migrations"), 6); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(db, filepath.Join("..", "migrations")); err != nil {
		t.Fatal(err)
	}
	current, err := legacy.GetContent(context.Background(), "project-1", content.ID)
	if err != nil || !reflect.DeepEqual(current, updated) {
		t.Fatalf("legacy content changed: %v", err)
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM revisions WHERE content_item_id=?", content.ID).Scan(&count); err != nil || count != 2 {
		t.Fatalf("legacy revisions=%d err=%v", count, err)
	}
	s, err := NewVersionStore(db, t.TempDir(), VersionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	publishTestVersion(t, s, "", "new-tree", nil)
	current, err = legacy.GetContent(context.Background(), "project-1", content.ID)
	if err != nil || !reflect.DeepEqual(current, updated) {
		t.Fatalf("file publication rewrote legacy content: %v", err)
	}
}
