package fileproject

import "fmt"

type FileVersionTarget struct {
	FileID         string     `json:"fileId"`
	Path           string     `json:"path"`
	Kind           LayoutKind `json:"kind"`
	ExpectedSHA256 string     `json:"expectedSha256"`
	StartByte      *int       `json:"startByte,omitempty"`
	EndByte        *int       `json:"endByte,omitempty"`
}

type CheckpointSummary struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	CreatedAt string `json:"createdAt"`
	FileCount int    `json:"fileCount"`
}

type FileCheckpointSummary struct {
	CheckpointID string `json:"checkpointId"`
	FileID       string `json:"fileId"`
	Message      string `json:"message"`
	CreatedAt    string `json:"createdAt"`
	Path         string `json:"path"`
	SHA256       string `json:"sha256"`
	Size         int64  `json:"size"`
}

func ListStableFiles(root string) ([]StableFile, error) {
	layout, err := Scan(root)
	if err != nil {
		return nil, err
	}
	idMap, files, err := AssignStableIDs(root, layout)
	if err != nil {
		return nil, err
	}
	if err := WriteIDMap(root, idMap); err != nil {
		return nil, err
	}
	return files, nil
}

func ValidateFileVersionTarget(root string, target FileVersionTarget) (StableFile, error) {
	if target.FileID == "" {
		return StableFile{}, fmt.Errorf("file target id is required")
	}
	file, err := ResolveStableFile(root, target.FileID)
	if err != nil {
		return StableFile{}, err
	}
	if target.Path != "" && target.Path != file.Path {
		return StableFile{}, ErrFileConflict
	}
	if target.Kind != "" && target.Kind != file.Kind {
		return StableFile{}, ErrFileConflict
	}
	if target.ExpectedSHA256 != "" && target.ExpectedSHA256 != file.SHA256 {
		return StableFile{}, ErrFileConflict
	}
	if target.StartByte != nil && target.EndByte != nil && *target.StartByte > *target.EndByte {
		return StableFile{}, fmt.Errorf("file target byte range is invalid")
	}
	return file, nil
}

func ListCheckpointSummaries(root string) ([]CheckpointSummary, error) {
	checkpoints, err := ListCheckpoints(root)
	if err != nil {
		return nil, err
	}
	summaries := make([]CheckpointSummary, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		summaries = append(summaries, CheckpointSummary{
			ID:        checkpoint.ID,
			Message:   checkpoint.Message,
			CreatedAt: checkpoint.CreatedAt.Format("2006-01-02T15:04:05Z"),
			FileCount: len(checkpoint.Files),
		})
	}
	return summaries, nil
}

func ListFileCheckpointHistory(root string, fileID string) ([]FileCheckpointSummary, error) {
	if err := validateFileID(fileID); err != nil {
		return nil, err
	}
	checkpoints, err := ListCheckpoints(root)
	if err != nil {
		return nil, err
	}
	var history []FileCheckpointSummary
	for _, checkpoint := range checkpoints {
		for _, file := range checkpoint.Files {
			if file.ID != fileID {
				continue
			}
			history = append(history, FileCheckpointSummary{
				CheckpointID: checkpoint.ID,
				FileID:       file.ID,
				Message:      checkpoint.Message,
				CreatedAt:    checkpoint.CreatedAt.Format("2006-01-02T15:04:05Z"),
				Path:         file.Path,
				SHA256:       file.SHA256,
				Size:         file.Size,
			})
		}
	}
	return history, nil
}
