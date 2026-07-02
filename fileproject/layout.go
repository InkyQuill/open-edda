package fileproject

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type LayoutKind string

const (
	LayoutKindStory         LayoutKind = "story"
	LayoutKindCharacter     LayoutKind = "character"
	LayoutKindWorldbuilding LayoutKind = "worldbuilding"
	LayoutKindStoryline     LayoutKind = "storyline"
	LayoutKindDraft         LayoutKind = "draft"
	LayoutKindGuidance      LayoutKind = "guidance"
	LayoutKindSkill         LayoutKind = "skill"
	LayoutKindMetadata      LayoutKind = "metadata"
)

type LayoutFile struct {
	Path   string     `json:"path"`
	Kind   LayoutKind `json:"kind"`
	Title  string     `json:"title"`
	SHA256 string     `json:"sha256"`
	Size   int64      `json:"size"`
}

type LayoutWarning struct {
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

type ProjectLayout struct {
	Root     string           `json:"root"`
	Metadata *ProjectMetadata `json:"metadata,omitempty"`
	Files    []LayoutFile     `json:"files"`
	Warnings []LayoutWarning  `json:"warnings"`
}

var recommendedRoots = []string{"story", "characters", "worldbuilding", "storyline", "drafts"}
var recommendedRootKinds = map[string]LayoutKind{
	"story":         LayoutKindStory,
	"characters":    LayoutKindCharacter,
	"worldbuilding": LayoutKindWorldbuilding,
	"storyline":     LayoutKindStoryline,
	"drafts":        LayoutKindDraft,
}

func Scan(root string) (ProjectLayout, error) {
	info, err := os.Stat(root)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("stat project root: %w", err)
	}
	if !info.IsDir() {
		return ProjectLayout{}, fmt.Errorf("project root is not a directory")
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("absolute project root: %w", err)
	}

	layout := ProjectLayout{Root: absRoot, Files: []LayoutFile{}, Warnings: []LayoutWarning{}}
	if metadata, err := ReadMetadata(absRoot); err == nil {
		layout.Metadata = &metadata
	} else if errors.Is(err, fs.ErrNotExist) {
		layout.Warnings = append(layout.Warnings, LayoutWarning{
			Code:    "missing_metadata",
			Path:    ".edda/project.json",
			Message: ".edda/project.json is missing; run edda init before managing this folder",
		})
	} else {
		return ProjectLayout{}, err
	}

	rootSeen := map[string]bool{}
	rootIndexSeen := map[string]bool{}
	hasLayoutIdentity := false

	err = filepath.WalkDir(absRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}

		if entry.IsDir() {
			if shouldSkipDir(rel) {
				return filepath.SkipDir
			}
			if root := strings.Split(rel, "/")[0]; contains(recommendedRoots, root) {
				rootSeen[root] = true
				hasLayoutIdentity = true
			}
			return nil
		}

		if shouldIgnoreFile(rel) {
			return nil
		}

		kind, ok := classify(rel)
		if !ok {
			return nil
		}
		if kind != LayoutKindMetadata {
			hasLayoutIdentity = true
		}

		if name := filepath.Base(rel); name == "_index.md" {
			root := strings.Split(rel, "/")[0]
			if contains(recommendedRoots, root) && rel == root+"/_index.md" {
				rootIndexSeen[root] = true
			}
		}

		file, err := layoutFile(path, rel, kind)
		if err != nil {
			return err
		}
		layout.Files = append(layout.Files, file)
		return nil
	})
	if err != nil {
		return ProjectLayout{}, fmt.Errorf("scan project layout: %w", err)
	}
	if !hasLayoutIdentity {
		return ProjectLayout{}, fmt.Errorf("project root does not look like an Edda layout")
	}

	for _, root := range recommendedRoots {
		if !rootSeen[root] {
			layout.Warnings = append(layout.Warnings, LayoutWarning{
				Code:    "missing_root",
				Path:    root,
				Message: fmt.Sprintf("recommended %s/ folder is missing", root),
			})
			continue
		}
		if !rootIndexSeen[root] {
			layout.Warnings = append(layout.Warnings, LayoutWarning{
				Code:    "missing_index",
				Path:    root + "/_index.md",
				Message: fmt.Sprintf("recommended %s/_index.md file is missing", root),
			})
		}
	}

	sort.Slice(layout.Files, func(i, j int) bool {
		if layout.Files[i].Kind == layout.Files[j].Kind {
			return layout.Files[i].Path < layout.Files[j].Path
		}
		return layout.Files[i].Kind < layout.Files[j].Kind
	})
	sort.Slice(layout.Warnings, func(i, j int) bool {
		if layout.Warnings[i].Code == layout.Warnings[j].Code {
			return layout.Warnings[i].Path < layout.Warnings[j].Path
		}
		return layout.Warnings[i].Code < layout.Warnings[j].Code
	})

	return layout, nil
}

func CountByKind(files []LayoutFile) map[LayoutKind]int {
	counts := map[LayoutKind]int{}
	for _, file := range files {
		counts[file.Kind]++
	}
	return counts
}

func shouldSkipDir(rel string) bool {
	switch rel {
	case ".git", "node_modules", ".edda/drafts", ".edda/conflicts":
		return true
	}
	if strings.HasPrefix(rel, ".edda/checkpoints/") {
		return true
	}
	return false
}

func shouldIgnoreFile(rel string) bool {
	base := filepath.Base(rel)
	if base == ".DS_Store" {
		return true
	}
	if rel == ".edda/state.local.json" {
		return true
	}
	return false
}

func classify(rel string) (LayoutKind, bool) {
	if rel == ".edda/project.json" {
		return LayoutKindMetadata, true
	}
	if rel == "AGENTS.md" || rel == "BOOTSTRAP.md" {
		return LayoutKindGuidance, true
	}
	if strings.HasPrefix(rel, ".agents/skills/") {
		return LayoutKindSkill, true
	}
	if !strings.HasSuffix(rel, ".md") {
		return "", false
	}
	kind, ok := recommendedRootKinds[strings.Split(rel, "/")[0]]
	return kind, ok
}

func layoutFile(path string, rel string, kind LayoutKind) (LayoutFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LayoutFile{}, err
	}
	sum := sha256.Sum256(data)
	return LayoutFile{
		Path:   rel,
		Kind:   kind,
		Title:  titleFromPath(rel),
		SHA256: hex.EncodeToString(sum[:]),
		Size:   int64(len(data)),
	}, nil
}

func titleFromPath(rel string) string {
	base := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.ReplaceAll(base, "_", " ")
	return strings.TrimSpace(base)
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
