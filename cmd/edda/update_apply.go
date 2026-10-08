package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/InkyQuill/open-edda/fileproject"
	"github.com/InkyQuill/open-edda/project"
	"golang.org/x/sys/unix"
)

func buildUpdateResult(ctx context.Context, root string, state checkout, p updatePlan, sources map[string]string) error {
	dir := filepath.Join(root, ".edda", state.Update)
	result := filepath.Join(dir, "result")
	localNodes, err := localSnapshot(ctx, filepath.Join(dir, "local"), "")
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(localNodes, p.Snapshot) {
		return errors.New("local recovery snapshot changed")
	}
	remoteInventory, err := fileproject.ScanInventory(ctx, filepath.Join(dir, "remote"), nil, p.Remote.Entries)
	if err != nil {
		return err
	}
	if len(remoteInventory.Problems) > 0 || !sameFiles(remoteInventory.Entries, p.Remote.Entries) {
		return errors.New("remote recovery snapshot changed")
	}
	if err := os.RemoveAll(result); err != nil {
		return err
	}
	if err := os.Mkdir(result, 0700); err != nil {
		return err
	}
	if _, err := localSnapshot(ctx, filepath.Join(dir, "local"), result); err != nil {
		return err
	}
	local := entryMap(p.Local.Entries)
	remote := entryMap(p.Remote.Entries)
	names := []string{}
	for name := range local {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool { return len(names[i]) > len(names[j]) })
	for _, name := range names {
		side, keep := sources[name]
		selected := local[name]
		if side == "remote" {
			selected = remote[name]
		}
		if keep && selected.Kind == local[name].Kind {
			continue
		}
		// Remove only portable entries. A directory containing excluded local data
		// fails rather than silently deleting those files.
		if err := os.Remove(filepath.Join(result, filepath.FromSlash(name))); err != nil {
			return fmt.Errorf("cannot replace %q (local-only contents are preserved): %w", name, err)
		}
	}
	names = names[:0]
	for name := range sources {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool { return len(names[i]) < len(names[j]) })
	for _, name := range names {
		if sources[name] != "remote" {
			continue
		}
		entry := remote[name]
		// The result already contains the verified local snapshot. Rewriting
		// identical bytes adds disk I/O and fsync without changing the result.
		if current, ok := local[name]; ok && current.Kind == entry.Kind && current.SHA256 == entry.SHA256 && current.Bytes == entry.Bytes {
			continue
		}
		target := filepath.Join(result, filepath.FromSlash(name))
		if entry.Kind == "directory" {
			if err := os.MkdirAll(target, 0700); err != nil {
				return err
			}
			continue
		}
		source, err := os.Open(filepath.Join(dir, "remote", filepath.FromSlash(name)))
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			source.Close()
			return err
		}
		_, copyErr := io.Copy(out, source)
		source.Close()
		if copyErr == nil {
			copyErr = out.Sync()
		}
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	_, err = localSnapshot(ctx, result, "")
	if err != nil {
		return err
	}
	return syncTreeDirectories(result)
}
func syncTreeDirectories(root string) error {
	dirs := []string{}
	err := filepath.WalkDir(root, func(name string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			dirs = append(dirs, name)
		}
		return nil
	})
	if err != nil {
		return err
	}
	for i := len(dirs) - 1; i >= 0; i-- {
		if err := syncDirectory(dirs[i]); err != nil {
			return err
		}
	}
	return nil
}
func nodesUnder(nodes []localNode, name string) []localNode {
	out := []localNode{}
	for _, n := range nodes {
		if n.Path == name || strings.HasPrefix(n.Path, name+"/") {
			out = append(out, n)
		}
	}
	return out
}
func exists(name string) bool { _, err := os.Lstat(name); return err == nil }

func applyUpdate(ctx context.Context, root string, state checkout, p updatePlan) error {
	return applyUpdateObserved(ctx, root, state, p, func(string) {})
}

