package fileproject

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"syscall"

	"github.com/InkyQuill/open-edda/project"
)

// Inventory is a transport inventory, independent of CWS content roles or the
// old fixed-layout scanner. Exclusions are always reported to the caller.
type Inventory struct {
	Root     string              `json:"root"`
	Entries  []project.TreeEntry `json:"entries"`
	Excluded []InventoryNotice   `json:"excluded"`
	Problems []InventoryNotice   `json:"problems"`
	Bytes    int64               `json:"bytes"`
}
type InventoryNotice struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

func ScanInventory(ctx context.Context, directory string, exclusions []string) (Inventory, error) {
	result := Inventory{Entries: []project.TreeEntry{}, Excluded: []InventoryNotice{}, Problems: []InventoryNotice{}}
	for _, name := range exclusions {
		if name == "." || !fs.ValidPath(name) || strings.Contains(name, "\\") {
			return result, fmt.Errorf("exclude requires an exact relative path: %q", name)
		}
	}
	abs, err := filepath.Abs(directory)
	if err != nil {
		return result, err
	}
	result.Root = abs
	root, err := os.OpenRoot(abs)
	if err != nil {
		return result, err
	}
	defer root.Close()
	err = fs.WalkDir(root.FS(), ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if walkErr != nil {
			return walkErr
		}
		if name == "." {
			return nil
		}
		reason := excludedInventoryPath(name, exclusions)
		if reason != "" {
			result.Excluded = append(result.Excluded, InventoryNotice{name, reason})
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if len(result.Entries) >= 10000 {
			return fmt.Errorf("project exceeds 10000 entries")
		}
		if entry.Type()&os.ModeSymlink != 0 {
			result.Problems = append(result.Problems, InventoryNotice{name, "symbolic link; exclude this path explicitly"})
			return nil
		}
		id := sha256.Sum256([]byte("import:" + name))
		item := project.TreeEntry{ID: hex.EncodeToString(id[:]), Path: name, Kind: "directory"}
		if !entry.IsDir() {
			if !entry.Type().IsRegular() {
				result.Problems = append(result.Problems, InventoryNotice{name, "unsupported special file"})
				return nil
			}
			file, err := openInventoryFile(root, name)
			if err != nil {
				return err
			}
			hash := sha256.New()
			n, copyErr := io.Copy(hash, io.LimitReader(inventoryReader{ctx, file}, (64<<20)+1))
			closeErr := file.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
			if n > 64<<20 {
				return fmt.Errorf("file %q exceeds 64 MiB", name)
			}
			item.Kind = "file"
			item.SHA256 = hex.EncodeToString(hash.Sum(nil))
			item.Bytes = n
			result.Bytes += n
			if result.Bytes > 1<<30 {
				return fmt.Errorf("project exceeds 1 GiB")
			}
		}
		result.Entries = append(result.Entries, item)
		return nil
	})
	if err != nil {
		return result, fmt.Errorf("inventory: %w", err)
	}
	sort.Slice(result.Entries, func(i, j int) bool { return result.Entries[i].Path < result.Entries[j].Path })
	if err := project.ValidateTree(result.Entries); err != nil {
		result.Problems = append(result.Problems, InventoryNotice{Reason: err.Error()})
	}
	return result, nil
}

func excludedInventoryPath(name string, exclusions []string) string {
	for _, excluded := range exclusions {
		if name == excluded || strings.HasPrefix(name, excluded+"/") {
			return "explicit exclusion"
		}
	}
	switch path.Base(name) {
	case ".git", ".edda", "node_modules", "__pycache__", ".DS_Store":
		return "local repository, state or cache"
	}
	base := path.Base(name)
	if base == ".env" || strings.HasPrefix(base, ".env.") {
		return "local environment configuration"
	}
	return ""
}

func openInventoryFile(root *os.Root, name string) (*os.File, error) {
	before, err := root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() {
		return nil, fmt.Errorf("%q is no longer a regular file", name)
	}
	// Nonblocking protects against a regular file being swapped for a FIFO.
	file, err := root.OpenFile(name, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	after, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	if !after.Mode().IsRegular() || !os.SameFile(before, after) {
		file.Close()
		return nil, fmt.Errorf("%q changed while opening", name)
	}
	return file, nil
}

type inventoryReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r inventoryReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

// StageInventory freezes verified bytes outside the author folder. The caller
// owns the empty private staging directory and removes it on every exit path.
func StageInventory(ctx context.Context, inventory Inventory, exclusions []string, directory string) error {
	if len(inventory.Problems) > 0 {
		return fmt.Errorf("resolve inventory problems before importing")
	}
	root, err := os.OpenRoot(inventory.Root)
	if err != nil {
		return err
	}
	defer root.Close()
	for _, entry := range inventory.Entries {
		if entry.Kind != "file" {
			continue
		}
		source, err := openInventoryFile(root, entry.Path)
		if err != nil {
			return err
		}
		target, err := os.OpenFile(filepath.Join(directory, entry.ID), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			source.Close()
			return err
		}
		hash := sha256.New()
		n, copyErr := io.Copy(io.MultiWriter(target, hash), io.LimitReader(inventoryReader{ctx, source}, entry.Bytes+1))
		sourceErr := source.Close()
		targetErr := target.Close()
		if copyErr != nil {
			return copyErr
		}
		if sourceErr != nil {
			return sourceErr
		}
		if targetErr != nil {
			return targetErr
		}
		if n != entry.Bytes || hex.EncodeToString(hash.Sum(nil)) != entry.SHA256 {
			return fmt.Errorf("%q changed after inventory; run import again", entry.Path)
		}
	}
	// Detect additions/deletions and directory changes as well as altered bytes.
	current, err := ScanInventory(ctx, inventory.Root, exclusions)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(inventory, current) {
		return fmt.Errorf("project changed during staging; run import again")
	}
	return nil
}

// ImportOperationID makes retries deterministic without writing client state or
// credentials into the author's folder. It is scoped by the server project.
func ImportOperationID(entries []project.TreeEntry) (string, error) {
	data, err := json.Marshal(entries)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return "import_" + hex.EncodeToString(hash[:]), nil
}
