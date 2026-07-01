package fileproject

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var (
	ErrInvalidFileID = errors.New("invalid file id")
	ErrFileConflict  = errors.New("saved file hash conflict")
)

var fileIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type DraftMetadata struct {
	SchemaVersion int       `json:"schemaVersion"`
	FileID        string    `json:"fileId"`
	BasePath      string    `json:"basePath"`
	BaseSHA256    string    `json:"baseSha256"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type Draft struct {
	Metadata     DraftMetadata `json:"metadata"`
	BodyMarkdown string        `json:"bodyMarkdown"`
}

type WriteDraftInput struct {
	FileID       string
	BasePath     string
	BaseSHA256   string
	BodyMarkdown string
}

type SaveCanonicalInput struct {
	FileID         string
	BodyMarkdown   string
	ExpectedSHA256 string
}

type SaveDraftInput struct {
	FileID         string
	ExpectedSHA256 string
}

type SavedFile struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

func WriteDraft(root string, input WriteDraftInput) (Draft, error) {
	if err := validateFileID(input.FileID); err != nil {
		return Draft{}, err
	}
	draft := Draft{
		Metadata: DraftMetadata{
			SchemaVersion: CurrentSchemaVersion,
			FileID:        input.FileID,
			BasePath:      filepath.ToSlash(input.BasePath),
			BaseSHA256:    input.BaseSHA256,
			UpdatedAt:     time.Now().UTC(),
		},
		BodyMarkdown: input.BodyMarkdown,
	}

	dir := filepath.Join(root, ".edda", "drafts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Draft{}, fmt.Errorf("create draft directory: %w", err)
	}
	if err := os.WriteFile(draftBodyPath(root, input.FileID), []byte(input.BodyMarkdown), 0o644); err != nil {
		return Draft{}, fmt.Errorf("write draft body: %w", err)
	}
	data, err := json.MarshalIndent(draft.Metadata, "", "  ")
	if err != nil {
		return Draft{}, fmt.Errorf("marshal draft metadata: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(draftMetadataPath(root, input.FileID), data, 0o644); err != nil {
		return Draft{}, fmt.Errorf("write draft metadata: %w", err)
	}
	return draft, nil
}

func ReadDraft(root string, fileID string) (Draft, error) {
	if err := validateFileID(fileID); err != nil {
		return Draft{}, err
	}
	body, err := os.ReadFile(draftBodyPath(root, fileID))
	if err != nil {
		if os.IsNotExist(err) {
			return Draft{}, fs.ErrNotExist
		}
		return Draft{}, fmt.Errorf("read draft body: %w", err)
	}
	data, err := os.ReadFile(draftMetadataPath(root, fileID))
	if err != nil {
		if os.IsNotExist(err) {
			return Draft{}, fs.ErrNotExist
		}
		return Draft{}, fmt.Errorf("read draft metadata: %w", err)
	}
	var metadata DraftMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return Draft{}, fmt.Errorf("parse draft metadata: %w", err)
	}
	if metadata.FileID != fileID {
		return Draft{}, fmt.Errorf("draft metadata file id %q does not match %q", metadata.FileID, fileID)
	}
	return Draft{Metadata: metadata, BodyMarkdown: string(body)}, nil
}

func DeleteDraft(root string, fileID string) error {
	if err := validateFileID(fileID); err != nil {
		return err
	}
	for _, path := range []string{draftBodyPath(root, fileID), draftMetadataPath(root, fileID)} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("delete draft: %w", err)
		}
	}
	return nil
}

func SaveCanonicalFile(root string, input SaveCanonicalInput) (SavedFile, error) {
	if err := validateFileID(input.FileID); err != nil {
		return SavedFile{}, err
	}
	stable, err := ResolveStableFile(root, input.FileID)
	if err != nil {
		return SavedFile{}, err
	}
	if input.ExpectedSHA256 != "" && stable.SHA256 != input.ExpectedSHA256 {
		return SavedFile{}, ErrFileConflict
	}

	targetPath := filepath.Join(root, filepath.FromSlash(stable.Path))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return SavedFile{}, fmt.Errorf("create target directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(targetPath), "."+filepath.Base(targetPath)+".tmp-*")
	if err != nil {
		return SavedFile{}, fmt.Errorf("create temporary file: %w", err)
	}
	tmpPath := tmp.Name()
	removeTmp := true
	defer func() {
		if removeTmp {
			_ = os.Remove(tmpPath)
		}
	}()
	if _, err := tmp.WriteString(input.BodyMarkdown); err != nil {
		_ = tmp.Close()
		return SavedFile{}, fmt.Errorf("write temporary file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return SavedFile{}, fmt.Errorf("close temporary file: %w", err)
	}
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return SavedFile{}, fmt.Errorf("replace saved file: %w", err)
	}
	removeTmp = false

	layout, err := Scan(root)
	if err != nil {
		return SavedFile{}, err
	}
	idMap, files, err := AssignStableIDs(root, layout)
	if err != nil {
		return SavedFile{}, err
	}
	if err := WriteIDMap(root, idMap); err != nil {
		return SavedFile{}, err
	}
	for _, file := range files {
		if file.ID == input.FileID {
			return SavedFile{
				ID:     file.ID,
				Path:   file.Path,
				SHA256: file.SHA256,
				Size:   file.Size,
			}, nil
		}
	}
	return SavedFile{}, fmt.Errorf("saved file %q disappeared from project layout", input.FileID)
}

func PromoteDraft(root string, input SaveDraftInput) (SavedFile, error) {
	draft, err := ReadDraft(root, input.FileID)
	if err != nil {
		return SavedFile{}, err
	}
	expected := input.ExpectedSHA256
	if expected == "" {
		expected = draft.Metadata.BaseSHA256
	}
	saved, err := SaveCanonicalFile(root, SaveCanonicalInput{
		FileID:         input.FileID,
		BodyMarkdown:   draft.BodyMarkdown,
		ExpectedSHA256: expected,
	})
	if err != nil {
		return SavedFile{}, err
	}
	if err := DeleteDraft(root, input.FileID); err != nil {
		return SavedFile{}, err
	}
	return saved, nil
}

func ResolveStableFile(root string, fileID string) (StableFile, error) {
	if err := validateFileID(fileID); err != nil {
		return StableFile{}, err
	}
	layout, err := Scan(root)
	if err != nil {
		return StableFile{}, err
	}
	idMap, files, err := AssignStableIDs(root, layout)
	if err != nil {
		return StableFile{}, err
	}
	if err := WriteIDMap(root, idMap); err != nil {
		return StableFile{}, err
	}
	for _, file := range files {
		if file.ID == fileID {
			return file, nil
		}
	}
	return StableFile{}, fmt.Errorf("file id %q not found", fileID)
}

func HashMarkdown(markdown string) string {
	sum := sha256.Sum256([]byte(markdown))
	return hex.EncodeToString(sum[:])
}

func validateFileID(fileID string) error {
	if !fileIDPattern.MatchString(fileID) {
		return ErrInvalidFileID
	}
	return nil
}

func draftBodyPath(root string, fileID string) string {
	return filepath.Join(root, ".edda", "drafts", fileID+".md")
}

func draftMetadataPath(root string, fileID string) string {
	return filepath.Join(root, ".edda", "drafts", fileID+".json")
}
