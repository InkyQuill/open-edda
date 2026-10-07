package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/InkyQuill/open-edda/fileproject"
	"github.com/InkyQuill/open-edda/project"
)

type updatePlan struct {
	Before    checkout               `json:"before"`
	Remote    project.ProjectVersion `json:"remote"`
	Local     fileproject.Inventory  `json:"local"`
	Snapshot  []localNode            `json:"snapshot"`
	Conflicts []string               `json:"conflicts"`
	Choices   map[string]string      `json:"choices"`
	Phase     string                 `json:"phase"`
	Units     []updateUnit           `json:"units,omitempty"`
}
type updateUnit struct {
	Name  string `json:"name"`
	Old   bool   `json:"old"`
	New   bool   `json:"new"`
	Phase string `json:"phase"`
}

func planPath(root string, state checkout) string {
	return filepath.Join(root, ".edda", state.Update, "plan.json")
}
func loadUpdate(root string, state checkout) (updatePlan, error) {
	var p updatePlan
	if !strings.HasPrefix(state.Update, "update-") || filepath.Base(state.Update) != state.Update {
		return p, errors.New("invalid update reference")
	}
	dir := filepath.Join(root, ".edda", state.Update)
	info, err := os.Lstat(dir)
	if err != nil {
		return p, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return p, errors.New("update snapshot must be a real directory")
	}
	data, err := os.ReadFile(planPath(root, state))
	if err != nil {
		return p, err
	}
	if len(data) > 128<<20 {
		return p, errors.New("update plan exceeds limit")
	}
	if err = json.Unmarshal(data, &p); err != nil {
		return p, err
	}
	if p.Before.Update != "" || p.Before.Base.ProjectID != state.Base.ProjectID || p.Before.Server != state.Server || p.Remote.ProjectID != state.Base.ProjectID || p.Remote.ID == "" {
		return p, errors.New("invalid update plan identity")
	}
	for _, entries := range [][]project.TreeEntry{p.Before.Base.Entries, p.Remote.Entries, p.Local.Entries} {
		if err = project.ValidateTree(entries); err != nil {
			return p, err
		}
	}
	for _, unit := range p.Units {
		if unit.Name == "." || unit.Name == ".edda" || !validUpdateName(unit.Name) {
			return p, errors.New("invalid update journal path")
		}
	}
	return p, nil
}
func validUpdateName(name string) bool {
	return name != "" && name != ".." && !strings.ContainsAny(name, "/\\")
}
func saveUpdate(root string, state checkout, p updatePlan) error {
	return writePrivateJSON(planPath(root, state), p)
}
func entryMap(entries []project.TreeEntry) map[string]project.TreeEntry {
	m := map[string]project.TreeEntry{}
	for _, e := range entries {
		m[e.Path] = e
	}
	return m
}
func equalEntry(a, b project.TreeEntry) bool { a.ID = ""; b.ID = ""; return a == b }
func mergeSelection(p updatePlan) (map[string]string, []string, error) {
	base, local, remote := entryMap(p.Before.Base.Entries), entryMap(p.Local.Entries), entryMap(p.Remote.Entries)
	paths := map[string]bool{}
	for _, m := range []map[string]project.TreeEntry{base, local, remote} {
		for name := range m {
			paths[name] = true
		}
	}
	sources := map[string]string{}
	conflicts := []string{}
	// A concurrent edit plus a remote identity-preserving rename is a tree-level
	// conflict, not a guess that could leave two apparent copies of one document.
	remoteIDs := map[string]project.TreeEntry{}
	for _, e := range remote {
		remoteIDs[e.ID] = e
	}
	localIDs := map[string]project.TreeEntry{}
	for _, e := range local {
		localIDs[e.ID] = e
	}
	renamedConflict := false
	for name, b := range base {
		if l, ok := localIDs[b.ID]; ok && l.Path != name && !equalEntry(remote[name], b) {
			renamedConflict = true
		}
		if r, ok := remoteIDs[b.ID]; ok && r.Path != name && !equalEntry(local[name], b) {
			renamedConflict = true
		}
	}
	if renamedConflict && p.Choices["."] == "" {
		return nil, []string{"."}, nil
	}
	for name := range paths {
		side := p.Choices["."]
		// Choosing a conflicted directory also chooses its subtree.
		for parent := name; side == "" && parent != "."; parent = path.Dir(parent) {
			if choice := p.Choices[parent]; choice != "" {
				side = choice
				break
			}
		}
		if side == "" {
			switch {
			case equalEntry(local[name], base[name]):
				side = "remote"
			case equalEntry(remote[name], base[name]):
				side = "local"
			case equalEntry(local[name], remote[name]):
				side = "remote"
			default:
				conflicts = append(conflicts, name)
				continue
			}
		}
		switch side {
		case "local":
			if _, ok := local[name]; ok {
				sources[name] = side
			}
		case "remote":
			if _, ok := remote[name]; ok {
				sources[name] = side
			}
		default:
			return nil, nil, errors.New("invalid conflict choice")
		}
	}
	sort.Strings(conflicts)
	if len(conflicts) > 0 {
		return sources, conflicts, nil
	}
	// A retained changed child also retains the directories it needs.
	for name := range sources {
		for parent := path.Dir(name); parent != "."; parent = path.Dir(parent) {
			if _, ok := sources[parent]; ok {
				continue
			}
			if local[parent].Kind == "directory" {
				sources[parent] = "local"
			} else if remote[parent].Kind == "directory" {
				sources[parent] = "remote"
			}
		}
	}
	entries := []project.TreeEntry{}
	for name, side := range sources {
		e := local[name]
		if side == "remote" {
			e = remote[name]
		}
		entries = append(entries, e)
	}
	if err := project.ValidateTree(entries); err != nil {
		if p.Choices["."] == "" {
			return nil, []string{"."}, nil
		}
		return nil, nil, fmt.Errorf("chosen tree is invalid: %w", err)
	}
	return sources, nil, nil
}
