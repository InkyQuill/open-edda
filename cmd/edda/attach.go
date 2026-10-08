package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/InkyQuill/open-edda/fileproject"
	"github.com/InkyQuill/open-edda/project"
)

// attach writes only checkout metadata. Existing author files are never replaced.
// A nonempty server must match exactly, because otherwise there is no shared
// ancestor from which deletions and unsent local additions can be distinguished.
func runAttach(args []string, output io.Writer) error {
	root, rest := splitOptionalPath(args)
	flags := flag.NewFlagSet("attach", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var excludes importExclusions
	flags.Var(&excludes, "exclude", "exact relative path to keep local (repeatable)")
	server := flags.String("server", "", "server URL")
	id := flags.String("project", "", "project ID")
	if err := flags.Parse(rest); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("usage: edda attach FOLDER --project ID [--server URL]")
	}
	if err := askProject(id, *server, output); err != nil {
		return err
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("attach requires a real directory")
	}
	if _, err := os.Lstat(filepath.Join(root, ".edda")); !os.IsNotExist(err) {
		return errors.New(".edda already exists or is inaccessible; attach never replaces local metadata")
	}
	c, err := resolveConnection(*server)
	if err != nil {
		return err
	}
	client, err := newImportClient(c.Server, *id, c.Token)
	if err != nil {
		return err
	}
	ctx := context.Background()
	head, err := client.version(ctx, "versions/current")
	if err != nil {
		return err
	}
	if err := validateRemoteSelection(checkout{Excludes: excludes}, head.Entries); err != nil {
		return err
	}
	inventory, err := fileproject.ScanInventory(ctx, root, excludes, head.Entries)
	if err != nil {
		return err
	}
	reportSyncExclusions(inventory, output)
	if len(inventory.Problems) > 0 {
		reportInventoryProblems(inventory, output)
		return errors.New("exclude these paths in .eddaignore or with --exclude, then retry attach")
	}
	if len(head.Entries) > 0 && !sameFiles(inventory.Entries, head.Entries) {
		return errors.New("nonempty remote differs from this folder; get a separate copy and reconcile first, or attach to an empty project")
	}
	if err := installBinding(root, c.Server, head, excludes); err != nil {
		return err
	}
	fmt.Fprintln(output, "Folder attached. Use edda status, edda send or edda take anywhere inside this folder.")
	return nil
}

// Persist the acknowledged version, not a newer remote head: later local edits
// and concurrent server changes must still be reconciled against this base.
func installBinding(root, server string, base project.ProjectVersion, excludes []string) error {
	if err := validateRemoteSelection(checkout{Excludes: excludes}, base.Entries); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(filepath.Dir(root), ".edda-attach-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	if err := writePrivateJSON(filepath.Join(stage, "checkout.json"), checkout{Schema: 1, Server: server, Base: base, Excludes: excludes}); err != nil {
		return err
	}
	if err := renameNoReplace(stage, filepath.Join(root, ".edda")); err != nil {
		return err
	}
	return syncDirectory(root)
}

func runCreateProject(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("create", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	title := flags.String("title", "", "project title")
	machine := flags.Bool("json", false, "output JSON for scripts")
	server := flags.String("server", "", "server URL")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("usage: edda create --title TITLE [--server URL]")
	}
	if *machine && *title == "" {
		return errors.New("--title is required with --json")
	}
	if err := askValue(title, "Project title (--title)", "", output); err != nil {
		return err
	}
	created, err := createProject(*server, *title)
	if err != nil {
		return err
	}
	if *machine {
		return json.NewEncoder(output).Encode(created)
	}
	fmt.Fprintf(output, "Created %s\nID: %s\nAttach a local folder with edda attach FOLDER --project %s, then edda send FOLDER.\n", terminalText(created.Title), created.ID, created.ID)
	return nil
}

func createProject(server, title string) (project.StoryProject, error) {
	var created project.StoryProject
	c, err := resolveConnection(server)
	if err != nil {
		return created, err
	}
	client, err := newImportClient(c.Server, "_", c.Token)
	if err != nil {
		return created, err
	}
	client.root = c.Server + "/api/"
	body, err := json.Marshal(map[string]string{"title": title, "storageMode": "files"})
	if err != nil {
		return created, err
	}
	response, err := client.request(context.Background(), "POST", "projects", bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return created, err
	}
	defer response.Body.Close()
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&created); err != nil {
		return created, err
	}
	return created, nil
}
