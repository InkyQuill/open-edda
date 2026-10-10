package project

import (
	"bytes"
	"context"
	"testing"
)

func TestBookManagementPreservesWireFieldsMissingChaptersAndFiles(t *testing.T) {
	s, _, _ := newTestVersions(t)
	raw := []byte(`{"schema_version":2,"book_id":"1b4f1cad-c846-4551-a497-a745087f5de2","future":9007199254740993,"chapters":[{"id":"0b4f1cad-c846-4551-a497-a745087f5de2","path":"one.md","extra":9007199254740995},{"id":"2b4f1cad-c846-4551-a497-a745087f5de2","path":"missing.md"}]}`)
	base := publishTestVersion(t, s, "", "import", []TreeEntry{uploadTestEntry(t, s, "manifest", ".pocket-editor.json", raw), uploadTestEntry(t, s, "one", "one.md", []byte("Ready text")), uploadTestEntry(t, s, "two", "two.md", []byte("Second"))})
	title := "Новое название"
	input := PocketBookAction{ExpectedVersion: base.ID, OperationID: "book", EntryID: "manifest", Title: "Книга", Chapters: []PocketBookChapter{{ID: "2b4f1cad-c846-4551-a497-a745087f5de2", Path: "missing.md"}, {ID: "0b4f1cad-c846-4551-a497-a745087f5de2", Path: "one.md", Title: &title}, {ID: "3b4f1cad-c846-4551-a497-a745087f5de2", Path: "two.md"}}}
	saved, err := s.ChangePocketBook(context.Background(), "author-1", "project-1", input)
	if err != nil {
		t.Fatal(err)
	}
	body := readTestFile(t, s, saved.ID, "manifest")
	if !bytes.Contains(body, []byte("9007199254740993")) || !bytes.Contains(body, []byte("9007199254740995")) {
		t.Fatal("unknown integer precision lost")
	}
	if bytes.Index(body, []byte("missing.md")) > bytes.Index(body, []byte("one.md")) {
		t.Fatal("order not saved")
	}
	if string(readTestFile(t, s, saved.ID, "one")) != "Ready text" {
		t.Fatal("prose changed")
	}
	replay, err := s.ChangePocketBook(context.Background(), "author-1", "project-1", input)
	if err != nil || replay.ID != saved.ID {
		t.Fatal("replay failed", err)
	}
	input.OperationID = "stale"
	if _, err = s.ChangePocketBook(context.Background(), "author-1", "project-1", input); err == nil {
		t.Fatal("stale write accepted")
	}
	input.ExpectedVersion = saved.ID
	input.OperationID = "duplicate"
	input.Chapters = append(input.Chapters, input.Chapters[0])
	if _, err = s.ChangePocketBook(context.Background(), "author-1", "project-1", input); err == nil {
		t.Fatal("duplicate chapter accepted")
	}
	input.OperationID = "remove"
	input.Chapters = input.Chapters[:1]
	removed, err := s.ChangePocketBook(context.Background(), "author-1", "project-1", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(removed.Entries) != len(saved.Entries) {
		t.Fatal("removing from spine deleted files")
	}
}

func TestBookCreatesAtomicallyAndAdoptsOnlyMatchingReviews(t *testing.T) {
	s, _, _ := newTestVersions(t)
	entries := pocketBook(t, s, "one", pocketFixture(t, "chapter.review.json"))
	entries[2] = uploadTestEntry(t, s, "one-manifest", "one/.pocket-editor.json", []byte(`{"schema_version":2,"book_id":"1b4f1cad-c846-4551-a497-a745087f5de2","chapters":[]}`))
	base := publishTestVersion(t, s, "", "import", entries)
	title := "Новая глава"
	input := PocketBookAction{ExpectedVersion: base.ID, OperationID: "create", EntryID: "one-manifest", Title: "Book", CreatePaths: []string{"new.md"}, Chapters: []PocketBookChapter{{ID: "0b4f1cad-c846-4551-a497-a745087f5de2", Path: "chapter.md"}, {ID: "3b4f1cad-c846-4551-a497-a745087f5de2", Path: "new.md", Title: &title}}}
	saved, err := s.ChangePocketBook(context.Background(), "author-1", "project-1", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.Entries) != len(base.Entries)+1 {
		t.Fatal("chapter not created atomically")
	}
	for _, e := range saved.Entries {
		if e.Path == "one/new.md" && string(readTestFile(t, s, saved.ID, e.ID)) != "# Новая глава\n" {
			t.Fatal("incorrect new chapter")
		}
	}
	review := pocketRead(t, s, saved.ID, "one-chapter")
	if len(review.Edits) == 0 {
		t.Fatal("adopted review lost")
	}
	replay, err := s.ChangePocketBook(context.Background(), "author-1", "project-1", input)
	if err != nil || replay.ID != saved.ID {
		t.Fatal("creation receipt replay failed", err)
	}
	input.OperationID = "wrong-review"
	input.Chapters[0].ID = "5b4f1cad-c846-4551-a497-a745087f5de2"
	if _, err = s.ChangePocketBook(context.Background(), "author-1", "project-1", input); err == nil {
		t.Fatal("adopted unrelated review")
	}
}
