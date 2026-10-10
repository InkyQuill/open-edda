package project

import (
	"bytes"
	"context"
	"testing"
)

func TestPocketMetadataPreservesBodyAndReviewAnchors(t *testing.T) {
	s, _, _ := newTestVersions(t)
	base := publishTestVersion(t, s, "", "import", pocketBook(t, s, "one", pocketFixture(t, "chapter.review.json")))
	before := pocketRead(t, s, base.ID, "one-chapter")
	input := PocketAction{ExpectedVersion: base.ID, OperationID: "metadata", EntryID: "one-chapter", Action: "metadata", Metadata: "title: Глава 🐦\nstatus: Как решит автор\ncustom: [one, two]"}
	saved, err := s.ChangePocketReview(context.Background(), "author-1", "project-1", input)
	if err != nil {
		t.Fatal(err)
	}
	after := pocketRead(t, s, saved.ID, "one-chapter")
	if !bytes.HasSuffix([]byte(after.Source), []byte(before.Source)) {
		t.Fatal("body bytes changed")
	}
	for i, r := range after.Edits {
		if r.Resolution.Kind != before.Edits[i].Resolution.Kind || r.Anchor.Source != pocketHash([]byte(after.Source)) {
			t.Fatalf("edit lost anchor: %+v", r)
		}
	}
	for i, r := range after.Signals {
		if r.Resolution.Kind != before.Signals[i].Resolution.Kind || r.Anchor.Source != pocketHash([]byte(after.Source)) {
			t.Fatalf("signal lost anchor: %+v", r)
		}
	}
	replay, err := s.ChangePocketReview(context.Background(), "author-1", "project-1", input)
	if err != nil || replay.ID != saved.ID {
		t.Fatalf("receipt replay: %v", err)
	}
	undone, err := s.ChangePocketReview(context.Background(), "author-1", "project-1", PocketAction{ExpectedVersion: saved.ID, OperationID: "undo-metadata", EntryID: "one-chapter", Action: "undo", ActionVersion: saved.ID})
	if err != nil {
		t.Fatal(err)
	}
	restored := pocketRead(t, s, undone.ID, "one-chapter")
	if restored.Source != before.Source {
		t.Fatal("undo did not restore exact source")
	}
}

func TestPlainMetadataKeepsBOMMixedBodyAndNoWorkflowFiles(t *testing.T) {
	for _, source := range []string{"\ufeffТело\r\nещё\n", "\ufeff---\r\ntitle: Старое\r\n...\r\nТело\r\nещё\n", "---\rtitle: Старое\r---\rТело\rещё\n"} {
		t.Run(source, func(t *testing.T) {
			s, _, _ := newTestVersions(t)
			base := publishTestVersion(t, s, "", "import", []TreeEntry{uploadTestEntry(t, s, "text", "text.md", []byte(source))})
			saved, err := s.ChangePocketReview(context.Background(), "author-1", "project-1", PocketAction{ExpectedVersion: base.ID, OperationID: "metadata", EntryID: "text", Action: "metadata", Metadata: "title: Новое"})
			if err != nil {
				t.Fatal(err)
			}
			if len(saved.Entries) != 1 {
				t.Fatal("metadata created workflow files")
			}
			result := readTestFile(t, s, saved.ID, "text")
			body := []byte("Тело\r\nещё\n")
			if source[0] != '\xef' {
				body = []byte("Тело\rещё\n")
			}
			if !bytes.HasSuffix(result, body) {
				t.Fatalf("body changed: %q", result)
			}
			if source[0] == '\xef' && !bytes.HasPrefix(result, []byte{0xef, 0xbb, 0xbf}) {
				t.Fatal("BOM lost")
			}
		})
	}
}
func TestMetadataRefusesInvalidYAMLAndOverlappingReview(t *testing.T) {
	for _, input := range []string{"title: a\ntitle: b", "[one, two]", "title: a\n---\nBody", "title: [", "title: Draft\n--- # scene\nProse", "title: Draft\n--- # second\nother: value", "title: Draft\n--- # empty"} {
		if _, err := replacePocketMetadata([]byte("Body"), &pocketDocument{raw: pocketJSON{}}, input); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
	source := []byte("---\ntitle: Old\n---\nBody")
	d := pocketDocument{raw: pocketJSON{}, Signals: []PocketRecord{{Resolution: PocketResolution{Kind: "resolved", Range: &PocketRange{From: 11, To: 14}}}}}
	if _, err := replacePocketMetadata(source, &d, "title: New"); err == nil {
		t.Fatal("overlapping review silently changed")
	}
	if _, err := replacePocketMetadata([]byte("---\ntitle: Old\nBody"), &pocketDocument{raw: pocketJSON{}}, "title: New"); err == nil {
		t.Fatal("unterminated metadata consumed body")
	}
}

func TestMetadataRefusesProseAndMalformedExistingBlocksWithoutPublishing(t *testing.T) {
	for _, block := range []string{"A scene between two breaks.", "- a\n- b", "title: [", "title: one\ntitle: two", "null", "title: Draft\n--- # scene\nProse before the closing delimiter", "title: Draft\n--- # second\nother: value", "title: Draft\n--- # empty"} {
		t.Run(block, func(t *testing.T) {
			s, db, _ := newTestVersions(t)
			source := []byte("---\n" + block + "\n---\nRest of the chapter\n")
			base := publishTestVersion(t, s, "", "import", []TreeEntry{uploadTestEntry(t, s, "text", "text.md", source)})
			count := versionCount(t, db)
			_, err := s.ChangePocketReview(context.Background(), "author-1", "project-1", PocketAction{ExpectedVersion: base.ID, OperationID: "metadata", EntryID: "text", Action: "metadata", Metadata: "title: New"})
			if err == nil {
				t.Fatal("replaced prose or malformed existing block")
			}
			if versionCount(t, db) != count {
				t.Fatal("refused metadata published a version")
			}
			if !bytes.Equal(readTestFile(t, s, base.ID, "text"), source) {
				t.Fatal("source changed")
			}
		})
	}
}
func TestMetadataAcceptsEmptyAndMappingBlocks(t *testing.T) {
	for _, block := range []string{"", "# comment only", "title: Old\ncustom: [a, b]", "description: |\n  --- # literal text\n  Narrative"} {
		source := []byte("---\n" + block + "\n---\nBody 🐦\r\n")
		next, err := replacePocketMetadata(source, &pocketDocument{raw: pocketJSON{}}, "title: New")
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.HasSuffix(next, []byte("Body 🐦\r\n")) {
			t.Fatal("body bytes changed")
		}
	}
}
