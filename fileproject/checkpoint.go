package fileproject

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type CheckpointFile struct {
	ID     string     `json:"id"`
	Path   string     `json:"path"`
	Kind   LayoutKind `json:"kind"`
	Title  string     `json:"title"`
	SHA256 string     `json:"sha256"`
	Size   int64      `json:"size"`
}

type Checkpoint struct {
	SchemaVersion int              `json:"schemaVersion"`
	ID            string           `json:"id"`
	Message       string           `json:"message"`
	CreatedAt     time.Time        `json:"createdAt"`
	Files         []CheckpointFile `json:"files"`
}

type CreateCheckpointInput struct {
	Message string
}

type CheckpointDiffStatus string

const (
	CheckpointDiffAdded    CheckpointDiffStatus = "added"
	CheckpointDiffDeleted  CheckpointDiffStatus = "deleted"
	CheckpointDiffModified CheckpointDiffStatus = "modified"
)

type CheckpointDiffEntry struct {
	ID         string               `json:"id"`
	Path       string               `json:"path"`
	Status     CheckpointDiffStatus `json:"status"`
	FromSHA256 string               `json:"fromSha256,omitempty"`
	ToSHA256   string               `json:"toSha256,omitempty"`
}

func CreateCheckpoint(root string, input CreateCheckpointInput) (Checkpoint, error) {
	layout, err := Scan(root)
	if err != nil {
		return Checkpoint{}, err
	}
	idMap, stableFiles, err := AssignStableIDs(root, layout)
	if err != nil {
		return Checkpoint{}, err
	}
	if err := WriteIDMap(root, idMap); err != nil {
		return Checkpoint{}, err
	}

	id, err := newCheckpointID(time.Now().UTC())
	if err != nil {
		return Checkpoint{}, err
	}
	checkpoint := Checkpoint{
		SchemaVersion: CurrentSchemaVersion,
		ID:            id,
		Message:       input.Message,
		CreatedAt:     time.Now().UTC(),
		Files:         make([]CheckpointFile, 0, len(stableFiles)),
	}
	baseDir := checkpointDir(root, id)
	for _, file := range stableFiles {
		if err := copyProjectFile(root, file.Path, filepath.Join(baseDir, "files", filepath.FromSlash(file.Path))); err != nil {
			return Checkpoint{}, err
		}
		checkpoint.Files = append(checkpoint.Files, CheckpointFile{
			ID:     file.ID,
			Path:   file.Path,
			Kind:   file.Kind,
			Title:  file.Title,
			SHA256: file.SHA256,
			Size:   file.Size,
		})
	}
	if err := writeCheckpointManifest(root, checkpoint); err != nil {
		return Checkpoint{}, err
	}
	return checkpoint, nil
}

func ListCheckpoints(root string) ([]Checkpoint, error) {
	entries, err := os.ReadDir(filepath.Join(root, ".edda", "checkpoints"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read checkpoints directory: %w", err)
	}
	checkpoints := make([]Checkpoint, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		checkpoint, err := ReadCheckpoint(root, entry.Name())
		if err != nil {
			return nil, err
		}
		checkpoints = append(checkpoints, checkpoint)
	}
	sort.Slice(checkpoints, func(i, j int) bool {
		if checkpoints[i].CreatedAt.Equal(checkpoints[j].CreatedAt) {
			return checkpoints[i].ID < checkpoints[j].ID
		}
		return checkpoints[i].CreatedAt.Before(checkpoints[j].CreatedAt)
	})
	return checkpoints, nil
}

func ReadCheckpoint(root string, id string) (Checkpoint, error) {
	if err := validateFileID(id); err != nil {
		return Checkpoint{}, err
	}
	data, err := os.ReadFile(filepath.Join(checkpointDir(root, id), "manifest.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return Checkpoint{}, fs.ErrNotExist
		}
		return Checkpoint{}, fmt.Errorf("read checkpoint manifest: %w", err)
	}
	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return Checkpoint{}, fmt.Errorf("parse checkpoint manifest: %w", err)
	}
	if checkpoint.ID != id {
		return Checkpoint{}, fmt.Errorf("checkpoint manifest id %q does not match %q", checkpoint.ID, id)
	}
	return checkpoint, nil
}

func DiffCheckpoint(root string, fromID string, toID string) ([]CheckpointDiffEntry, error) {
	from, err := ReadCheckpoint(root, fromID)
	if err != nil {
		return nil, err
	}
	fromFiles := checkpointFileMap(from.Files)
	var toFiles map[string]CheckpointFile
	if toID == "" {
		toFiles, err = workingTreeCheckpointFiles(root)
	} else {
		to, err := ReadCheckpoint(root, toID)
		if err != nil {
			return nil, err
		}
		toFiles = checkpointFileMap(to.Files)
	}
	if err != nil {
		return nil, err
	}
	return diffCheckpointMaps(fromFiles, toFiles), nil
}

