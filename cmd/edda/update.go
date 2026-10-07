package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"syscall"

	"github.com/InkyQuill/open-edda/fileproject"
)

func runNetworkTake(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("take", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root, rest := splitOptionalPath(args)
	restart := flags.Bool("restart", false, "archive an un-applied plan and inspect the current files again")
	if err := flags.Parse(rest); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("take accepts one checkout path")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	if _, err = readCheckout(root); err != nil {
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if state.Update != "" {
		p, err := loadUpdate(root, state)
		if err != nil {
			return err
		}
		switch p.Phase {
		case "applying", "committed", "rolled-back":
			if err := recoverUpdate(root, state, p); err != nil {
				return err
			}
			fmt.Fprintf(output, "Recovered interrupted update. Snapshots and displaced edits: %s\n", filepath.Join(root, ".edda", state.Update))
			return nil
		case "ready":
			if !*restart {
				return finishTake(ctx, root, state, p, output)
			}
			state.Update = ""
			if err = writePrivateJSON(checkoutPath(root), state); err != nil {
				return err
			}
		default:
			return errors.New("unknown update phase")
		}
	}
	c, err := resolveConnection(state.Server)
	if err != nil {
		return err
	}
	client, err := newImportClient(c.Server, state.Base.ProjectID, c.Token)
	if err != nil {
		return err
	}
	if state.Pending != nil {
		receipt, receiptErr := client.version(ctx, "operations/"+state.Pending.Operation)
		if receiptErr == nil {
			if err = acknowledgeSend(root, state, receipt, output); err != nil {
				return err
			}
			state, err = readCheckout(root)
			if err != nil {
				return err
			}
		} else {
			var status *importHTTPError
			if !errors.As(receiptErr, &status) || status.code != http.StatusNotFound {
				return receiptErr
			}
		}
	}
	remote, err := client.version(ctx, "versions/current")
	if err != nil {
		return err
	}
	if remote.ID == state.Base.ID {
		if state.Pending != nil {
			return errors.New("send is still unacknowledged; retry send before taking updates")
		}
		fmt.Fprintln(output, "Already at the remote version. Local edits are unchanged.")
		return nil
	}
	if err = validateRemoteSelection(state, remote.Entries); err != nil {
		return err
	}
	inventory, err := fileproject.ScanInventory(ctx, root, state.Excludes)
	if err != nil {
		return err
	}
	if len(inventory.Problems) > 0 {
		return errors.New("unsupported local entries; inspect edda import --dry-run before updating")
	}
	directory, err := os.MkdirTemp(filepath.Join(root, ".edda"), "update-")
	if err != nil {
		return err
	}
	recorded := false
	defer func() {
		if !recorded {
			os.RemoveAll(directory)
		}
	}()
	for _, name := range []string{"local", "remote", "base"} {
		if err = os.Mkdir(filepath.Join(directory, name), 0700); err != nil {
			return err
		}
	}
	snapshot, err := localSnapshot(ctx, root, filepath.Join(directory, "local"))
	if err != nil {
		return err
	}
	if err = downloadCheckout(ctx, client, remote, filepath.Join(directory, "remote"), state.Server); err != nil {
		return err
	}
	if err = downloadCheckout(ctx, client, state.Base, filepath.Join(directory, "base"), state.Server); err != nil {
		return err
	}
	current, err := localSnapshot(ctx, root, "")
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(current, snapshot) {
		return errors.New("local files changed while preparing; retry take")
	}
	verify, err := fileproject.ScanInventory(ctx, root, state.Excludes)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(inventory, verify) {
		return errors.New("local inventory changed while preparing; retry take")
	}
	inventory.Entries = localIdentities(state, inventory.Entries)
	p := updatePlan{Before: state, Remote: remote, Local: inventory, Snapshot: snapshot, Choices: map[string]string{}, Phase: "ready"}
	_, p.Conflicts, err = mergeSelection(p)
	if err != nil {
		return err
	}
	state.Update = filepath.Base(directory)
	if err = saveUpdate(root, state, p); err != nil {
		return err
	}
	// Once publishing the reference is attempted, retain snapshots even if the
	// final directory sync fails after the reference rename succeeded.
	recorded = true
	if err = writePrivateJSON(checkoutPath(root), state); err != nil {
		return err
	}
	return finishTake(ctx, root, state, p, output)
}
func finishTake(ctx context.Context, root string, state checkout, p updatePlan, output io.Writer) error {
	_, conflicts, err := mergeSelection(p)
	if err != nil {
		return err
	}
	if len(conflicts) > 0 {
		for _, name := range conflicts {
			if !conflictPathExists(p, name) {
				p.Conflicts = append(p.Conflicts, name)
			}
		}
		if err := saveUpdate(root, state, p); err != nil {
			return err
		}
		for _, name := range conflicts {
			fmt.Fprintf(output, "Conflict: %q\n", name)
		}
		fmt.Fprintf(output, "Base/local/remote snapshots: %s\n", filepath.Join(root, ".edda", state.Update))
		return errors.New("update requires choices; use edda resolve --path PATH --use local|remote, then edda take")
	}
	if err = applyUpdate(ctx, root, state, p); err != nil {
		return fmt.Errorf("update stopped; run take to recover or inspect status: %w", err)
	}
	fmt.Fprintf(output, "Updated to %s. Local-only changes remain unsent. Backups: %s\n", p.Remote.ID, filepath.Join(root, ".edda", state.Update))
	return nil
}
func runNetworkConflicts(root string, output io.Writer) error {
	state, err := readCheckout(root)
	if err != nil {
		return err
	}
	if state.Update == "" {
		fmt.Fprintln(output, "No prepared update conflicts.")
		return nil
	}
	p, err := loadUpdate(root, state)
	if err != nil {
		return err
	}
	fmt.Fprintf(output, "Update phase: %s\nSnapshots: %s\n", p.Phase, filepath.Join(root, ".edda", state.Update))
	_, conflicts, err := mergeSelection(p)
	if err != nil {
		return err
	}
	for _, name := range conflicts {
		fmt.Fprintf(output, "Conflict: %q\n", name)
	}
	return nil
}
func runNetworkResolve(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("resolve", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root, rest := splitOptionalPath(args)
	name := flags.String("path", "", "conflicting relative path, or . for the whole tree")
	side := flags.String("use", "", "local or remote")
	if err := flags.Parse(rest); err != nil {
		return err
	}
	if *name == "" && interactiveInput() {
		if err := runNetworkConflicts(root, output); err != nil {
			return err
		}
	}
	if err := askValue(name, "Conflict path (--path)", "", output); err != nil {
		return err
	}
	if err := askChoice(side, "Use version (--use)", []string{"local", "remote"}, output); err != nil {
		return err
	}
	if flags.NArg() != 0 || !safeConflictPath(*name) || *name == "" || (*side != "local" && *side != "remote") {
		return errors.New("usage: edda resolve CHECKOUT --path PATH --use local|remote")
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
	p, err := loadUpdate(root, state)
	if err != nil {
		return err
	}
	if p.Phase != "ready" {
		return errors.New("run take to recover the interrupted application first")
	}
	if !conflictPathExists(p, *name) {
		return errors.New("path is not a recorded conflict")
	}
	if p.Choices == nil {
		p.Choices = map[string]string{}
	}
	p.Choices[*name] = *side
	if err = saveUpdate(root, state, p); err != nil {
		return err
	}
	fmt.Fprintln(output, "Choice recorded. Run take to apply after resolving all conflicts.")
	return nil
}
