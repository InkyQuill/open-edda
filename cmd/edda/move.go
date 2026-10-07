package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/InkyQuill/open-edda/fileproject"
	"github.com/InkyQuill/open-edda/project"
)

type pendingMove struct {
	Untracked []string            `json:"untracked,omitempty"`
	Excludes  []string            `json:"excludes,omitempty"`
	From      string              `json:"from"`
	To        string              `json:"to"`
	Entries   []project.TreeEntry `json:"entries"`
}

func localIdentities(state checkout, entries []project.TreeEntry) []project.TreeEntry {
	known := state.Base.Entries
	if state.Identity != nil {
		known = state.Identity
	}
	byPath := entryMap(known)
	result := append([]project.TreeEntry{}, entries...)
	for i := range result {
		old, ok := byPath[result[i].Path]
		if ok && old.Kind == result[i].Kind {
			result[i].ID = old.ID
		} else {
			result[i].ID = randomSyncID()
		}
	}
	return result
}

func runMove(args []string, output io.Writer) error {
	root, rest := splitOptionalPath(args)
	flags := flag.NewFlagSet("move", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	from := flags.String("from", "", "relative source path")
	to := flags.String("to", "", "relative destination path")
	if err := flags.Parse(rest); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("move accepts one checkout")
	}
	if _, err := readCheckout(root); err != nil {
		return err
	}
	unlock, err := lockCheckout(root)
	if err != nil {
		return err
	}
	defer unlock()
	state, err := readCheckout(root)
	if err != nil {
		return err
	}
	if state.Move != nil {
		return finishMove(root, state, output)
	}
	if state.Pending != nil || state.Update != "" {
		return errors.New("finish pending send/update before moving files")
	}
	if err := askValue(from, "Source relative path (--from)", "", output); err != nil {
		return err
	}
	if err := askValue(to, "Destination relative path (--to)", "", output); err != nil {
		return err
	}
	if *from == *to || strings.HasPrefix(*to, *from+"/") {
		return errors.New("usage: edda move CHECKOUT --from PATH --to PATH (existing destination parents required)")
	}
	for _, name := range state.Untracked {
		// Exclusions below the source travel with it; all other exclusions must
		// remain disjoint from the destination, including its parents/children.
		if name == *from || strings.HasPrefix(name, *from+"/") {
			continue
		}
		if name == *to || strings.HasPrefix(*to, name+"/") || strings.HasPrefix(name, *to+"/") {
			return fmt.Errorf("destination overlaps untracked path %q; use edda rm --undo first", name)
		}
	}
	inventory, err := fileproject.ScanInventory(context.Background(), root, state.localExclusions(), state.Base.Entries, state.Identity)
	if err != nil {
		return err
	}
	if len(inventory.Problems) > 0 {
		return errors.New("unsupported local entries; inspect import --dry-run")
	}
	entries := localIdentities(state, inventory.Entries)
	found := false
	for i := range entries {
		if entries[i].Path == *from || strings.HasPrefix(entries[i].Path, *from+"/") {
			entries[i].Path = *to + strings.TrimPrefix(entries[i].Path, *from)
			found = true
		}
	}
	if !found {
		return errors.New("source is not a synchronized file or directory")
	}
	if err := validateCheckoutTree(entries); err != nil {
		return err
	}
	if _, err := os.Lstat(filepath.Join(root, filepath.FromSlash(*to))); !os.IsNotExist(err) {
		return errors.New("destination exists or is inaccessible")
	}
	excludes := append([]string{}, state.Excludes...)
	for i, name := range excludes {
		if name == *from || strings.HasPrefix(name, *from+"/") {
			excludes[i] = *to + strings.TrimPrefix(name, *from)
		}
	}
	untracked := append([]string{}, state.Untracked...)
	for i, name := range untracked {
		if name == *from || strings.HasPrefix(name, *from+"/") {
			untracked[i] = *to + strings.TrimPrefix(name, *from)
		}
	}
	state.Move = &pendingMove{Untracked: untracked, From: *from, To: *to, Entries: entries, Excludes: excludes}
	if err := writePrivateJSON(checkoutPath(root), state); err != nil {
		return err
	}
	return finishMove(root, state, output)
}
func finishMove(root string, state checkout, output io.Writer) error {
	move := state.Move
	source := filepath.Join(root, filepath.FromSlash(move.From))
	target := filepath.Join(root, filepath.FromSlash(move.To))
	_, sourceErr := os.Lstat(source)
	_, targetErr := os.Lstat(target)
	switch {
	case sourceErr == nil && os.IsNotExist(targetErr):
		// OpenRoot rejects ancestor symlinks escaping the checkout; Lstat alone would
		// not. Inventory preparation already rejected synchronized symlink entries.
		r, err := os.OpenRoot(root)
		if err != nil {
			return err
		}
		defer r.Close()
		if _, err := r.Stat(filepath.ToSlash(filepath.Dir(move.To))); err != nil {
			return err
		}
		if err := renameNoReplace(source, target); err != nil {
			return err
		}
	case os.IsNotExist(sourceErr) && targetErr == nil:
		// The filesystem rename completed before checkout metadata was committed.
	default:
		return errors.New("move recovery found both/neither paths; inspect files before retrying")
	}
	if err := syncDirectory(filepath.Dir(source)); err != nil {
		return err
	}
	if err := syncDirectory(filepath.Dir(target)); err != nil {
		return err
	}
	state.Excludes = move.Excludes
	state.Untracked = move.Untracked
	state.Identity = move.Entries
	state.Move = nil
	if err := writePrivateJSON(checkoutPath(root), state); err != nil {
		return err
	}
	fmt.Fprintln(output, "Moved locally with file identities preserved. Run send to publish.")
	return nil
}
