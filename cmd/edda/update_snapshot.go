package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type localNode struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
	Hash string `json:"hash,omitempty"`
	Link string `json:"link,omitempty"`
	Mode uint32 `json:"mode"`
}

// Snapshot includes excluded local files as well as portable files. Only the
// checkout's own .edda is omitted. Symlinks are recorded, never followed.
func localSnapshot(ctx context.Context, source, destination string, only ...string) ([]localNode, error) {
	nodes := []localNode{}
	root, err := os.OpenRoot(source)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	err = fs.WalkDir(root.FS(), ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if name == "." {
			return nil
		}
		if len(only) > 0 && strings.Split(name, "/")[0] != only[0] {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if name == ".edda" {
			return fs.SkipDir
		}
		info, err := root.Lstat(name)
		if err != nil {
			return err
		}
		node := localNode{Path: name, Mode: uint32(info.Mode().Perm())}
		target := ""
		if destination != "" {
			target = filepath.Join(destination, filepath.FromSlash(name))
		}
		switch {
		case info.IsDir():
			node.Kind = "directory"
			if target != "" {
				if err := os.Mkdir(target, 0700); err != nil {
					return err
				}
			}
		case info.Mode()&os.ModeSymlink != 0:
			node.Kind = "symlink"
			node.Link, err = root.Readlink(name)
			if err != nil {
				return err
			}
			if target != "" {
				if err := os.Symlink(node.Link, target); err != nil {
					return err
				}
			}
		case info.Mode().IsRegular():
			node.Kind = "file"
			file, err := root.OpenFile(name, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
			if err != nil {
				return err
			}
			opened, err := file.Stat()
			if err != nil {
				file.Close()
				return err
			}
			if !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
				file.Close()
				return fmt.Errorf("local file changed during snapshot: %q", name)
			}
			hash := sha256.New()
			var writer io.Writer = hash
			var output *os.File
			if target != "" {
				output, err = os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
				if err != nil {
					file.Close()
					return err
				}
				writer = io.MultiWriter(hash, output)
			}
			_, copyErr := io.Copy(writer, &updateReader{ctx: ctx, reader: file})
			file.Close()
			if output != nil {
				if copyErr == nil {
					copyErr = output.Chmod(info.Mode().Perm())
				}
				if copyErr == nil {
					copyErr = output.Sync()
				}
				closeErr := output.Close()
				if copyErr == nil {
					copyErr = closeErr
				}
			}
			if copyErr != nil {
				return copyErr
			}
			node.Hash = hex.EncodeToString(hash.Sum(nil))
		default:
			return fmt.Errorf("cannot preserve special local file %q during update", name)
		}
		nodes = append(nodes, node)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if destination != "" {
		for i := len(nodes) - 1; i >= 0; i-- {
			if nodes[i].Kind == "directory" {
				path := filepath.Join(destination, filepath.FromSlash(nodes[i].Path))
				if err := os.Chmod(path, os.FileMode(nodes[i].Mode)); err != nil {
					return nil, err
				}
				if err := syncDirectory(path); err != nil {
					return nil, err
				}
			}
		}
		if err := syncDirectory(destination); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

type updateReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *updateReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}
