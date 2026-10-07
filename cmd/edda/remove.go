package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func runRemove(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("rm", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	undo := flags.Bool("undo", false, "resume tracking locally kept paths")
	if err := flags.Parse(args); err != nil {
		return err
	}
	paths := flags.Args()
	if len(paths) == 0 {
		var name string
		if err := askValue(&name, "Path to stop tracking", "", output); err != nil {
			return err
		}
		paths = []string{name}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err := findProjectRoot(cwd)
	if err != nil {
		return err
	}
	if root == "" {
		return errors.New("run edda rm inside a connected project")
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
	if state.Pending != nil || state.Move != nil || state.Update != "" {
		return errors.New("finish pending send, move or update before changing tracked paths")
	}
	known := state.Base.Entries
	if state.Identity != nil {
		known = state.Identity
	}
	for _, arg := range paths {
		absolute, err := filepath.Abs(arg)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, absolute)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(relative)
		if !filepath.IsLocal(relative) || name == "." || name == ".edda" || strings.HasPrefix(name, ".edda/") {
			return fmt.Errorf("path %q must be inside the project and outside .edda", arg)
		}
		if *undo {
			if !slices.Contains(state.Untracked, name) {
				return fmt.Errorf("%q was not removed with edda rm; use the originally removed path", name)
			}
			state.Untracked = slices.DeleteFunc(state.Untracked, func(p string) bool { return p == name })
			continue
		}
		if slices.Contains(state.Untracked, name) {
			continue
		}
		found := false
		for _, entry := range known {
			if entry.Path == name || strings.HasPrefix(entry.Path, name+"/") {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%q is not tracked; use .eddaignore for new local-only files", name)
		}
		state.Untracked = append(state.Untracked, name)
	}
	if err := writePrivateJSON(checkoutPath(root), state); err != nil {
		return err
	}
	if *undo {
		fmt.Fprintln(output, "Tracking exclusion removed. Local files are unchanged; other ignore rules still apply. Check edda status, then send.")
	} else {
		fmt.Fprintln(output, "Stopped tracking; local files are unchanged. Run edda send to remove them from the server tree. Use edda rm --undo PATH to undo.")
	}
	return nil
}
