package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/InkyQuill/open-edda/fileproject"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stdout)
		return nil
	}

	if args[0] == "help" && len(args) > 1 {
		return printCommandHelp(args[1], stdout)
	}
	if _, ok := commandHelp[args[0]]; ok && wantsCommandHelp(args[1:]) {
		return printCommandHelp(args[0], stdout)
	}
	prepared, err := prepareInteractivePath(args[0], args[1:], stderr)
	if err != nil {
		return err
	}
	args = append([]string{args[0]}, prepared...)

	switch args[0] {
	case "init", "ids", "save", "checkpoint", "files", "diff":
		root, _ := splitOptionalPath(args[1:])
		if _, err := os.Lstat(checkoutPath(root)); err == nil {
			return errors.New("prototype snapshot command is not available in a network checkout; use status, send, take or history")
		}
	}
	switch args[0] {
	case "import":
		return runImport(args[1:], stdout, stderr)
	case "login":
		return runLogin(args[1:], os.Stdin, stdout)
	case "logout":
		if len(args) != 1 {
			return errors.New("logout takes no arguments")
		}
		return runLogout(stdout)
	case "projects":
		return runProjects(args[1:], stdout)
	case "backup", "verify-backup", "restore-backup":
		return runBackup(args[0], args[1:], stdout)
	case "create":
		return runCreateProject(args[1:], stdout)
	case "attach":
		return runAttach(args[1:], stdout)
	case "move":
		return runMove(args[1:], stdout)
	case "get":
		return runNetworkGet(args[1:], stdout)
	case "status":
		return runStatus(args[1:], stdout)
	case "ids":
		return runIDs(args[1:], stdout)
	case "init":
		return runInit(args[1:], stdout)
	case "save":
		return runSave(args[1:], stdout)
	case "send":
		return runNetworkSend(args[1:], stdout)
	case "take":
		return runNetworkTake(args[1:], stdout)
	case "checkpoint":
		return runCheckpoint(args[1:], stdout)
	case "history":
		root, _ := splitOptionalPath(args[1:])
		if _, err := os.Lstat(checkoutPath(root)); err == nil {
			return runNetworkHistory(args[1:], stdout)
		}
		return runHistory(args[1:], stdout)
	case "files":
		return runFiles(args[1:], stdout)
	case "diff":
		return runDiff(args[1:], stdout)
	case "restore":
		root, _ := splitOptionalPath(args[1:])
		if _, err := os.Lstat(checkoutPath(root)); err == nil {
			return runNetworkRestore(args[1:], stdout)
		}
		return runRestore(args[1:], stdout)
	case "conflicts":
		root, _ := splitOptionalPath(args[1:])
		if _, err := os.Lstat(checkoutPath(root)); err == nil {
			root, err := networkRoot("conflicts", args[1:])
			if err != nil {
				return err
			}
			return runNetworkConflicts(root, stdout)
		}
		return runConflicts(args[1:], stdout)
	case "resolve":
		root, _ := splitOptionalPath(args[1:])
		if _, err := os.Lstat(checkoutPath(root)); err == nil {
			return runNetworkResolve(args[1:], stdout)
		}
		return runResolve(args[1:], stdout)
	case "help", "-h", "--help":
		printUsage(stdout)
		return nil
	default:
		printUsage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runStatus(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root, flagArgs := splitOptionalPath(args)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}

	if _, err := os.Lstat(checkoutPath(root)); err == nil {
		return runNetworkStatus(root, stdout)
	}
	layout, err := fileproject.Scan(root)
	if err != nil {
		return err
	}

	if layout.Metadata != nil {
		fmt.Fprintf(stdout, "Project: %s (%s)\n", layout.Metadata.Title, layout.Metadata.ID)
	} else {
		fmt.Fprintln(stdout, "Project: uninitialized Edda folder")
	}
	fmt.Fprintf(stdout, "Root: %s\n", layout.Root)
	if _, err := os.Stat(filepath.Join(layout.Root, ".edda", "ids.json")); err == nil {
		fmt.Fprintln(stdout, "Stable IDs: present")
	} else if errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(stdout, "Stable IDs: missing")
	} else {
		return fmt.Errorf("stat stable IDs: %w", err)
	}

	counts := fileproject.CountByKind(layout.Files)
	kinds := make([]fileproject.LayoutKind, 0, len(counts))
	for kind := range counts {
		kinds = append(kinds, kind)
	}
	sort.Slice(kinds, func(i, j int) bool { return kinds[i] < kinds[j] })
	fmt.Fprintln(stdout, "Files:")
	for _, kind := range kinds {
		fmt.Fprintf(stdout, "  %s: %d\n", kind, counts[kind])
	}
	if len(kinds) == 0 {
		fmt.Fprintln(stdout, "  none")
	}

	if len(layout.Warnings) > 0 {
		fmt.Fprintln(stdout, "Warnings:")
		for _, warning := range layout.Warnings {
			if warning.Path != "" {
				fmt.Fprintf(stdout, "  %s: %s (%s)\n", warning.Code, warning.Message, warning.Path)
			} else {
				fmt.Fprintf(stdout, "  %s: %s\n", warning.Code, warning.Message)
			}
		}
	}

	return nil
}

