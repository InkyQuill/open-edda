package fileproject

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type ConflictVersion string

const (
	ConflictVersionLocal  ConflictVersion = "local"
	ConflictVersionServer ConflictVersion = "server"
)

type ConflictRecord struct {
	SchemaVersion int        `json:"schemaVersion"`
	FileID        string     `json:"fileId"`
	Path          string     `json:"path"`
	BaseSHA256    string     `json:"baseSha256"`
	LocalSHA256   string     `json:"localSha256"`
	ServerSHA256  string     `json:"serverSha256"`
	DetectedAt    time.Time  `json:"detectedAt"`
	ResolvedAt    *time.Time `json:"resolvedAt,omitempty"`
	Resolution    string     `json:"resolution,omitempty"`
}

type PreserveConflictInput struct {
	FileID         string
	Path           string
	BaseMarkdown   string
	LocalMarkdown  string
	ServerMarkdown string
}

type ResolveConflictInput struct {
	FileID       string
	Use          ConflictVersion
	BodyMarkdown string
	Resolution   string
}

func PreserveConflict(root string, input PreserveConflictInput) (ConflictRecord, error) {
	if err := validateFileID(input.FileID); err != nil {
		return ConflictRecord{}, err
	}
	record := ConflictRecord{
		SchemaVersion: CurrentSchemaVersion,
		FileID:        input.FileID,
		Path:          filepath.ToSlash(input.Path),
		BaseSHA256:    HashMarkdown(input.BaseMarkdown),
		LocalSHA256:   HashMarkdown(input.LocalMarkdown),
		ServerSHA256:  HashMarkdown(input.ServerMarkdown),
		DetectedAt:    time.Now().UTC(),
	}
	dir := conflictDir(root, input.FileID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ConflictRecord{}, fmt.Errorf("create conflict directory: %w", err)
	}
	files := []struct {
		name string
		body string
	}{
		{name: "base.md", body: input.BaseMarkdown},
		{name: "local.md", body: input.LocalMarkdown},
		{name: "server.md", body: input.ServerMarkdown},
	}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(dir, file.name), []byte(file.body), 0o644); err != nil {
			return ConflictRecord{}, fmt.Errorf("write conflict %s: %w", file.name, err)
		}
	}
	if err := writeConflictRecord(root, record); err != nil {
		return ConflictRecord{}, err
	}
	return record, nil
}

func ReadConflict(root string, fileID string) (ConflictRecord, error) {
	if err := validateFileID(fileID); err != nil {
		return ConflictRecord{}, err
	}
	data, err := os.ReadFile(filepath.Join(conflictDir(root, fileID), "metadata.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return ConflictRecord{}, fs.ErrNotExist
		}
		return ConflictRecord{}, fmt.Errorf("read conflict metadata: %w", err)
	}
	var record ConflictRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return ConflictRecord{}, fmt.Errorf("parse conflict metadata: %w", err)
	}
	if record.FileID != fileID {
		return ConflictRecord{}, fmt.Errorf("conflict metadata file id %q does not match %q", record.FileID, fileID)
	}
	return record, nil
}

func ListConflicts(root string) ([]ConflictRecord, error) {
	entries, err := os.ReadDir(filepath.Join(root, ".edda", "conflicts"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read conflicts directory: %w", err)
	}
	records := make([]ConflictRecord, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		record, err := ReadConflict(root, entry.Name())
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, err
		}
		if record.ResolvedAt == nil {
			records = append(records, record)
		}
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].Path == records[j].Path {
			return records[i].FileID < records[j].FileID
		}
		return records[i].Path < records[j].Path
	})
	return records, nil
}

func ResolveConflict(root string, input ResolveConflictInput) (ConflictRecord, SavedFile, error) {
	record, err := ReadConflict(root, input.FileID)
	if err != nil {
		return ConflictRecord{}, SavedFile{}, err
	}
	if record.ResolvedAt != nil {
		return ConflictRecord{}, SavedFile{}, fmt.Errorf("conflict %s already resolved", input.FileID)
	}
	body := input.BodyMarkdown
	resolution := input.Resolution
	if body == "" {
		switch input.Use {
		case ConflictVersionLocal:
			data, err := os.ReadFile(filepath.Join(conflictDir(root, input.FileID), "local.md"))
			if err != nil {
				return ConflictRecord{}, SavedFile{}, fmt.Errorf("read local conflict version: %w", err)
			}
			body = string(data)
			resolution = "local"
		case ConflictVersionServer:
			data, err := os.ReadFile(filepath.Join(conflictDir(root, input.FileID), "server.md"))
			if err != nil {
				return ConflictRecord{}, SavedFile{}, fmt.Errorf("read server conflict version: %w", err)
			}
			body = string(data)
			resolution = "server"
		default:
			return ConflictRecord{}, SavedFile{}, fmt.Errorf("resolve requires local, server, or explicit body")
		}
	}
	saved, err := SaveCanonicalFile(root, SaveCanonicalInput{
		FileID:       input.FileID,
		BodyMarkdown: body,
	})
	if err != nil {
		return ConflictRecord{}, SavedFile{}, err
	}
	now := time.Now().UTC()
	record.ResolvedAt = &now
	if resolution == "" {
		resolution = "body-file"
	}
	record.Resolution = resolution
	if err := writeConflictRecord(root, record); err != nil {
		return ConflictRecord{}, SavedFile{}, err
	}
	return record, saved, nil
}

func writeConflictRecord(root string, record ConflictRecord) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal conflict metadata: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(conflictDir(root, record.FileID), "metadata.json"), data, 0o644); err != nil {
		return fmt.Errorf("write conflict metadata: %w", err)
	}
	return nil
}

func conflictDir(root string, fileID string) string {
	return filepath.Join(root, ".edda", "conflicts", fileID)
}
