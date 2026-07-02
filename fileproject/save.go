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
	"sync"
	"time"
)

var (
	ErrInvalidFileID = errors.New("invalid file id")
	ErrFileConflict  = errors.New("saved file hash conflict")
)

var fileIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
var saveLocks sync.Map

type DraftMetadata struct {
	SchemaVersion int       `json:"schemaVersion"`
	FileID        string    `json:"fileId"`
	BasePath      string    `json:"basePath"`
	BaseSHA256    string    `json:"baseSha256"`
	BodySHA256    string    `json:"bodySha256"`
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
			BodySHA256:    HashMarkdown(input.BodyMarkdown),
			UpdatedAt:     time.Now().UTC(),
		},
		BodyMarkdown: input.BodyMarkdown,
	}

	dir := filepath.Join(root, ".edda", "drafts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Draft{}, fmt.Errorf("create draft directory: %w", err)
	}
	data, err := json.MarshalIndent(draft.Metadata, "", "  ")
	if err != nil {
		return Draft{}, fmt.Errorf("marshal draft metadata: %w", err)
	}
	data = append(data, '\n')
	bodyPath := draftBodyPath(root, input.FileID)
	metadataPath := draftMetadataPath(root, input.FileID)
	bodyTmp, err := writeTempFile(filepath.Dir(bodyPath), "draft-body-*.md", []byte(input.BodyMarkdown), 0o644)
	if err != nil {
		return Draft{}, fmt.Errorf("write temporary draft body: %w", err)
	}
	defer os.Remove(bodyTmp)
	metadataTmp, err := writeTempFile(filepath.Dir(metadataPath), "draft-meta-*.json", data, 0o644)
	if err != nil {
		return Draft{}, fmt.Errorf("write temporary draft metadata: %w", err)
	}
	defer os.Remove(metadataTmp)
	if err := os.Rename(metadataTmp, metadataPath); err != nil {
		return Draft{}, fmt.Errorf("write draft metadata: %w", err)
	}
	if err := os.Rename(bodyTmp, bodyPath); err != nil {
		return Draft{}, fmt.Errorf("write draft body: %w", err)
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
	if metadata.BodySHA256 != "" && metadata.BodySHA256 != HashMarkdown(string(body)) {
		return Draft{}, fmt.Errorf("draft body hash does not match metadata")
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
	unlock := lockProjectSave(root)
	defer unlock()

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
	mode := fs.FileMode(0o644)
	if info, err := os.Stat(targetPath); err == nil {
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return SavedFile{}, fmt.Errorf("stat saved file: %w", err)
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
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return SavedFile{}, fmt.Errorf("set temporary file permissions: %w", err)
	}
	if _, err := tmp.WriteString(input.BodyMarkdown); err != nil {
		_ = tmp.Close()
		return SavedFile{}, fmt.Errorf("write temporary file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return SavedFile{}, fmt.Errorf("close temporary file: %w", err)
	}
	if input.ExpectedSHA256 != "" {
		current, err := os.ReadFile(targetPath)
		if err != nil {
			return SavedFile{}, fmt.Errorf("read saved file before replace: %w", err)
		}
		if HashMarkdown(string(current)) != input.ExpectedSHA256 {
			return SavedFile{}, ErrFileConflict
		}
	}
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return SavedFile{}, fmt.Errorf("replace saved file: %w", err)
	}
	removeTmp = false

	return SavedFile{
		ID:     input.FileID,
		Path:   stable.Path,
		SHA256: HashMarkdown(input.BodyMarkdown),
		Size:   int64(len(input.BodyMarkdown)),
	}, nil
}

func lockProjectSave(root string) func() {
	key, err := filepath.Abs(root)
	if err != nil {
		key = root
	}
	value, _ := saveLocks.LoadOrStore(key, &sync.Mutex{})
	mutex := value.(*sync.Mutex)
	mutex.Lock()
	return mutex.Unlock
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
		return saved, fmt.Errorf("promote saved canonical file but failed to delete draft: %w", err)
	}
	return saved, nil
}

func ResolveStableFile(root string, fileID string) (StableFile, error) {
	return resolveStableFile(root, fileID)
}

func resolveStableFile(root string, fileID string) (StableFile, error) {
	if err := validateFileID(fileID); err != nil {
		return StableFile{}, err
	}
	layout, err := Scan(root)
	if err != nil {
		return StableFile{}, err
	}
	_, files, err := AssignStableIDs(root, layout)
	if err != nil {
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

func writeTempFile(dir string, pattern string, data []byte, mode fs.FileMode) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	return tmpPath, nil
}

func draftBodyPath(root string, fileID string) string {
	return filepath.Join(root, ".edda", "drafts", fileID+".md")
}

func draftMetadataPath(root string, fileID string) string {
	return filepath.Join(root, ".edda", "drafts", fileID+".json")
}
