package project

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

// ArchivePlan contains staged immutable objects, but does not change any version.
// The author reviews this tree before publishing through the normal version API.
type ArchivePlan struct {
	Entries []TreeEntry `json:"entries"`
	Omitted []string    `json:"omitted"`
}

func archiveExcluded(name string) bool {
	for _, part := range strings.Split(name, "/") {
		switch part {
		case ".git", ".edda", "node_modules", "__pycache__", ".DS_Store", ".env":
			return true
		}
		if strings.HasPrefix(part, ".env.") {
			return true
		}
	}
	return false
}

// StageArchive reads ZIP entries without extracting paths onto the filesystem.
func (s *VersionStore) StageArchive(ctx context.Context, author, project string, source io.ReaderAt, size int64) (ArchivePlan, error) {
	plan := ArchivePlan{Entries: []TreeEntry{}, Omitted: []string{}}
	if err := s.authorize(ctx, author, project); err != nil {
		return plan, err
	}
	invalid := func(reason string) (ArchivePlan, error) {
		return ArchivePlan{}, fmt.Errorf("%w: %s", ErrInvalidTree, reason)
	}
	if size > 64<<20 {
		return invalid("ZIP exceeds 64 MiB")
	}
	archive, err := zip.NewReader(source, size)
	if err != nil {
		return invalid("cannot read ZIP archive")
	}
	if len(archive.File) > s.limits.MaxEntries {
		return invalid("too many ZIP entries")
	}
	paths := map[string]TreeEntry{}
	explicit := map[string]bool{}
	files := map[string]*zip.File{}
	var total uint64
	for _, file := range archive.File {
		if err := ctx.Err(); err != nil {
			return ArchivePlan{}, err
		}
		name := strings.TrimSuffix(file.Name, "/")
		// Never clean an unsafe path into an apparently safe one, even if excluded.
		if name == "" || strings.HasPrefix(name, "/") || strings.ContainsAny(name, "\\:") {
			return invalid("invalid ZIP path")
		}
		for _, part := range strings.Split(name, "/") {
			if part == "" || part == "." || part == ".." {
				return invalid("invalid ZIP path")
			}
		}
		if !file.Mode().IsRegular() && !file.Mode().IsDir() {
			return invalid("ZIP links and special files are unsupported")
		}
		if explicit[name] {
			return invalid("duplicate ZIP path")
		}
		explicit[name] = true
		if archiveExcluded(name) {
			plan.Omitted = append(plan.Omitted, name)
			continue
		}
		if err := validateTreePath(name); err != nil {
			return ArchivePlan{}, err
		}
		entry := TreeEntry{Path: name, Kind: "directory"}
		if !file.FileInfo().IsDir() {
			if file.UncompressedSize64 > uint64(s.limits.MaxObjectBytes) || file.UncompressedSize64 > uint64(min(s.limits.MaxProjectBytes, 256<<20))-total {
				return invalid("ZIP expanded size exceeds limit")
			}
			total += file.UncompressedSize64
			entry.Kind = "file"
			entry.Bytes = int64(file.UncompressedSize64)
			entry.SHA256 = strings.Repeat("0", 64)
			files[name] = file
		}
		if old, ok := paths[name]; ok && (old.Kind != "directory" || entry.Kind != "directory") {
			return invalid("ZIP file blocks a directory")
		}
		paths[name] = entry
		for parent := path.Dir(name); parent != "."; parent = path.Dir(parent) {
			if old, ok := paths[parent]; ok && old.Kind != "directory" {
				return invalid("ZIP file blocks a directory")
			}
			paths[parent] = TreeEntry{Path: parent, Kind: "directory"}
		}
		if len(paths) > s.limits.MaxEntries {
			return invalid("too many expanded ZIP entries")
		}
	}
	for _, entry := range paths {
		entry.ID, err = versionID()
		if err != nil {
			return ArchivePlan{}, err
		}
		plan.Entries = append(plan.Entries, entry)
	}
	// Validate case/Unicode collisions and all inferred directories before staging.
	plan.Entries, _, err = canonicalTree(PublishVersionInput{OperationID: "archive", Entries: plan.Entries}, s.limits, "")
	if err != nil {
		return ArchivePlan{}, err
	}
	for index := range plan.Entries {
		entry := &plan.Entries[index]
		if entry.Kind != "file" {
			continue
		}
		if err := s.stageArchiveMember(ctx, author, project, files[entry.Path], entry); err != nil {
			return ArchivePlan{}, err
		}
	}
	return plan, nil
}

// stageArchiveMember bounds heap use independently of the expanded member size.
// Its private spool is removed on success, corrupt input, cancellation or upload failure.
func (s *VersionStore) stageArchiveMember(ctx context.Context, author, project string, member *zip.File, entry *TreeEntry) error {
	id, err := versionID()
	if err != nil {
		return err
	}
	name := ".archive-" + id
	spool, err := s.objects.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer s.objects.Remove(name)
	defer spool.Close()

	reader, err := member.Open()
	if err != nil {
		return fmt.Errorf("%w: cannot open ZIP member", ErrInvalidTree)
	}
	digest := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(spool, digest), io.LimitReader(contextReader{ctx: ctx, r: reader}, entry.Bytes+1))
	closeErr := reader.Close()
	if err := ctx.Err(); err != nil {
		return err
	}
	// Filesystem failures are operational errors, not invalid archive input.
	var diskErr *os.PathError
	if errors.As(copyErr, &diskErr) {
		return copyErr
	}
	if copyErr != nil || closeErr != nil || n != entry.Bytes {
		return fmt.Errorf("%w: ZIP member failed size or checksum verification", ErrInvalidTree)
	}
	entry.SHA256 = hex.EncodeToString(digest.Sum(nil))
	if _, err := spool.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return s.UploadObject(ctx, author, project, entry.SHA256, entry.Bytes, spool)
}
