package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"syscall"

	"github.com/InkyQuill/open-edda/fileproject"
	"github.com/InkyQuill/open-edda/project"
)

type checkout struct {
	Untracked []string               `json:"untracked,omitempty"`
	Excludes  []string               `json:"excludes,omitempty"`
	Identity  []project.TreeEntry    `json:"identity,omitempty"`
	Move      *pendingMove           `json:"move,omitempty"`
	Update    string                 `json:"update,omitempty"`
	Schema    int                    `json:"schema"`
	Server    string                 `json:"server"`
	Base      project.ProjectVersion `json:"base"`
	Pending   *pendingSend           `json:"pending,omitempty"`
}
type pendingSend struct {
	Operation string                `json:"operation"`
	Directory string                `json:"directory"`
	Inventory fileproject.Inventory `json:"inventory"`
	Entries   []project.TreeEntry   `json:"entries"`
}

func checkoutPath(root string) string { return filepath.Join(root, ".edda", "checkout.json") }
func checkoutMetadata(root string) error {
	info, err := os.Lstat(filepath.Join(root, ".edda"))
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New(".edda must be a real directory")
	}
	return nil
}
func readCheckout(root string) (checkout, error) {
	var state checkout
	if err := checkoutMetadata(root); err != nil {
		return state, fmt.Errorf("not a network checkout; use edda get: %w", err)
	}
	info, err := os.Lstat(checkoutPath(root))
	if err != nil {
		return state, err
	}
	if !info.Mode().IsRegular() {
		return state, errors.New("checkout state must be a regular file")
	}
	if info.Size() > 64<<20 {
		return state, errors.New("checkout state exceeds limit")
	}
	data, err := os.ReadFile(checkoutPath(root))
	if err != nil {
		return state, err
	}
	if err = json.Unmarshal(data, &state); err != nil {
		return state, err
	}
	if state.Schema != 1 || state.Server == "" || state.Base.ID == "" || state.Base.ProjectID == "" {
		return state, errors.New("invalid checkout state")
	}
	if err = validateCheckoutTree(state.Base.Entries); err != nil {
		return state, err
	}
	if state.Identity != nil {
		if err := validateCheckoutTree(state.Identity); err != nil {
			return state, err
		}
	}
	if state.Move != nil {
		if err := validateCheckoutTree(state.Move.Entries); err != nil {
			return state, err
		}
		for _, name := range []string{state.Move.From, state.Move.To} {
			if !safeConflictPath(name) || name == "." || strings.Split(name, "/")[0] == ".edda" {
				return state, errors.New("invalid move path")
			}
		}
	}
	if state.Update != "" && (!strings.HasPrefix(state.Update, "update-") || filepath.Base(state.Update) != state.Update) {
		return state, errors.New("invalid update reference")
	}
	if p := state.Pending; p != nil {
		if !strings.HasPrefix(p.Directory, "pending-") || filepath.Base(p.Directory) != p.Directory || p.Operation == "" {
			return state, errors.New("invalid pending send")
		}
		if err = project.ValidateTree(p.Entries); err != nil {
			return state, err
		}
		if err = project.ValidateTree(p.Inventory.Entries); err != nil {
			return state, err
		}
		if !sameFiles(p.Inventory.Entries, p.Entries) {
			return state, errors.New("pending manifest does not match staged inventory")
		}
	}
	return state, nil
}
func validateCheckoutTree(entries []project.TreeEntry) error {
	if err := project.ValidateTree(entries); err != nil {
		return err
	}
	for _, entry := range entries {
		for _, part := range strings.Split(entry.Path, "/") {
			switch part {
			case ".edda", ".git", "node_modules", "__pycache__", ".DS_Store":
				return fmt.Errorf("remote path %q is reserved for local state/cache and cannot be checked out", entry.Path)
			}
			if part == ".env" || strings.HasPrefix(part, ".env.") {
				return fmt.Errorf("remote path %q is excluded from synchronization", entry.Path)
			}
		}
	}
	return nil
}
func sameFiles(a, b []project.TreeEntry) bool {
	if len(a) != len(b) {
		return false
	}
	byPath := map[string]project.TreeEntry{}
	for _, e := range a {
		e.ID = ""
		byPath[e.Path] = e
	}
	for _, e := range b {
		e.ID = ""
		if old, ok := byPath[e.Path]; !ok || old != e {
			return false
		}
	}
	return true
}
func randomSyncID() string { return "sync_" + rand.Text() }
func lockCheckout(root string) (func(), error) {
	if err := checkoutMetadata(root); err != nil {
		return nil, err
	}
	path := filepath.Join(root, ".edda", "sync.lock")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close()
		return nil, errors.New("another CLI operation holds this checkout")
	}
	return func() { syscall.Flock(int(file.Fd()), syscall.LOCK_UN); file.Close() }, nil
}