func RestoreCheckpoint(root string, id string) (Checkpoint, error) {
	checkpoint, err := ReadCheckpoint(root, id)
	if err != nil {
		return Checkpoint{}, err
	}

	layout, err := Scan(root)
	if err != nil {
		return Checkpoint{}, err
	}
	restoredPaths := map[string]string{}
	for _, file := range checkpoint.Files {
		restoredPaths[file.Path] = file.ID
	}
	for _, file := range layout.Files {
		if !isStableIDFile(file) {
			continue
		}
		if _, keep := restoredPaths[file.Path]; keep {
			continue
		}
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(file.Path))); err != nil && !os.IsNotExist(err) {
			return Checkpoint{}, fmt.Errorf("remove file not present in checkpoint: %w", err)
		}
	}
	for _, file := range checkpoint.Files {
		source := filepath.Join(checkpointDir(root, id), "files", filepath.FromSlash(file.Path))
		target := filepath.Join(root, filepath.FromSlash(file.Path))
		if err := copyFile(source, target); err != nil {
			return Checkpoint{}, err
		}
	}
	if err := WriteIDMap(root, IDMap{SchemaVersion: CurrentSchemaVersion, Items: restoredPaths}); err != nil {
		return Checkpoint{}, err
	}
	layout, err = Scan(root)
	if err != nil {
		return Checkpoint{}, err
	}
	idMap, _, err := AssignStableIDs(root, layout)
	if err != nil {
		return Checkpoint{}, err
	}
	if err := WriteIDMap(root, idMap); err != nil {
		return Checkpoint{}, err
	}
	return checkpoint, nil
}

func writeCheckpointManifest(root string, checkpoint Checkpoint) error {
	dir := checkpointDir(root, checkpoint.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create checkpoint directory: %w", err)
	}
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal checkpoint manifest: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0o644); err != nil {
		return fmt.Errorf("write checkpoint manifest: %w", err)
	}
	return nil
}

func newCheckpointID(now time.Time) (string, error) {
	var data [4]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("generate checkpoint id: %w", err)
	}
	return "checkpoint-" + now.Format("20060102T150405Z") + "-" + hex.EncodeToString(data[:]), nil
}

func checkpointDir(root string, id string) string {
	return filepath.Join(root, ".edda", "checkpoints", id)
}

func copyProjectFile(root string, rel string, target string) error {
	source := filepath.Join(root, filepath.FromSlash(rel))
	return copyFile(source, target)
}

func copyFile(source string, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}
	in, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer in.Close()
	out, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("create target file: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}
	return nil
}

func checkpointFileMap(files []CheckpointFile) map[string]CheckpointFile {
	result := map[string]CheckpointFile{}
	for _, file := range files {
		result[file.ID] = file
	}
	return result
}

func workingTreeCheckpointFiles(root string) (map[string]CheckpointFile, error) {
	layout, err := Scan(root)
	if err != nil {
		return nil, err
	}
	idMap, stableFiles, err := AssignStableIDs(root, layout)
	if err != nil {
		return nil, err
	}
	if err := WriteIDMap(root, idMap); err != nil {
		return nil, err
	}
	files := make([]CheckpointFile, 0, len(stableFiles))
	for _, file := range stableFiles {
		files = append(files, CheckpointFile{
			ID:     file.ID,
			Path:   file.Path,
			Kind:   file.Kind,
			Title:  file.Title,
			SHA256: file.SHA256,
			Size:   file.Size,
		})
	}
	return checkpointFileMap(files), nil
}

func diffCheckpointMaps(fromFiles map[string]CheckpointFile, toFiles map[string]CheckpointFile) []CheckpointDiffEntry {
	ids := map[string]bool{}
	for id := range fromFiles {
		ids[id] = true
	}
	for id := range toFiles {
		ids[id] = true
	}
	orderedIDs := make([]string, 0, len(ids))
	for id := range ids {
		orderedIDs = append(orderedIDs, id)
	}
	sort.Strings(orderedIDs)

	var entries []CheckpointDiffEntry
	for _, id := range orderedIDs {
		from, hasFrom := fromFiles[id]
		to, hasTo := toFiles[id]
		switch {
		case !hasFrom:
			entries = append(entries, CheckpointDiffEntry{ID: id, Path: to.Path, Status: CheckpointDiffAdded, ToSHA256: to.SHA256})
		case !hasTo:
			entries = append(entries, CheckpointDiffEntry{ID: id, Path: from.Path, Status: CheckpointDiffDeleted, FromSHA256: from.SHA256})
		case from.SHA256 != to.SHA256 || from.Path != to.Path:
			path := to.Path
			if path == "" {
				path = from.Path
			}
			entries = append(entries, CheckpointDiffEntry{ID: id, Path: path, Status: CheckpointDiffModified, FromSHA256: from.SHA256, ToSHA256: to.SHA256})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Path == entries[j].Path {
			return entries[i].ID < entries[j].ID
		}
		return entries[i].Path < entries[j].Path
	})
	return entries
}
