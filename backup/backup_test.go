package backup_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/InkyQuill/open-edda/backup"
	"github.com/InkyQuill/open-edda/project"
	"github.com/InkyQuill/open-edda/store"
	"github.com/pressly/goose/v3"
	"os"
	"path/filepath"
	"testing"
)

func TestOnlineBackupRestoresAllHistoryAndRejectsCorruption(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	dbPath := filepath.Join(root, "live.db")
	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(db, "../migrations"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO authors VALUES('a','a@example.invalid','hash','now')"); err != nil {
		t.Fatal(err)
	}
	p, err := project.NewService(db).CreateProject(ctx, project.CreateProjectInput{AuthorID: "a", Title: "Book", StorageMode: "files"})
	if err != nil {
		t.Fatal(err)
	}
	versions, err := project.NewVersionStore(db, root, project.VersionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	defer versions.Close()
	head, err := versions.Version(ctx, "a", p.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{"first", "second"} {
		hash := sha256.Sum256([]byte(body))
		entry := project.TreeEntry{ID: "chapter", Path: "chapter.md", Kind: "file", SHA256: hex.EncodeToString(hash[:]), Bytes: int64(len(body))}
		if err := versions.UploadObject(ctx, "a", p.ID, entry.SHA256, entry.Bytes, bytes.NewBufferString(body)); err != nil {
			t.Fatal(err)
		}
		head, err = versions.Publish(ctx, project.PublishVersionInput{AuthorID: "a", ProjectID: p.ID, ExpectedVersion: head.ID, OperationID: body, Entries: []project.TreeEntry{entry}})
		if err != nil {
			t.Fatal(err)
		}
	}
	saved := filepath.Join(root, "snapshot")
	if err := backup.Create(ctx, dbPath, root, saved); err != nil {
		t.Fatal(err)
	}
	if err := backup.Verify(ctx, saved); err != nil {
		t.Fatal(err)
	}
	// Source stays live; a new publication must not change the saved snapshot.
	if _, err := versions.Publish(ctx, project.PublishVersionInput{AuthorID: "a", ProjectID: p.ID, ExpectedVersion: head.ID, OperationID: "delete", Entries: []project.TreeEntry{}}); err != nil {
		t.Fatal(err)
	}
	restored := filepath.Join(root, "restored")
	if err := backup.Restore(ctx, saved, restored); err != nil {
		t.Fatal(err)
	}
	copyDB, err := store.Open(filepath.Join(restored, "edda.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer copyDB.Close()
	copyStore, err := project.NewVersionStore(copyDB, restored, project.VersionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	defer copyStore.Close()
	copyHead, err := copyStore.Version(ctx, "a", p.ID, "")
	if err != nil || copyHead.ID != head.ID {
		t.Fatal("head not preserved", err)
	}
	history, err := copyStore.History(ctx, "a", p.ID, "")
	if err != nil || len(history.Versions) != 3 {
		t.Fatal("history missing", err)
	}
	for _, v := range history.Versions {
		full, err := copyStore.Version(ctx, "a", p.ID, v.ID)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range full.Entries {
			file, err := copyStore.OpenVersionFile(ctx, "a", p.ID, v.ID, e.ID)
			if err != nil {
				t.Fatal(err)
			}
			file.Close()
		}
	}
	if err := backup.Restore(ctx, saved, restored); err == nil {
		t.Fatal("restore overwrote existing data")
	}
	if err := os.WriteFile(filepath.Join(saved, "objects", head.Entries[0].SHA256), []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := backup.Verify(ctx, saved); err == nil {
		t.Fatal("corruption accepted")
	}
	if err := backup.Restore(ctx, saved, filepath.Join(root, "bad")); err == nil {
		t.Fatal("corruption restored")
	}
	if _, err := os.Stat(filepath.Join(root, "bad")); !os.IsNotExist(err) {
		t.Fatal("partial restore exposed")
	}
}
