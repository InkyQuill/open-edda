package fileproject

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	CurrentSchemaVersion = 1
	CurrentLayoutVersion = 1
)

type ProjectMetadata struct {
	SchemaVersion int    `json:"schemaVersion"`
	LayoutVersion int    `json:"layoutVersion"`
	ID            string `json:"id"`
	Title         string `json:"title"`
	ServerURL     string `json:"serverUrl,omitempty"`
}

type InitMetadataInput struct {
	ID        string
	Title     string
	ServerURL string
}

func ReadMetadata(root string) (ProjectMetadata, error) {
	path := filepath.Join(root, ".edda", "project.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ProjectMetadata{}, fs.ErrNotExist
		}
		return ProjectMetadata{}, fmt.Errorf("read project metadata: %w", err)
	}
	var metadata ProjectMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return ProjectMetadata{}, fmt.Errorf("parse project metadata: %w", err)
	}
	if metadata.ID == "" {
		return ProjectMetadata{}, fmt.Errorf("project metadata id is required")
	}
	if metadata.Title == "" {
		return ProjectMetadata{}, fmt.Errorf("project metadata title is required")
	}
	return metadata, nil
}

func InitMetadata(root string, input InitMetadataInput) (ProjectMetadata, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return ProjectMetadata{}, fmt.Errorf("project title is required")
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		var err error
		id, err = randomProjectID()
		if err != nil {
			return ProjectMetadata{}, err
		}
	}

	metadata := ProjectMetadata{
		SchemaVersion: CurrentSchemaVersion,
		LayoutVersion: CurrentLayoutVersion,
		ID:            id,
		Title:         title,
		ServerURL:     strings.TrimSpace(input.ServerURL),
	}

	eddaDir := filepath.Join(root, ".edda")
	if err := os.MkdirAll(eddaDir, 0o755); err != nil {
		return ProjectMetadata{}, fmt.Errorf("create .edda directory: %w", err)
	}
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return ProjectMetadata{}, fmt.Errorf("marshal project metadata: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(eddaDir, "project.json"), data, 0o644); err != nil {
		return ProjectMetadata{}, fmt.Errorf("write project metadata: %w", err)
	}
	return metadata, nil
}

func randomProjectID() (string, error) {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("generate project id: %w", err)
	}
	return "project-" + hex.EncodeToString(data[:]), nil
}
