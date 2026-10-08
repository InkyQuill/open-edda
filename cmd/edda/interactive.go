package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/InkyQuill/open-edda/fileproject"
	"github.com/InkyQuill/open-edda/project"
	"golang.org/x/term"
)

func interactiveInput() bool { return term.IsTerminal(int(os.Stdin.Fd())) }

func terminalText(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, s)
}

// Explicit values always win. Only missing values are requested from a terminal.
func askValue(value *string, label, fallback string, output io.Writer) error {
	if *value != "" {
		return nil
	}
	if !interactiveInput() {
		return fmt.Errorf("%s is required; pass it as an argument or run this command in a terminal", strings.ToLower(label))
	}
	prompt := label
	if fallback != "" {
		prompt += " [" + terminalText(fallback) + "]"
	}
	for {
		answer, err := promptLine(os.Stdin, output, prompt+": ")
		if err != nil {
			return err
		}
		if answer == "" {
			answer = fallback
		}
		if answer != "" {
			*value = answer
			return nil
		}
		fmt.Fprintln(output, "Please enter a value, or press Ctrl+C to cancel.")
	}
}

func askChoice(value *string, label string, choices []string, output io.Writer) error {
	if *value != "" {
		return nil
	}
	for {
		if err := askValue(value, label+" ("+strings.Join(choices, " / ")+")", "", output); err != nil {
			return err
		}
		for _, choice := range choices {
			if *value == choice {
				return nil
			}
		}
		fmt.Fprintln(output, "Choose one of the listed values.")
		*value = ""
	}
}

func projectList(server string) ([]project.StoryProject, error) {
	c, err := resolveConnection(server)
	if err != nil {
		return nil, err
	}
	client, err := newImportClient(c.Server, "_", c.Token)
	if err != nil {
		return nil, err
	}
	client.root = c.Server + "/api/"
	response, err := client.request(context.Background(), "GET", "projects", nil, 0)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	var projects []project.StoryProject
	err = json.NewDecoder(io.LimitReader(response.Body, 8<<20)).Decode(&projects)
	return projects, err
}

func askProject(id *string, server string, output io.Writer) error {
	if *id != "" {
		return nil
	}
	if !interactiveInput() {
		return errors.New("--project ID is required; use edda projects to list IDs, or run this command in a terminal")
	}
	projects, err := projectList(server)
	if err != nil {
		return err
	}
	var choices []project.StoryProject
	for _, p := range projects {
		if p.StorageMode == "files" {
			choices = append(choices, p)
		}
	}
	if len(choices) == 0 {
		return errors.New("no file projects available; create one with edda create, or upload a local folder with edda send FOLDER")
	}
	for i, p := range choices {
		fmt.Fprintf(output, "  %d. %s (%s)\n", i+1, terminalText(p.Title), terminalText(p.ID))
	}
	for {
		var answer string
		if err := askValue(&answer, "Project number or ID (--project)", "", output); err != nil {
			return err
		}
		n, err := strconv.Atoi(answer)
		if err == nil && n > 0 && n <= len(choices) {
			*id = choices[n-1].ID
			return nil
		}
		for _, p := range choices {
			if answer == p.ID {
				*id = answer
				return nil
			}
		}
		fmt.Fprintln(output, "Choose a project from the list.")
	}
}

func prepareFirstSend(root, server, title, id string, excludes []string, output io.Writer, quiet bool) error {
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("send requires a real directory")
	}
	if _, err := os.Lstat(filepath.Join(root, ".edda")); !os.IsNotExist(err) {
		return errors.New("folder already has .edda metadata; first send will not replace it. Use a separate folder for a network checkout")
	}
	ctx, _, finish := observeSync(context.Background(), output, quiet)
	inventory, err := fileproject.ScanInventory(ctx, root, excludes)
	finish()
	if err != nil {
		return err
	}
	printInventory(inventory, output)
	if len(inventory.Problems) > 0 {
		return errors.New("resolve unsupported entries or add --exclude PATH before sending")
	}
	if _, err := resolveConnection(server); err != nil {
		return err
	}
	if id == "" && title == "" {
		if !interactiveInput() {
			return errors.New("folder is not attached; use edda send FOLDER --title TITLE to create a project, or --project ID to attach to one")
		}
		var choice string
		if err := askChoice(&choice, "First send: create a new project or select an existing project", []string{"new", "existing"}, output); err != nil {
			return err
		}
		if choice == "existing" {
			if err := askProject(&id, server, output); err != nil {
				return err
			}
		} else {
			if err := askValue(&title, "New project title (--title)", filepath.Base(root), output); err != nil {
				return err
			}
		}
	}
	if id == "" {
		created, err := createProject(server, title)
		if err != nil {
			return err
		}
		id = created.ID
		fmt.Fprintf(output, "Created %s (%s).\n", terminalText(created.Title), id)
	}
	args := []string{root, "--project", id}
	if server != "" {
		args = append(args, "--server", server)
	}
	for _, exclude := range excludes {
		args = append(args, "--exclude", exclude)
	}
	ctx, output, finish = observeSync(context.Background(), output, quiet)
	defer finish()
	syncPhase(ctx, "Attaching project")
	if err := runAttachContext(ctx, args, output); err != nil {
		return fmt.Errorf("project %s exists, but attachment failed (retry with --project %s): %w", id, id, err)
	}
	return nil
}

func printInventory(inventory fileproject.Inventory, output io.Writer) {
	files, dirs := 0, 0
	for _, entry := range inventory.Entries {
		if entry.Kind == "file" {
			files++
		} else {
			dirs++
		}
	}
	fmt.Fprintf(output, "Folder: %s\n%d files, %d directories, %d bytes\n", terminalText(inventory.Root), files, dirs, inventory.Bytes)
	reportSyncExclusions(inventory, output)
	reportInventoryProblems(inventory, output)
}

func reportInventoryProblems(inventory fileproject.Inventory, output io.Writer) {
	for _, problem := range inventory.Problems {
		fmt.Fprintf(output, "Cannot include %q: %s\n", problem.Path, terminalText(problem.Reason))
	}
}
