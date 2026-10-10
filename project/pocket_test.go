package project

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

const pocketEditID = "a65eef9e-7318-4777-b2a2-c58a169bfcf6"

func pocketFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/pocket/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func pocketBook(t *testing.T, s *VersionStore, root string, review []byte) []TreeEntry {
	t.Helper()
	return []TreeEntry{{ID: root, Path: root, Kind: "directory"}, uploadTestEntry(t, s, root+"-chapter", root+"/chapter.md", pocketFixture(t, "chapter.md")), uploadTestEntry(t, s, root+"-manifest", root+"/.pocket-editor.json", pocketFixture(t, "manifest.json")), uploadTestEntry(t, s, root+"-review", root+"/chapter.review.json", review)}
}
func pocketRead(t *testing.T, s *VersionStore, v, id string) PocketReview {
	t.Helper()
	r, e := s.ReadPocketReview(context.Background(), "author-1", "project-1", v, id)
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func pocketJSONEqual(t *testing.T, a, b []byte) {
	t.Helper()
	var x, y any
	da, db := json.NewDecoder(bytes.NewReader(a)), json.NewDecoder(bytes.NewReader(b))
	da.UseNumber()
	db.UseNumber()
	if da.Decode(&x) != nil || db.Decode(&y) != nil || !reflect.DeepEqual(x, y) {
		t.Fatalf("JSON differs:\n%s\n%s", a, b)
	}
}

func TestPocketGalleyApplyAtomicScopedAndRetryable(t *testing.T) {
	s, db, _ := newTestVersions(t)
	ctx := context.Background()
	entries := append(pocketBook(t, s, "one", pocketFixture(t, "chapter.review.json")), pocketBook(t, s, "two", pocketFixture(t, "chapter.review.json"))...)
	base := publishTestVersion(t, s, "", "import-books", entries)
	r := pocketRead(t, s, base.ID, "one-chapter")
	if len(r.Edits) != 2 || len(r.Signals) != 2 || r.Edits[0].Resolution.Kind != "resolved" {
		t.Fatalf("unexpected projection: %+v", r)
	}
	input := PocketAction{ExpectedVersion: base.ID, OperationID: "apply-one", EntryID: "one-chapter", Action: "apply", RecordID: pocketEditID}
	saved, err := s.ChangePocketReview(ctx, "author-1", "project-1", input)
	if err != nil {
		t.Fatal(err)
	}
	if versionCount(t, db) != 2 {
		t.Fatal("apply should publish exactly one version")
	}
	if !bytes.Equal(readTestFile(t, s, saved.ID, "one-chapter"), pocketFixture(t, "applied.md")) {
		t.Fatal("UTF-8/CRLF source differs from Galley")
	}
	pocketJSONEqual(t, readTestFile(t, s, saved.ID, "one-review"), pocketFixture(t, "applied.review.json"))
	for _, id := range []string{"one-manifest", "two-chapter", "two-manifest", "two-review"} {
		if !bytes.Equal(readTestFile(t, s, base.ID, id), readTestFile(t, s, saved.ID, id)) {
			t.Fatalf("unrelated %s changed", id)
		}
	}
	// Advance the head, then replay a lost response from the earlier immutable base.
	next := publishTestVersion(t, s, saved.ID, "another-version", saved.Entries)
	replay, err := s.ChangePocketReview(ctx, "author-1", "project-1", input)
	if err != nil || replay.ID != saved.ID {
		t.Fatalf("receipt replay: %s %v", replay.ID, err)
	}
	head, err := s.Version(ctx, "author-1", "project-1", "")
	if err != nil || head.ID != next.ID || versionCount(t, db) != 3 {
		t.Fatal("replay advanced or rolled back head")
	}
	input.Action = "remove"
	if _, err = s.ChangePocketReview(ctx, "author-1", "project-1", input); !errors.Is(err, ErrOperationConflict) {
		t.Fatalf("operation reused for another decision: %v", err)
	}
	input.OperationID = "stale-base"
	if _, err = s.ChangePocketReview(ctx, "author-1", "project-1", input); !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("stale version accepted: %v", err)
	}
	if _, err = s.ReadPocketReview(ctx, "another-author", "project-1", saved.ID, "one-chapter"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("ownership: %v", err)
	}
}

