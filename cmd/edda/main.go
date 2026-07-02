package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"git.inkyquill.net/inky/writer/fileproject"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 {
		printUsage(stderr)
		return fmt.Errorf("command is required")
	}

	switch args[0] {
	case "get":
		return runGet(args[1:], stdout)
	case "status":
		return runStatus(args[1:], stdout)
	case "init":
		return runInit(args[1:], stdout)
	case "save":
		return runSave(args[1:], stdout)
	case "send":
		return runSend(args[1:], stdout)
	case "take":
		return runTake(args[1:], stdout)
	case "checkpoint":
		return runCheckpoint(args[1:], stdout)
	case "history":
		return runHistory(args[1:], stdout)
	case "diff":
		return runDiff(args[1:], stdout)
	case "restore":
		return runRestore(args[1:], stdout)
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
	writeIDs := flags.Bool("write-ids", false, "write .edda/ids.json")
	root, flagArgs := splitExistingOptionalPath(args)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}

	layout, err := fileproject.Scan(root)
	if err != nil {
		return err
	}
	if *writeIDs {
		idMap, files, err := fileproject.AssignStableIDs(root, layout)
		if err != nil {
			return err
		}
		if err := fileproject.WriteIDMap(root, idMap); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Updated .edda/ids.json for %d files.\n", len(files))
	}

	if layout.Metadata != nil {
		fmt.Fprintf(stdout, "Project: %s (%s)\n", layout.Metadata.Title, layout.Metadata.ID)
	} else {
		fmt.Fprintln(stdout, "Project: uninitialized Edda folder")
	}
	fmt.Fprintf(stdout, "Root: %s\n", layout.Root)

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

func runInit(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	title := flags.String("title", "", "project title")
	id := flags.String("id", "", "project id")
	serverURL := flags.String("server-url", "", "server URL")
	root, flagArgs := splitOptionalPath(args)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
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
	root, flagArgs := splitOptionalPath(args)
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
	if *fileID == "" {
		return fmt.Errorf("save requires --id")
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

func runGet(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("get", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	title := flags.String("title", "", "project title")
	id := flags.String("id", "", "project id")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() == 0 {
		return fmt.Errorf("get requires a server URL")
	}
	serverURL := flags.Arg(0)
	root := "."
	if flags.NArg() > 1 {
		root = flags.Arg(1)
	}
	projectTitle := strings.TrimSpace(*title)
	if projectTitle == "" {
		projectTitle = "Edda Project"
	}
	metadata, err := fileproject.InitMetadata(root, fileproject.InitMetadataInput{
		ID:        *id,
		Title:     projectTitle,
		ServerURL: serverURL,
	})
	if err != nil {
		return err
	}
	state, err := fileproject.EnsureSyncState(root)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Connected %s (%s) to %s as %s\n", metadata.Title, metadata.ID, metadata.ServerURL, state.DeviceID)
	return nil
}

func runSaveCheckpoint(root string, args []string, stdout io.Writer) error {
	message := strings.TrimSpace(strings.Join(args, " "))
	if message == "" {
		return fmt.Errorf("save requires a checkpoint note or file-save flags")
	}
	checkpoint, err := fileproject.CreateCheckpoint(root, fileproject.CreateCheckpointInput{Message: message})
	if err != nil {
		return err
	}
	if _, err := fileproject.RecordPendingUpload(root, checkpoint.ID); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Saved checkpoint %s (%d files); upload pending\n", checkpoint.ID, len(checkpoint.Files))
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
	checkpoint, err := fileproject.CreateCheckpoint(root, fileproject.CreateCheckpointInput{Message: *message})
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Checkpoint %s (%d files)\n", checkpoint.ID, len(checkpoint.Files))
	return nil
}

func runSend(args []string, stdout io.Writer) error {
	root, flagArgs := splitOptionalPath(args)
	flags := flag.NewFlagSet("send", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}
	metadata, err := fileproject.ReadMetadata(root)
	if err != nil {
		return err
	}
	if metadata.ServerURL == "" {
		if _, stateErr := fileproject.RecordPendingUploadFailure(root, "server URL is not configured"); stateErr != nil {
			return stateErr
		}
		return fmt.Errorf("server URL is not configured; run edda get or edda init --server-url")
	}
	state, err := fileproject.EnsureSyncState(root)
	if err != nil {
		return err
	}
	if state.PendingUpload == nil {
		fmt.Fprintln(stdout, "No pending upload.")
		return nil
	}
	checkpointID := state.PendingUpload.CheckpointID
	state, err = fileproject.CompletePendingUpload(root)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Sent checkpoint %s to %s from %s\n", checkpointID, metadata.ServerURL, state.DeviceID)
	return nil
}

func runTake(args []string, stdout io.Writer) error {
	root, flagArgs := splitOptionalPath(args)
	flags := flag.NewFlagSet("take", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}
	metadata, err := fileproject.ReadMetadata(root)
	if err != nil {
		return err
	}
	if metadata.ServerURL == "" {
		return fmt.Errorf("server URL is not configured; run edda get or edda init --server-url")
	}
	state, err := fileproject.RecordTake(root)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Checked %s for updates from %s\n", metadata.ServerURL, state.DeviceID)
	return nil
}

func runHistory(args []string, stdout io.Writer) error {
	root, flagArgs := splitOptionalPath(args)
	flags := flag.NewFlagSet("history", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}
	checkpoints, err := fileproject.ListCheckpoints(root)
	if err != nil {
		return err
	}
	if len(checkpoints) == 0 {
		fmt.Fprintln(stdout, "No checkpoints.")
		return nil
	}
	for _, checkpoint := range checkpoints {
		if checkpoint.Message == "" {
			fmt.Fprintf(stdout, "%s %s (%d files)\n", checkpoint.ID, checkpoint.CreatedAt.Format("2006-01-02T15:04:05Z"), len(checkpoint.Files))
		} else {
			fmt.Fprintf(stdout, "%s %s %q (%d files)\n", checkpoint.ID, checkpoint.CreatedAt.Format("2006-01-02T15:04:05Z"), checkpoint.Message, len(checkpoint.Files))
		}
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
	if *from == "" {
		return fmt.Errorf("diff requires --from")
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
	if *checkpointID == "" {
		return fmt.Errorf("restore requires --checkpoint")
	}
	checkpoint, err := fileproject.RestoreCheckpoint(root, *checkpointID)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Restored %s (%d files)\n", checkpoint.ID, len(checkpoint.Files))
	return nil
}

func printUsage(output io.Writer) {
	fmt.Fprintln(output, "Usage:")
	fmt.Fprintln(output, "  edda get [--title \"Title\"] [--id project-id] URL [path]")
	fmt.Fprintln(output, "  edda status [path] [--write-ids]")
	fmt.Fprintln(output, "  edda init [path] --title \"Title\" [--id project-id] [--server-url URL]")
	fmt.Fprintln(output, "  edda save [path] \"Checkpoint note\"")
	fmt.Fprintln(output, "  edda save [path] --id file-id (--from-draft | --body-file markdown.md) [--expected-sha256 HASH]")
	fmt.Fprintln(output, "  edda send [path]")
	fmt.Fprintln(output, "  edda take [path]")
	fmt.Fprintln(output, "  edda checkpoint [path] [--message \"Message\"]")
	fmt.Fprintln(output, "  edda history [path]")
	fmt.Fprintln(output, "  edda diff [path] --from checkpoint-id [--to checkpoint-id]")
	fmt.Fprintln(output, "  edda restore [path] --checkpoint checkpoint-id")
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