func downloadCheckout(ctx context.Context, client *importClient, version project.ProjectVersion, root, server string) error {
	if err := validateCheckoutTree(version.Entries); err != nil {
		return err
	}
	// The destination is a private, newly created directory, never an author tree.
	directories := []string{}
	for _, entry := range version.Entries {
		if entry.Kind == "directory" {
			directories = append(directories, entry.Path)
		}
	}
	sort.Slice(directories, func(i, j int) bool { return len(directories[i]) < len(directories[j]) })
	for _, name := range directories {
		if err := os.Mkdir(filepath.Join(root, filepath.FromSlash(name)), 0700); err != nil {
			return err
		}
	}
	for _, entry := range version.Entries {
		if entry.Kind != "file" {
			continue
		}
		response, err := client.request(ctx, "GET", "versions/"+url.PathEscape(version.ID)+"/entries/"+url.PathEscape(entry.ID), nil, 0)
		if err != nil {
			return err
		}
		file, err := os.OpenFile(filepath.Join(root, filepath.FromSlash(entry.Path)), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			response.Body.Close()
			return err
		}
		hash := sha256.New()
		n, copyErr := io.Copy(io.MultiWriter(file, hash), io.LimitReader(response.Body, entry.Bytes+1))
		response.Body.Close()
		if copyErr == nil {
			copyErr = file.Sync()
		}
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if n != entry.Bytes || hex.EncodeToString(hash.Sum(nil)) != entry.SHA256 {
			return fmt.Errorf("download checksum/size mismatch: %s", entry.Path)
		}
	}
	if err := os.Mkdir(filepath.Join(root, ".edda"), 0700); err != nil {
		return err
	}
	if err := writePrivateJSON(checkoutPath(root), checkout{Schema: 1, Server: server, Base: version}); err != nil {
		return err
	}
	for _, directory := range directories {
		if err := syncDirectory(filepath.Join(root, directory)); err != nil {
			return err
		}
	}
	return syncDirectory(root)
}

func stagedMatchesReceipt(state checkout, v project.ProjectVersion) bool {
	return state.Pending != nil && v.ID != "" && v.ProjectID == state.Base.ProjectID && v.ParentID == state.Base.ID && v.OperationID == state.Pending.Operation && reflect.DeepEqual(v.Entries, state.Pending.Entries)
}

func validateRemoteSelection(state checkout, entries []project.TreeEntry) error {
	if err := validateCheckoutTree(entries); err != nil {
		return err
	}
	for _, entry := range entries {
		for _, excluded := range state.Excludes {
			if entry.Path == excluded || strings.HasPrefix(entry.Path, excluded+"/") {
				return fmt.Errorf("remote path %q overlaps the local exclusion %q; reconcile explicitly", entry.Path, excluded)
			}
		}
	}
	return nil
}

func (s checkout) localExclusions() []string {
	return append(append([]string{}, s.Excludes...), s.Untracked...)
}
