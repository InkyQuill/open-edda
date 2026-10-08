package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"

	"github.com/InkyQuill/open-edda/fileproject"
	"github.com/InkyQuill/open-edda/project"
	"golang.org/x/sys/unix"
)

func runNetworkGet(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("get", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	destination, flagArgs := splitOptionalPath(args)
	server := flags.String("server", "", "server override")
	id := flags.String("project", "", "project ID")
	versionID := flags.String("version", "current", "saved version ID")
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() != 0 || destination == "." {
		return errors.New("usage: edda get NEW_DIRECTORY --project ID [--server URL]")
	}
	if err := askProject(id, *server, output); err != nil {
		return err
	}
	absolute, err := filepath.Abs(destination)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(absolute); err == nil {
		return errors.New("destination already exists; get never replaces local work")
	} else if !os.IsNotExist(err) {
		return err
	}
	c, err := resolveConnection(*server)
	if err != nil {
		return err
	}
	client, err := newImportClient(c.Server, *id, c.Token)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	version, err := client.version(ctx, "versions/"+url.PathEscape(*versionID))
	if err != nil {
		return err
	}
	stage, err := os.MkdirTemp(filepath.Dir(absolute), ".edda-get-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	if err = downloadCheckout(ctx, client, version, stage, c.Server); err != nil {
		return err
	}
	// Linux no-replace rename prevents a destination created during download from
	// being overwritten. All files and metadata are durable before installation.
	if err = unix.Renameat2(unix.AT_FDCWD, stage, unix.AT_FDCWD, absolute, unix.RENAME_NOREPLACE); err != nil {
		return fmt.Errorf("install checkout without overwriting: %w", err)
	}
	if err = syncDirectory(filepath.Dir(absolute)); err != nil {
		return err
	}
	fmt.Fprintf(output, "Downloaded version %s to %s.\n", version.ID, absolute)
	return nil
}

func networkRoot(command string, args []string) (string, error) {
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root, rest := splitOptionalPath(args)
	if err := flags.Parse(rest); err != nil {
		return "", err
	}
	if flags.NArg() != 0 {
		return "", fmt.Errorf("%s accepts one checkout directory", command)
	}
	return filepath.Abs(root)
}
func runNetworkStatus(root string, output io.Writer) error {
	state, err := readCheckout(root)
	if err != nil {
		return err
	}
	inventory, err := fileproject.ScanInventory(context.Background(), root, state.localExclusions(), state.Base.Entries, state.Identity)
	if err != nil {
		return err
	}
	reportSyncExclusions(inventory, output)
	fmt.Fprintf(output, "Server: %s\nProject: %s\nBase version: %s\n", state.Server, state.Base.ProjectID, state.Base.ID)
	for _, name := range state.Untracked {
		fmt.Fprintf(output, "Local only (edda rm): %q\n", name)
	}
	if state.Move != nil {
		fmt.Fprintln(output, "Unfinished move; run edda move CHECKOUT to recover.")
	}
	if state.Update != "" {
		fmt.Fprintf(output, "Prepared update: %s; run edda conflicts or take to resolve/recover.\n", state.Update)
	}
	if state.Pending != nil {
		fmt.Fprintf(output, "Pending send: %s (retry edda send to check the receipt)\n", state.Pending.Operation)
	}
	if len(inventory.Problems) > 0 {
		reportInventoryProblems(inventory, output)
		return errors.New("unsupported local entries")
	}
	if sameFiles(inventory.Entries, state.Base.Entries) {
		fmt.Fprintln(output, "No local changes against the downloaded/acknowledged version.")
	} else {
		fmt.Fprintln(output, "Local changes are not sent.")
	}
	fmt.Fprintln(output, "Remote head was not checked (offline status).")
	return nil
}
func runNetworkSend(args []string, output io.Writer) error {
	root, rest := splitOptionalPath(args)
	flags := flag.NewFlagSet("send", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	quiet := flags.Bool("quiet", false, "hide progress")
	title := flags.String("title", "", "create a new project on first send")
	id := flags.String("project", "", "attach to an existing project on first send")
	server := flags.String("server", "", "server URL for first send")
	var excludes importExclusions
	flags.Var(&excludes, "exclude", "relative path to keep local on first send (repeatable)")
	if err := flags.Parse(rest); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("send accepts one folder")
	}
	if *title != "" && *id != "" {
		return errors.New("choose either --title for a new project or --project for an existing project")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(checkoutPath(root)); os.IsNotExist(err) {
		if err := prepareFirstSend(root, *server, *title, *id, excludes, output, *quiet); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if *title != "" || *id != "" || *server != "" || len(excludes) > 0 {
		return errors.New("this folder is already attached; run edda send FOLDER without first-send options")
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
		return errors.New("unfinished move; run edda move CHECKOUT to recover")
	}
	if state.Update != "" {
		return errors.New("update is in progress; run take or resolve before sending")
	}
	c, err := resolveConnection(state.Server)
	if err != nil {
		return err
	}
	client, err := newImportClient(c.Server, state.Base.ProjectID, c.Token)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, output, finishProgress := observeSync(ctx, output, *quiet)
	defer finishProgress()
	present := objectSizes(state.Base.Entries)

	if state.Pending == nil {
		inventory, err := fileproject.ScanInventory(ctx, root, state.localExclusions(), state.Base.Entries, state.Identity)
		if err != nil {
			return err
		}
		if len(inventory.Problems) > 0 {
			reportInventoryProblems(inventory, output)
			return errors.New("resolve unsupported local entries before sending")
		}
		reportSyncExclusions(inventory, output)
		syncPhase(ctx, "Checking remote manifest")
		head, err := client.version(ctx, "versions/current")
		if err != nil {
			return err
		}
		if head.ID != state.Base.ID {
			return errors.New("remote version advanced; local work is unchanged. Run edda take to reconcile changes")
		}
		if sameFiles(inventory.Entries, state.Base.Entries) {
			fmt.Fprintln(output, "No changes to send; remote version matches.")
			return nil
		}
		entries := localIdentities(state, inventory.Entries)
		stage, err := os.MkdirTemp(filepath.Join(root, ".edda"), "pending-")
		if err != nil {
			return err
		}
		// Unrecorded staging has no recovery role; once recorded, retain it on errors.
		recorded := false
		defer func() {
			if !recorded {
				os.RemoveAll(stage)
			}
		}()
		if err = fileproject.StageInventoryMissing(ctx, inventory, state.localExclusions(), stage, present); err != nil {
			return err
		}
		for _, entry := range inventory.Entries {
			if entry.Kind == "file" && !hasObject(present, entry) {
				file, err := os.Open(filepath.Join(stage, entry.ID))
				if err != nil {
					return err
				}
				syncErr := file.Sync()
				file.Close()
				if syncErr != nil {
					return syncErr
				}
			}
		}
		if err = syncDirectory(stage); err != nil {
			return err
		}
		state.Pending = &pendingSend{Operation: randomSyncID(), Directory: filepath.Base(stage), Inventory: inventory, Entries: entries}
		// Retain bytes if the state rename succeeds but its directory sync fails.
		recorded = true
		if err = writePrivateJSON(checkoutPath(root), state); err != nil {
			return err
		}
	}
	syncPhase(ctx, "Checking publication receipt")
	receipt, err := client.version(ctx, "operations/"+state.Pending.Operation)
	if err == nil {
		return acknowledgeSend(root, state, receipt, output)
	}
	var status *importHTTPError
	if !errors.As(err, &status) || status.code != http.StatusNotFound {
		return err
	}
	pending := state.Pending
	body, err := json.Marshal(struct {
		ExpectedVersion string              `json:"expectedVersion"`
		OperationID     string              `json:"operationId"`
		Entries         []project.TreeEntry `json:"entries"`
	}{state.Base.ID, pending.Operation, pending.Entries})
	if err != nil {
		return err
	}
	if len(body) > 8<<20 {
		return errors.New("manifest exceeds 8 MiB; pending send retained")
	}
	stagePath := filepath.Join(root, ".edda", pending.Directory)
	info, err := os.Lstat(stagePath)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("pending staging must be a real directory")
	}
	stage, err := os.OpenRoot(stagePath)
	if err != nil {
		return err
	}
	defer stage.Close()
	var sent, total int64
	var sentBytes atomic.Int64
	missing := map[string]bool{}
	for _, entry := range pending.Inventory.Entries {
		if entry.Kind == "file" && !hasObject(present, entry) && !missing[entry.SHA256] {
			missing[entry.SHA256] = true
			total++
		}
	}
	fileproject.ReportProgress(ctx, fileproject.Progress{Phase: "Uploading changed files", Total: total})
	for _, entry := range pending.Inventory.Entries {
		if entry.Kind != "file" {
			continue
		}
		if hasObject(present, entry) {
			continue
		}
		info, err := stage.Lstat(entry.ID)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return errors.New("staged file must be regular")
		}
		file, err := stage.OpenFile(entry.ID, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
		if err != nil {
			return err
		}
		completed := sent
		reader := progressReader{Reader: file, advance: func(n int) {
			sentBytes.Add(int64(n))
			fileproject.ReportProgress(ctx, fileproject.Progress{Phase: "Uploading changed files", Done: completed, Total: total, Bytes: sentBytes.Load(), Path: entry.Path})
		}}
		response, uploadErr := client.request(ctx, "PUT", "objects/"+entry.SHA256, reader, entry.Bytes)
		file.Close()
		if uploadErr != nil {
			return uploadErr
		}
		response.Body.Close()
		sent++
		fileproject.ReportProgress(ctx, fileproject.Progress{Phase: "Uploading changed files", Done: sent, Total: total, Bytes: sentBytes.Load(), Path: entry.Path})
		present[entry.SHA256] = entry.Bytes
	}
	syncPhase(ctx, "Publishing version (server verification)")
	response, err := client.request(ctx, "POST", "versions", bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return fmt.Errorf("send not acknowledged; prepared snapshot retained: %w", err)
	}
	defer response.Body.Close()
	if err = json.NewDecoder(io.LimitReader(response.Body, (8<<20)+1)).Decode(&receipt); err != nil {
		return err
	}
	return acknowledgeSend(root, state, receipt, output)
}
func acknowledgeSend(root string, state checkout, receipt project.ProjectVersion, output io.Writer) error {
	if !stagedMatchesReceipt(state, receipt) {
		return errors.New("publication receipt mismatch; prepared snapshot retained")
	}
	directory := state.Pending.Directory
	state.Base = receipt
	state.Identity = nil
	state.Pending = nil
	if err := writePrivateJSON(checkoutPath(root), state); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(root, ".edda", directory)); err != nil {
		fmt.Fprintln(output, "Sent; old staging could not be removed.")
	}
	fmt.Fprintf(output, "Acknowledged version %s. Any edits made after staging remain local; check edda status.\n", receipt.ID)
	return nil
}

func reportSyncExclusions(inventory fileproject.Inventory, output io.Writer) {
	if len(inventory.Excluded) > 0 {
		fmt.Fprintf(output, "Excluded %d paths/subtrees (.eddaignore, defaults and --exclude). Use edda import --dry-run --verbose to list them.\n", len(inventory.Excluded))
	}
}

func objectSizes(entries []project.TreeEntry) map[string]int64 {
	result := map[string]int64{}
	for _, entry := range entries {
		if entry.Kind == "file" {
			result[entry.SHA256] = entry.Bytes
		}
	}
	return result
}
func hasObject(objects map[string]int64, entry project.TreeEntry) bool {
	n, ok := objects[entry.SHA256]
	return ok && n == entry.Bytes
}
