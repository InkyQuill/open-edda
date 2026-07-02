package fileproject

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type IDMap struct {
	SchemaVersion int               `json:"schemaVersion"`
	Items         map[string]string `json:"items"`
}

type StableFile struct {
	LayoutFile
	ID string `json:"id"`
}

func ReadIDMap(root string) (IDMap, error) {
	path := filepath.Join(root, ".edda", "ids.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return IDMap{}, fs.ErrNotExist
		}
		return IDMap{}, fmt.Errorf("read ids map: %w", err)
	}
	var idMap IDMap
	if err := json.Unmarshal(data, &idMap); err != nil {
		return IDMap{}, fmt.Errorf("parse ids map: %w", err)
	}
	if idMap.SchemaVersion != CurrentSchemaVersion {
		return IDMap{}, fmt.Errorf("unsupported ids map schema version %d", idMap.SchemaVersion)
	}
	if idMap.Items == nil {
		idMap.Items = map[string]string{}
	}
	return idMap, nil
}

func WriteIDMap(root string, idMap IDMap) error {
	if idMap.SchemaVersion == 0 {
		idMap.SchemaVersion = CurrentSchemaVersion
	}
	if idMap.Items == nil {
		idMap.Items = map[string]string{}
	}
	eddaDir := filepath.Join(root, ".edda")
	if err := os.MkdirAll(eddaDir, 0o755); err != nil {
		return fmt.Errorf("create .edda directory: %w", err)
	}
	data, err := json.MarshalIndent(idMap, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal ids map: %w", err)
	}
	data = append(data, '\n')
	path := filepath.Join(eddaDir, "ids.json")
	tmp, err := os.CreateTemp(eddaDir, "ids-*.json")
	if err != nil {
		return fmt.Errorf("create temporary ids map: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temporary ids map: %w", err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temporary ids map: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync temporary ids map: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary ids map: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("write ids map: %w", err)
	}
	return nil
}

func AssignStableIDs(root string, layout ProjectLayout) (IDMap, []StableFile, error) {
	idMap, err := ReadIDMap(root)
	if err != nil {
		if !isNotExist(err) {
			return IDMap{}, nil, err
		}
		idMap = IDMap{SchemaVersion: CurrentSchemaVersion, Items: map[string]string{}}
	}
	if idMap.Items == nil {
		idMap.Items = map[string]string{}
	}

	nextItems := map[string]string{}
	stableFiles := make([]StableFile, 0, len(layout.Files))
	usedIDs := map[string]bool{}
	for _, file := range layout.Files {
		if !isStableIDFile(file) {
			continue
		}
		id := idMap.Items[file.Path]
		if id == "" || usedIDs[id] {
			var err error
			id, err = randomItemID(file.Kind)
			if err != nil {
				return IDMap{}, nil, err
			}
		}
		usedIDs[id] = true
		nextItems[file.Path] = id
		stableFiles = append(stableFiles, StableFile{LayoutFile: file, ID: id})
	}

	sort.Slice(stableFiles, func(i, j int) bool {
		return stableFiles[i].Path < stableFiles[j].Path
	})

	return IDMap{SchemaVersion: CurrentSchemaVersion, Items: nextItems}, stableFiles, nil
}

func isStableIDFile(file LayoutFile) bool {
	switch file.Kind {
	case LayoutKindStory, LayoutKindCharacter, LayoutKindWorldbuilding, LayoutKindStoryline, LayoutKindDraft, LayoutKindGuidance, LayoutKindSkill:
		return true
	default:
		return false
	}
}

func randomItemID(kind LayoutKind) (string, error) {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("generate item id: %w", err)
	}
	prefix := strings.ReplaceAll(string(kind), "_", "-")
	return prefix + "-" + hex.EncodeToString(data[:]), nil
}

func isNotExist(err error) bool {
	return err == fs.ErrNotExist || os.IsNotExist(err)
}