// observe marks durable boundaries; subprocess tests terminate at these points.
func applyUpdateObserved(ctx context.Context, root string, state checkout, p updatePlan, observe func(string)) error {
	sources, conflicts, err := mergeSelection(p)
	if err != nil {
		return err
	}
	if len(conflicts) > 0 {
		return errors.New("resolve all conflicts before applying")
	}
	if err = buildUpdateResult(ctx, root, state, p, sources); err != nil {
		return err
	}
	dir := filepath.Join(root, ".edda", state.Update)
	after, err := localSnapshot(ctx, filepath.Join(dir, "result"), "")
	if err != nil {
		return err
	}
	before, err := localSnapshot(ctx, root, "")
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(before, p.Snapshot) {
		return errors.New("local files changed since update preparation; run take --restart to prepare again (snapshots are retained)")
	}
	names := map[string]bool{}
	for _, list := range [][]localNode{before, after} {
		for _, n := range list {
			names[strings.Split(n.Path, "/")[0]] = true
		}
	}
	ordered := []string{}
	for name := range names {
		ordered = append(ordered, name)
	}
	sort.Strings(ordered)
	p.Units = nil
	for _, name := range ordered {
		oldNodes, newNodes := nodesUnder(before, name), nodesUnder(after, name)
		if !reflect.DeepEqual(oldNodes, newNodes) {
			p.Units = append(p.Units, updateUnit{Name: name, Old: len(oldNodes) > 0, New: len(newNodes) > 0, Phase: "pending"})
		}
	}
	for _, name := range []string{"backup", "displaced"} {
		if err = os.MkdirAll(filepath.Join(dir, name), 0700); err != nil {
			return err
		}
	}
	p.Phase = "applying"
	if err = saveUpdate(root, state, p); err != nil {
		return err
	}
	for i := range p.Units {
		if err = ctx.Err(); err != nil {
			return err
		}
		u := &p.Units[i]
		u.Phase = "moving"
		if err = saveUpdate(root, state, p); err != nil {
			return err
		}
		if u.Old {
			if err = renameNoReplace(filepath.Join(root, u.Name), filepath.Join(dir, "backup", u.Name)); err != nil {
				return err
			}
			if err = syncDirectory(root); err != nil {
				return err
			}
			if err = syncDirectory(filepath.Join(dir, "backup")); err != nil {
				return err
			}
		}
		observe("backup")
		if u.Old {
			moved, checkErr := localSnapshot(ctx, filepath.Join(dir, "backup"), "", u.Name)
			if checkErr != nil {
				return checkErr
			}
			if !reflect.DeepEqual(moved, nodesUnder(p.Snapshot, u.Name)) {
				return errors.New("local edits detected during application; run take to recover")
			}
		}
		u.Phase = "backed-up"
		if err = saveUpdate(root, state, p); err != nil {
			return err
		}
		if u.New {
			if err = renameNoReplace(filepath.Join(dir, "result", u.Name), filepath.Join(root, u.Name)); err != nil {
				return err
			}
			if err = syncDirectory(root); err != nil {
				return err
			}
			if err = syncDirectory(filepath.Join(dir, "result")); err != nil {
				return err
			}
		}
		observe("install")
		u.Phase = "installed"
		if err = saveUpdate(root, state, p); err != nil {
			return err
		}
	}
	observe("before-commit")
	// Commit point: all installed directories/files were synced before the base.
	state.Identity = []project.TreeEntry{}
	local, remote := entryMap(p.Local.Entries), entryMap(p.Remote.Entries)
	for name, side := range sources {
		entry := local[name]
		if side == "remote" {
			entry = remote[name]
		}
		state.Identity = append(state.Identity, entry)
	}
	sort.Slice(state.Identity, func(i, j int) bool { return state.Identity[i].Path < state.Identity[j].Path })
	state.Base = p.Remote
	state.Pending = nil
	if err = writePrivateJSON(checkoutPath(root), state); err != nil {
		return err
	}
	observe("after-commit")
	p.Phase = "committed"
	if err = saveUpdate(root, state, p); err != nil {
		return err
	}
	state.Update = ""
	return writePrivateJSON(checkoutPath(root), state)
}

// Interrupted application rolls back all touched top-level entries. Any files
// edited after the interruption are moved into displaced/, never overwritten.
func recoverUpdate(root string, state checkout, p updatePlan) error {
	if state.Base.ID == p.Remote.ID {
		p.Phase = "committed"
		if err := saveUpdate(root, state, p); err != nil {
			return err
		}
		state.Update = ""
		return writePrivateJSON(checkoutPath(root), state)
	}
	dir := filepath.Join(root, ".edda", state.Update)
	if err := os.MkdirAll(filepath.Join(dir, "displaced"), 0700); err != nil {
		return err
	}
	for i := len(p.Units) - 1; i >= 0; i-- {
		u := &p.Units[i]
		if u.Phase == "pending" || u.Phase == "restored" {
			continue
		}
		original := filepath.Join(dir, "backup", u.Name)
		current := filepath.Join(root, u.Name)
		if u.Old && !exists(original) {
			if (u.Phase == "moving" || u.Phase == "restoring") && exists(current) {
				u.Phase = "restored"
				if err := saveUpdate(root, state, p); err != nil {
					return err
				}
				continue
			}
			return fmt.Errorf("missing update backup for %q; retained local snapshot requires inspection", u.Name)
		}
		u.Phase = "restoring"
		if err := saveUpdate(root, state, p); err != nil {
			return err
		}
		if exists(current) {
			if err := renameNoReplace(current, filepath.Join(dir, "displaced", u.Name+"-"+randomSyncID())); err != nil {
				return err
			}
			if err := syncDirectory(root); err != nil {
				return err
			}
			if err := syncDirectory(filepath.Join(dir, "displaced")); err != nil {
				return err
			}
		}
		if u.Old {
			if err := renameNoReplace(original, current); err != nil {
				return err
			}
			if err := syncDirectory(root); err != nil {
				return err
			}
			if err := syncDirectory(filepath.Join(dir, "backup")); err != nil {
				return err
			}
		}
		u.Phase = "restored"
		if err := saveUpdate(root, state, p); err != nil {
			return err
		}
	}
	p.Phase = "rolled-back"
	if err := saveUpdate(root, state, p); err != nil {
		return err
	}
	return writePrivateJSON(checkoutPath(root), p.Before)
}

func conflictPathExists(p updatePlan, name string) bool {
	for _, conflict := range p.Conflicts {
		if name == conflict {
			return true
		}
	}
	return false
}
func safeConflictPath(name string) bool {
	return name == "." || (path.IsAbs(name) == false && path.Clean(name) == name && name != ".." && !strings.HasPrefix(name, "../"))
}

func renameNoReplace(from, to string) error {
	return unix.Renameat2(unix.AT_FDCWD, from, unix.AT_FDCWD, to, unix.RENAME_NOREPLACE)
}