func TestPocketUnknownMetadataAndDecisions(t *testing.T) {
	s, _, _ := newTestVersions(t)
	ctx := context.Background()
	raw, _ := pocketObject(pocketFixture(t, "chapter.review.json"))
	raw["future_metadata"] = json.RawMessage(`{"integer":9007199254740993123,"flags":[null,true]}`)
	var signals []pocketJSON
	if err := json.Unmarshal(raw["signals"], &signals); err != nil {
		t.Fatal(err)
	}
	signals[0]["author_metadata"] = json.RawMessage(`{"name":"Рецензент"}`)
	anchor, _ := pocketObject(signals[0]["anchor"])
	anchor["future_anchor"] = json.RawMessage(`"keep"`)
	_ = pocketSet(signals[0], "anchor", anchor)
	_ = pocketSet(raw, "signals", signals)
	data, _ := json.Marshal(raw)
	base := publishTestVersion(t, s, "", "import", pocketBook(t, s, "one", data))
	saved, err := s.ChangePocketReview(ctx, "author-1", "project-1", PocketAction{ExpectedVersion: base.ID, OperationID: "apply", EntryID: "one-chapter", Action: "apply", RecordID: pocketEditID})
	if err != nil {
		t.Fatal(err)
	}
	after, _ := pocketObject(readTestFile(t, s, saved.ID, "one-review"))
	pocketJSONEqual(t, raw["future_metadata"], after["future_metadata"])
	var result []pocketJSON
	_ = json.Unmarshal(after["signals"], &result)
	pocketJSONEqual(t, signals[0]["author_metadata"], result[0]["author_metadata"])
	resultAnchor, _ := pocketObject(result[0]["anchor"])
	pocketJSONEqual(t, anchor["future_anchor"], resultAnchor["future_anchor"])
	if len(result) != 2 {
		t.Fatal("consumed signal was lost")
	}
	note, err := s.ChangePocketReview(ctx, "author-1", "project-1", PocketAction{ExpectedVersion: saved.ID, OperationID: "note", EntryID: "one-chapter", Action: "note", Note: "Новая заметка"})
	if err != nil {
		t.Fatal(err)
	}
	r := pocketRead(t, s, note.ID, "one-chapter")
	if r.Note != "Новая заметка" || len(r.Edits) != 1 || len(r.Signals) != 2 {
		t.Fatal("note affected records")
	}
	removed, err := s.ChangePocketReview(ctx, "author-1", "project-1", PocketAction{ExpectedVersion: note.ID, OperationID: "remove-signal", EntryID: "one-chapter", Action: "remove", RecordID: r.Signals[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	r = pocketRead(t, s, removed.ID, "one-chapter")
	if len(r.Signals) != 1 || len(r.Edits) != 1 {
		t.Fatal("remove cascaded")
	}
	if !bytes.Equal(readTestFile(t, s, saved.ID, "one-chapter"), readTestFile(t, s, removed.ID, "one-chapter")) {
		t.Fatal("sidecar decision changed source")
	}
}

func TestPocketUnavailableEditsStayVisibleAndCannotApply(t *testing.T) {
	s, db, _ := newTestVersions(t)
	ctx := context.Background()
	original := pocketFixture(t, "chapter.review.json")
	for _, kind := range []string{"stale", "ambiguous", "conflicting"} {
		t.Run(kind, func(t *testing.T) {
			raw, _ := pocketObject(original)
			var edits []pocketJSON
			_ = json.Unmarshal(raw["edits"], &edits)
			anchor, _ := pocketObject(edits[0]["anchor"])
			switch kind {
			case "stale":
				_ = pocketSet(anchor, "selection_sha256", strings.Repeat("0", 64))
			case "ambiguous":
				_ = pocketSet(anchor, "source_sha256", strings.Repeat("0", 64))
				_ = pocketSet(anchor, "prefix", "")
				_ = pocketSet(anchor, "suffix", "")
			case "conflicting":
				copy := pocketJSON{}
				for k, v := range edits[0] {
					copy[k] = v
				}
				_ = pocketSet(copy, "id", "c65eef9e-7318-4777-b2a2-c58a169bfcf6")
				edits = append(edits, copy)
			}
			_ = pocketSet(edits[0], "anchor", anchor)
			_ = pocketSet(raw, "edits", edits)
			data, _ := json.Marshal(raw)
			entries := pocketBook(t, s, kind, data)
			if kind == "ambiguous" {
				entries[1] = uploadTestEntry(t, s, kind+"-chapter", kind+"/chapter.md", []byte("тихо тихо"))
			}
			current, _ := s.Version(ctx, "author-1", "project-1", "")
			base := publishTestVersion(t, s, current.ID, "import-"+kind, entries)
			review := pocketRead(t, s, base.ID, kind+"-chapter")
			if review.Edits[0].Resolution.Kind != kind {
				t.Fatalf("got %s", review.Edits[0].Resolution.Kind)
			}
			count := versionCount(t, db)
			_, err := s.ChangePocketReview(ctx, "author-1", "project-1", PocketAction{ExpectedVersion: base.ID, OperationID: "apply-" + kind, EntryID: kind + "-chapter", Action: "apply", RecordID: pocketEditID})
			if !errors.Is(err, ErrInvalidTree) || versionCount(t, db) != count {
				t.Fatalf("unavailable edit applied: %v", err)
			}
		})
	}
}

func TestPocketIdentityAndMalformedData(t *testing.T) {
	for _, change := range []struct {
		field string
		value any
	}{{"chapter_id", "not-uuid"}, {"source_path", "../chapter.md"}, {"signals", nil}, {"edits", 42}, {"schema_version", 2}} {
		raw, _ := pocketObject(pocketFixture(t, "chapter.review.json"))
		_ = pocketSet(raw, change.field, change.value)
		data, _ := json.Marshal(raw)
		if _, err := decodePocket(data); !errors.Is(err, ErrInvalidTree) {
			t.Fatalf("accepted %s", change.field)
		}
	}
	s, _, _ := newTestVersions(t)
	raw, _ := pocketObject(pocketFixture(t, "chapter.review.json"))
	_ = pocketSet(raw, "source_path", "another.md")
	data, _ := json.Marshal(raw)
	base := publishTestVersion(t, s, "", "wrong-source", pocketBook(t, s, "one", data))
	if _, err := s.ReadPocketReview(context.Background(), "author-1", "project-1", base.ID, "one-chapter"); !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("wrong identity: %v", err)
	}
	entries := pocketBook(t, s, "one", pocketFixture(t, "chapter.review.json"))
	entries = append(entries, uploadTestEntry(t, s, "alternate", "one/chapter.md.review.json", pocketFixture(t, "chapter.review.json")))
	next := publishTestVersion(t, s, base.ID, "duplicate-sidecar", entries)
	if _, err := s.ReadPocketReview(context.Background(), "author-1", "project-1", next.ID, "one-chapter"); !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("ambiguous sidecars: %v", err)
	}
	// Pocket Editor's committed Android fixture is also accepted without rewriting.
	if d, err := decodePocket(pocketFixture(t, "pocket-editor.review.json")); err != nil || len(d.Signals) != 1 || len(d.Edits) != 1 {
		t.Fatalf("Android fixture: %+v %v", d, err)
	}
}

func TestPocketComposeFirstReviewAndUndoRedo(t *testing.T) {
	s, _, _ := newTestVersions(t)
	ctx := context.Background()
	entries := pocketBook(t, s, "one", pocketFixture(t, "chapter.review.json"))[:3]
	base := publishTestVersion(t, s, "", "empty-review", entries)
	r := pocketRead(t, s, base.ID, "one-chapter")
	if len(r.Edits) != 0 || len(r.Signals) != 0 || r.Sidecar.SHA256 != "" {
		t.Fatal("expected virtual empty review")
	}
	from := bytes.Index([]byte(r.Source), []byte("тихо"))
	add := PocketAction{ExpectedVersion: base.ID, OperationID: "new-review", EntryID: "one-chapter", Action: "add", RecordID: pocketEditID, SignalType: "edit", From: from, To: from + len("тихо"), Selected: "тихо", After: "негромко"}
	saved, err := s.ChangePocketReview(ctx, "author-1", "project-1", add)
	if err != nil {
		t.Fatal(err)
	}
	r = pocketRead(t, s, saved.ID, "one-chapter")
	if len(r.Edits) != 1 || r.Edits[0].After != "негромко" {
		t.Fatal("proposal not saved")
	}
	if !bytes.Equal(readTestFile(t, s, base.ID, "one-chapter"), readTestFile(t, s, saved.ID, "one-chapter")) {
		t.Fatal("proposal changed canonical source")
	}
	update := add
	update.Action = "update"
	update.ExpectedVersion = saved.ID
	update.OperationID = "update-proposal"
	update.After = "совсем тихо"
	changed, err := s.ChangePocketReview(ctx, "author-1", "project-1", update)
	if err != nil {
		t.Fatal(err)
	}
	if pocketRead(t, s, changed.ID, "one-chapter").Edits[0].After != "совсем тихо" {
		t.Fatal("proposal update failed")
	}
	undo, err := s.ChangePocketReview(ctx, "author-1", "project-1", PocketAction{ExpectedVersion: changed.ID, OperationID: "undo-update", EntryID: "one-chapter", Action: "undo", ActionVersion: changed.ID})
	if err != nil {
		t.Fatal(err)
	}
	if pocketRead(t, s, undo.ID, "one-chapter").Edits[0].After != "негромко" {
		t.Fatal("update undo failed")
	}
	undoCreation, err := s.ChangePocketReview(ctx, "author-1", "project-1", PocketAction{ExpectedVersion: undo.ID, OperationID: "undo-create", EntryID: "one-chapter", Action: "undo", ActionVersion: saved.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(undoCreation.Entries) != len(base.Entries) {
		t.Fatal("undo must remove newly created sidecar")
	}
	redo, err := s.ChangePocketReview(ctx, "author-1", "project-1", PocketAction{ExpectedVersion: undoCreation.ID, OperationID: "redo-create", EntryID: "one-chapter", Action: "undo", ActionVersion: undoCreation.ID})
	if err != nil {
		t.Fatal(err)
	}
	if pocketRead(t, s, redo.ID, "one-chapter").Edits[0].After != "негромко" {
		t.Fatal("creation redo failed")
	}
}

func TestPocketUndoPreservesOtherFilesAndRefusesSubsequentChapterEdits(t *testing.T) {
	s, _, _ := newTestVersions(t)
	ctx := context.Background()
	base := publishTestVersion(t, s, "", "import", pocketBook(t, s, "one", pocketFixture(t, "chapter.review.json")))
	applied, err := s.ChangePocketReview(ctx, "author-1", "project-1", PocketAction{ExpectedVersion: base.ID, OperationID: "apply", EntryID: "one-chapter", Action: "apply", RecordID: pocketEditID})
	if err != nil {
		t.Fatal(err)
	}
	other := uploadTestEntry(t, s, "kb", "knowledge.md", []byte("Авторская база знаний без обязательных этапов."))
	head := publishTestVersion(t, s, applied.ID, "kb-change", append(applied.Entries, other))
	input := PocketAction{ExpectedVersion: head.ID, OperationID: "undo-apply", EntryID: "one-chapter", Action: "undo", ActionVersion: applied.ID}
	undone, err := s.ChangePocketReview(ctx, "author-1", "project-1", input)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(readTestFile(t, s, undone.ID, "one-chapter"), pocketFixture(t, "chapter.md")) {
		t.Fatal("source not restored")
	}
	pocketJSONEqual(t, readTestFile(t, s, undone.ID, "one-review"), pocketFixture(t, "chapter.review.json"))
	if string(readTestFile(t, s, undone.ID, "kb")) != "Авторская база знаний без обязательных этапов." {
		t.Fatal("undo overwrote KB")
	}
	later := publishTestVersion(t, s, undone.ID, "later", undone.Entries)
	replay, err := s.ChangePocketReview(ctx, "author-1", "project-1", input)
	if err != nil || replay.ID != undone.ID {
		t.Fatalf("undo lost-response receipt: %v", err)
	}
	edited := append([]TreeEntry(nil), later.Entries...)
	for i, e := range edited {
		if e.ID == "one-chapter" {
			edited[i] = uploadTestEntry(t, s, e.ID, e.Path, []byte("Новый авторский текст"))
		}
	}
	remote := publishTestVersion(t, s, later.ID, "author-edits", edited)
	_, err = s.ChangePocketReview(ctx, "author-1", "project-1", PocketAction{ExpectedVersion: remote.ID, OperationID: "unsafe-redo", EntryID: "one-chapter", Action: "undo", ActionVersion: undone.ID})
	if !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("redo overwrote later author changes: %v", err)
	}
}

func TestPocketNewAnnotationsValidateExactBytesAndPreserveAnchorsOnUpdate(t *testing.T) {
	s, _, _ := newTestVersions(t)
	ctx := context.Background()
	base := publishTestVersion(t, s, "", "import", pocketBook(t, s, "one", pocketFixture(t, "chapter.review.json")))
	source := pocketFixture(t, "chapter.md")
	from := bytes.Index(source, []byte("🐦"))
	input := PocketAction{ExpectedVersion: base.ID, OperationID: "new-signal", EntryID: "one-chapter", Action: "add", RecordID: "99b2f145-faa5-4de8-8fb2-050dc805978e", SignalType: "note", From: from, To: from + len("🐦"), Selected: "🐦", Comment: "Заголовок"}
	invalid := input
	invalid.From++
	if _, err := s.ChangePocketReview(ctx, "author-1", "project-1", invalid); !errors.Is(err, ErrInvalidTree) {
		t.Fatal("split UTF-8 accepted")
	}
	invalid = input
	invalid.Selected = "другое"
	if _, err := s.ChangePocketReview(ctx, "author-1", "project-1", invalid); !errors.Is(err, ErrInvalidTree) {
		t.Fatal("mismatched selection accepted")
	}
	added, err := s.ChangePocketReview(ctx, "author-1", "project-1", input)
	if err != nil {
		t.Fatal(err)
	}
	original := pocketRead(t, s, added.ID, "one-chapter").Signals[2]
	input.ExpectedVersion = added.ID
	input.OperationID = "update-signal"
	input.Action = "update"
	input.Comment = "Новый комментарий"
	updated, err := s.ChangePocketReview(ctx, "author-1", "project-1", input)
	if err != nil {
		t.Fatal(err)
	}
	r := pocketRead(t, s, updated.ID, "one-chapter").Signals[2]
	if r.Comment != input.Comment || r.Anchor != original.Anchor || r.ID != original.ID {
		t.Fatal("comment update changed anchor or identity")
	}
}

func TestPocketStartIsExplicitAndPreservesOrdinaryProject(t *testing.T) {
	s, db, _ := newTestVersions(t)
	ctx := context.Background()
	chapter := uploadTestEntry(t, s, "chapter", "chapter.md", pocketFixture(t, "chapter.md"))
	kb := uploadTestEntry(t, s, "kb", "knowledge.md", []byte("KB"))
	base := publishTestVersion(t, s, "", "plain-project", []TreeEntry{chapter, kb})
	if _, err := s.ReadPocketReview(ctx, "author-1", "project-1", base.ID, "chapter"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("plain project unexpectedly has a review")
	}
	if versionCount(t, db) != 1 {
		t.Fatal("reading created a version")
	}
	input := PocketAction{ExpectedVersion: base.ID, OperationID: "start", EntryID: "chapter", Action: "start"}
	started, err := s.ChangePocketReview(ctx, "author-1", "project-1", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(started.Entries) != 3 {
		t.Fatal("start should add only the adjacent manifest")
	}
	r := pocketRead(t, s, started.ID, "chapter")
	if len(r.Edits) != 0 || r.Source != string(pocketFixture(t, "chapter.md")) {
		t.Fatal("start changed chapter")
	}
	replay, err := s.ChangePocketReview(ctx, "author-1", "project-1", input)
	if err != nil || replay.ID != started.ID {
		t.Fatalf("start receipt: %v", err)
	}
	var manifest TreeEntry
	for _, e := range started.Entries {
		if e.Path == ".pocket-editor.json" {
			manifest = e
		}
	}
	raw, _ := pocketObject(readTestFile(t, s, started.ID, manifest.ID))
	raw["author_data"] = json.RawMessage(`{"keep":true}`)
	data, _ := json.Marshal(raw)
	nextEntries := append([]TreeEntry(nil), started.Entries...)
	for i, e := range nextEntries {
		if e.ID == manifest.ID {
			nextEntries[i] = uploadTestEntry(t, s, e.ID, e.Path, data)
		}
	}
	next := publishTestVersion(t, s, started.ID, "metadata", nextEntries)
	enrolled, err := s.ChangePocketReview(ctx, "author-1", "project-1", PocketAction{ExpectedVersion: next.ID, OperationID: "start-kb", EntryID: "kb", Action: "start"})
	if err != nil {
		t.Fatal(err)
	}
	actual, _ := pocketObject(readTestFile(t, s, enrolled.ID, manifest.ID))
	pocketJSONEqual(t, actual["author_data"], raw["author_data"])
	if pocketRead(t, s, enrolled.ID, "chapter").document.ChapterID != r.document.ChapterID {
		t.Fatal("enrolling another file changed existing identity")
	}
	if string(readTestFile(t, s, enrolled.ID, "kb")) != "KB" {
		t.Fatal("KB changed")
	}
}