func runIDs(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		var action string
		if err := askChoice(&action, "IDs action", []string{"sync"}, stdout); err != nil {
			return err
		}
		args = []string{action}
	}
	switch args[0] {
	case "sync":
		prepared, err := prepareInteractivePath("files", args[1:], stdout)
		if err != nil {
			return err
		}
		return runIDSync(prepared, stdout)
	default:
		return fmt.Errorf("unknown ids subcommand %q", args[0])
	}
}

func runIDSync(args []string, stdout io.Writer) error {
	root, flagArgs := splitOptionalPath(args)
	flags := flag.NewFlagSet("ids sync", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}
	_, files, err := fileproject.SyncStableIDs(root)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Updated .edda/ids.json for %d files.\n", len(files))
	return nil
}

func runInit(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	title := flags.String("title", "", "project title")
	id := flags.String("id", "", "project id")
	serverURL := flags.String("server-url", "", "server URL")
	root, flagArgs := splitExistingOptionalPath(args)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}

	if *title == "" && interactiveInput() {
		absolute, err := filepath.Abs(root)
		if err != nil {
			return err
		}
		if err := askValue(title, "Project title (--title)", filepath.Base(absolute), stdout); err != nil {
			return err
		}
	}
	metadata, err := fileproject.InitMetadata(root, fileproject.InitMetadataInput{
		ID:        *id,
		Title:     *title,
		ServerURL: *serverURL,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Initialized Edda project: %s (%s)\n", metadata.Title, metadata.ID)
	return nil
}

func runSave(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("save", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	fileID := flags.String("id", "", "stable file id")
	fromDraft := flags.Bool("from-draft", false, "promote .edda/drafts/<id>.md")
	bodyFile := flags.String("body-file", "", "markdown file to save")
	expectedSHA256 := flags.String("expected-sha256", "", "expected current saved file hash")
	root, flagArgs := splitExistingOptionalPath(args)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		if *fileID != "" || *fromDraft || *bodyFile != "" {
			root = flags.Arg(0)
		}
	}
	if *fileID == "" && !*fromDraft && *bodyFile == "" {
		return runSaveCheckpoint(root, flags.Args(), stdout)
	}
	if err := askValue(fileID, "File ID (--id)", "", stdout); err != nil {
		return err
	}
	if !*fromDraft && *bodyFile == "" {
		var choice string
		if err := askChoice(&choice, "Save source", []string{"draft", "file"}, stdout); err != nil {
			return err
		}
		if choice == "draft" {
			*fromDraft = true
		} else if err := askValue(bodyFile, "Markdown source (--body-file)", "", stdout); err != nil {
			return err
		}
	}
	if *fromDraft == (*bodyFile != "") {
		return fmt.Errorf("save requires exactly one of --from-draft or --body-file")
	}

	var (
		saved fileproject.SavedFile
		err   error
	)
	if *fromDraft {
		saved, err = fileproject.PromoteDraft(root, fileproject.SaveDraftInput{
			FileID:         *fileID,
			ExpectedSHA256: *expectedSHA256,
		})
	} else {
		body, readErr := os.ReadFile(*bodyFile)
		if readErr != nil {
			return fmt.Errorf("read body file: %w", readErr)
		}
		saved, err = fileproject.SaveCanonicalFile(root, fileproject.SaveCanonicalInput{
			FileID:         *fileID,
			BodyMarkdown:   string(body),
			ExpectedSHA256: *expectedSHA256,
		})
	}
	if err != nil {
		if errors.Is(err, fileproject.ErrFileConflict) {
			return fmt.Errorf("saved file changed since draft base: %w", err)
		}
		return err
	}
	fmt.Fprintf(stdout, "Saved %s (%s, %d bytes)\n", saved.Path, saved.SHA256, saved.Size)
	return nil
}

func runSaveCheckpoint(root string, args []string, stdout io.Writer) error {
	message := strings.TrimSpace(strings.Join(args, " "))
	if err := askValue(&message, "Checkpoint note", "", stdout); err != nil {
		return err
	}
	checkpoint, err := fileproject.CreateCheckpoint(root, fileproject.CreateCheckpointInput{Message: message})
	if err != nil {
		return err
	}
	if _, err := fileproject.RecordPendingUpload(root, checkpoint.ID); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Saved prototype checkpoint %s (%d files); this queue is separate from network checkout sends\n", checkpoint.ID, len(checkpoint.Files))
	return nil
}

func runCheckpoint(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("checkpoint", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	message := flags.String("message", "", "checkpoint message")
	root, flagArgs := splitOptionalPath(args)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}
	if *message == "" && interactiveInput() {
		if err := askValue(message, "Checkpoint note (--message)", "", stdout); err != nil {
			return err
		}
	}
	checkpoint, err := fileproject.CreateCheckpoint(root, fileproject.CreateCheckpointInput{Message: *message})
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Checkpoint %s (%d files)\n", checkpoint.ID, len(checkpoint.Files))
	return nil
}

func runConflicts(args []string, stdout io.Writer) error {
	root, flagArgs := splitOptionalPath(args)
	flags := flag.NewFlagSet("conflicts", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}
	conflicts, err := fileproject.ListConflicts(root)
	if err != nil {
		return err
	}
	if len(conflicts) == 0 {
		fmt.Fprintln(stdout, "No conflicts.")
		return nil
	}
	for _, conflict := range conflicts {
		fmt.Fprintf(stdout, "%s %s\n", conflict.FileID, conflict.Path)
	}
	return nil
}

func runResolve(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("resolve", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	fileID := flags.String("id", "", "stable file id")
	use := flags.String("use", "", "conflict version to use: local or server")
	bodyFile := flags.String("body-file", "", "resolved markdown file")
	root, flagArgs := splitOptionalPath(args)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}
	if *fileID == "" && interactiveInput() {
		if err := runConflicts([]string{root}, stdout); err != nil {
			return err
		}
	}
	if err := askValue(fileID, "Conflict file ID (--id)", "", stdout); err != nil {
		return err
	}
	if *bodyFile == "" && *use == "" && interactiveInput() {
		var choice string
		if err := askChoice(&choice, "Use version", []string{"local", "server", "file"}, stdout); err != nil {
			return err
		}
		if choice == "file" {
			if err := askValue(bodyFile, "Resolved Markdown file (--body-file)", "", stdout); err != nil {
				return err
			}
		} else {
			*use = choice
		}
	}
	if (*bodyFile != "") == (*use != "") {
		return fmt.Errorf("resolve requires exactly one of --use or --body-file")
	}
	var body string
	if *bodyFile != "" {
		data, err := os.ReadFile(*bodyFile)
		if err != nil {
			return fmt.Errorf("read body file: %w", err)
		}
		body = string(data)
	}
	record, saved, err := fileproject.ResolveConflict(root, fileproject.ResolveConflictInput{
		FileID:       *fileID,
		Use:          fileproject.ConflictVersion(*use),
		BodyMarkdown: body,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Resolved %s to %s (%s)\n", record.FileID, saved.Path, saved.SHA256)
	return nil
}

func runHistory(args []string, stdout io.Writer) error {
	root, flagArgs := splitOptionalPath(args)
	flags := flag.NewFlagSet("history", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	machine := flags.Bool("json", false, "output JSON for scripts")
	fileID := flags.String("id", "", "stable file id")
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}
	if *fileID != "" {
		history, err := fileproject.ListFileCheckpointHistory(root, *fileID)
		if err != nil {
			return err
		}
		if *machine {
			return json.NewEncoder(stdout).Encode(history)
		}
		if len(history) == 0 {
			fmt.Fprintln(stdout, "No file history.")
			return nil
		}
		for _, entry := range history {
			if entry.Message == "" {
				fmt.Fprintf(stdout, "%s %s %s %s %s\n", entry.CheckpointID, entry.FileID, entry.CreatedAt, entry.SHA256, entry.Path)
			} else {
				fmt.Fprintf(stdout, "%s %s %s %s %s %q\n", entry.CheckpointID, entry.FileID, entry.CreatedAt, entry.SHA256, entry.Path, entry.Message)
			}
		}
		return nil
	}
	checkpoints, err := fileproject.ListCheckpointSummaries(root)
	if err != nil {
		return err
	}
	if *machine {
		return json.NewEncoder(stdout).Encode(checkpoints)
	}
	if len(checkpoints) == 0 {
		fmt.Fprintln(stdout, "No checkpoints.")
		return nil
	}
	for _, checkpoint := range checkpoints {
		if checkpoint.Message == "" {
			fmt.Fprintf(stdout, "%s %s (%d files)\n", checkpoint.ID, checkpoint.CreatedAt, checkpoint.FileCount)
		} else {
			fmt.Fprintf(stdout, "%s %s %q (%d files)\n", checkpoint.ID, checkpoint.CreatedAt, checkpoint.Message, checkpoint.FileCount)
		}
	}
	return nil
}

func runFiles(args []string, stdout io.Writer) error {
	root, flagArgs := splitOptionalPath(args)
	flags := flag.NewFlagSet("files", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}
	files, err := fileproject.ListStableFiles(root)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Fprintln(stdout, "No files.")
		return nil
	}
	for _, file := range files {
		fmt.Fprintf(stdout, "%s %s %s %s\n", file.ID, file.Kind, file.SHA256, file.Path)
	}
	return nil
}

func runDiff(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("diff", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	from := flags.String("from", "", "source checkpoint id")
	to := flags.String("to", "", "target checkpoint id")
	root, flagArgs := splitOptionalPath(args)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}
	if *from == "" && interactiveInput() {
		if err := runHistory([]string{root}, stdout); err != nil {
			return err
		}
	}
	if err := askValue(from, "Source checkpoint ID (--from)", "", stdout); err != nil {
		return err
	}
	entries, err := fileproject.DiffCheckpoint(root, *from, *to)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		fmt.Fprintln(stdout, "No changes.")
		return nil
	}
	for _, entry := range entries {
		fmt.Fprintf(stdout, "%s %s\n", entry.Status, entry.Path)
	}
	return nil
}

func runRestore(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("restore", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	checkpointID := flags.String("checkpoint", "", "checkpoint id")
	root, flagArgs := splitOptionalPath(args)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}
	if *checkpointID == "" && interactiveInput() {
		if err := runHistory([]string{root}, stdout); err != nil {
			return err
		}
	}
	if err := askValue(checkpointID, "Checkpoint ID (--checkpoint)", "", stdout); err != nil {
		return err
	}
	checkpoint, err := fileproject.RestoreCheckpoint(root, *checkpointID)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Restored %s (%d files)\n", checkpoint.ID, len(checkpoint.Files))
	return nil
}

func printUsage(output io.Writer) {
	fmt.Fprintln(output, `Edda — keep local project folders in sync with your server.

Upload an existing local project:
  edda login
  edda send ./my-book
  The first send offers to create a project or select an existing one.
  Later, use the same command to send your changes.

Download and work with a server project:
  edda projects               List your projects
  edda get                   Choose a project and a new destination folder
  edda status ./my-book      Inspect local changes (offline)
  edda send ./my-book        Upload changes
  edda take ./my-book        Receive and merge server changes

Other project commands:
  create      Create an empty project on the server
  attach      Connect a local folder to an existing project without uploading
  import      Upload into an empty project and connect the local folder
  history     List saved versions
  restore     Restore a saved version
  conflicts   List conflicts
  resolve     Choose a conflict resolution
  move        Move a file while preserving its identity
  logout      Remove the saved login

Server administration: backup, verify-backup, restore-backup
Local prototype tools: init, ids sync, save, checkpoint, files, diff
  Local prototype tools do not connect or upload your folder.

Run edda COMMAND --help (or edda help COMMAND) for options and examples.
In a terminal, missing required values are requested; explicit arguments skip
prompts. Optional settings retain defaults. Folder defaults to the current directory
where applicable; connected commands find .edda in parent directories.
Scripts must supply required values. JSON is opt-in via --json
on projects, create, history and import. Login is shared across commands.`)
}

func splitOptionalPath(args []string) (string, []string) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return ".", args
	}
	return args[0], args[1:]
}

func splitExistingOptionalPath(args []string) (string, []string) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return ".", args
	}
	info, err := os.Stat(args[0])
	if err == nil && info.IsDir() {
		return args[0], args[1:]
	}
	return ".", args
}
