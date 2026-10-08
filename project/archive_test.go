package project

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io/fs"
	"testing"
)

type zipMember struct {
	name, body string
	mode       fs.FileMode
}

func testZIP(t *testing.T, members ...zipMember) []byte {
	t.Helper()
	var data bytes.Buffer
	writer := zip.NewWriter(&data)
	for _, member := range members {
		header := &zip.FileHeader{Name: member.name, Method: zip.Deflate}
		if member.mode != 0 {
			header.SetMode(member.mode)
		}
		file, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = file.Write([]byte(member.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}
func TestArchiveStageAndPublish(t *testing.T) {
	s, db, _ := newTestVersions(t)
	data := testZIP(t, zipMember{name: "book/章.md", body: "# 原稿\n"}, zipMember{name: "book/empty/", mode: fs.ModeDir | 0755}, zipMember{name: "cover.bin", body: "\x00\xff"}, zipMember{name: ".env", body: "secret"})
	plan, err := s.StageArchive(context.Background(), "author-1", "project-1", bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if versionCount(t, db) != 0 {
		t.Fatal("staging published a version")
	}
	if len(plan.Entries) != 4 || len(plan.Omitted) != 1 {
		t.Fatalf("unexpected inventory: %+v", plan)
	}
	v := publishTestVersion(t, s, "", "archive-test", plan.Entries)
	for _, e := range v.Entries {
		if e.Path == "book/章.md" && string(readTestFile(t, s, v.ID, e.ID)) != "# 原稿\n" {
			t.Fatal("text changed")
		}
		if e.Path == "cover.bin" && string(readTestFile(t, s, v.ID, e.ID)) != "\x00\xff" {
			t.Fatal("binary changed")
		}
	}
}
func TestArchiveRejectsUnsafeTrees(t *testing.T) {
	for name, members := range map[string][]zipMember{
		"traversal": {{name: "../escape"}}, "absolute": {{name: "/escape"}}, "backslash": {{name: "a\\b"}},
		"link": {{name: "link", body: "outside", mode: fs.ModeSymlink | 0777}}, "duplicate": {{name: "a"}, {name: "a"}},
		"case": {{name: "A"}, {name: "a"}}, "parent": {{name: "a"}, {name: "a/b"}}, "reverse-parent": {{name: "a/b"}, {name: "a"}},
		"excluded-traversal": {{name: ".git/../escape"}},
	} {
		t.Run(name, func(t *testing.T) {
			s, db, _ := newTestVersions(t)
			data := testZIP(t, members...)
			_, err := s.StageArchive(context.Background(), "author-1", "project-1", bytes.NewReader(data), int64(len(data)))
			if !errors.Is(err, ErrInvalidTree) {
				t.Fatalf("got %v", err)
			}
			if versionCount(t, db) != 0 {
				t.Fatal("changed head")
			}
		})
	}
}
func TestArchiveLimitsAndAuthorization(t *testing.T) {
	s, _, _ := newTestVersions(t)
	data := testZIP(t, zipMember{name: "a", body: "12345"})
	if _, err := s.StageArchive(context.Background(), "other", "project-1", bytes.NewReader(data), int64(len(data))); err == nil {
		t.Fatal("unauthorized import")
	}
	s.limits.MaxObjectBytes = 4
	if _, err := s.StageArchive(context.Background(), "author-1", "project-1", bytes.NewReader(data), int64(len(data))); !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("limit: %v", err)
	}
}

func TestArchiveCorruptionAndExpandedLimits(t *testing.T) {
	s, db, _ := newTestVersions(t)
	data := testZIP(t, zipMember{name: "a", body: "content to verify"})
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	offset, err := archive.File[0].DataOffset()
	if err != nil {
		t.Fatal(err)
	}
	data[offset] ^= 0xff
	if _, err := s.StageArchive(context.Background(), "author-1", "project-1", bytes.NewReader(data), int64(len(data))); !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("corrupt ZIP accepted: %v", err)
	}
	if versionCount(t, db) != 0 {
		t.Fatal("corruption changed project")
	}
	data = testZIP(t, zipMember{name: "a/b/c", body: "x"})
	s.limits.MaxEntries = 2
	if _, err := s.StageArchive(context.Background(), "author-1", "project-1", bytes.NewReader(data), int64(len(data))); !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("implicit directories bypassed entry limit: %v", err)
	}
	s.limits.MaxEntries = 10
	s.limits.MaxProjectBytes = 5
	data = testZIP(t, zipMember{name: "a", body: "123"}, zipMember{name: "b", body: "456"})
	if _, err := s.StageArchive(context.Background(), "author-1", "project-1", bytes.NewReader(data), int64(len(data))); !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("expanded bytes bypassed limit: %v", err)
	}
}
