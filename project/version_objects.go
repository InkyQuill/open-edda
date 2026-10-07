package project

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
)

// VersionStore owns the generic project-version feature. The caller owns db;
// Close releases only its filesystem handle. The data directory must be private
// to the service: external programs must use exported working copies, not edit
// these immutable objects. No garbage collection runs in this implementation.
type VersionStore struct {
	db      *sql.DB
	objects *os.Root
	limits  VersionLimits
}

func NewVersionStore(db *sql.DB, dataDir string, limits VersionLimits) (*VersionStore, error) {
	if db == nil {
		return nil, errors.New("version store requires a database")
	}
	if limits.MaxObjectBytes == 0 {
		limits.MaxObjectBytes = 64 << 20
	}
	if limits.MaxProjectBytes == 0 {
		limits.MaxProjectBytes = 1 << 30
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = 10000
	}
	if limits.MaxObjectBytes < 1 || limits.MaxObjectBytes == math.MaxInt64 || limits.MaxProjectBytes < 1 || limits.MaxEntries < 1 {
		return nil, errors.New("invalid version store limits")
	}
	// Require an existing data root so a typo cannot silently select new storage.
	root, err := os.OpenRoot(dataDir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	if err := root.Mkdir("objects", 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return nil, err
	}
	info, err := root.Lstat("objects")
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("objects must be a real directory")
	}
	objects, err := root.OpenRoot("objects")
	if err != nil {
		return nil, err
	}
	if err := syncRoot(root); err != nil {
		objects.Close()
		return nil, err
	}
	return &VersionStore{db: db, objects: objects, limits: limits}, nil
}

func (s *VersionStore) Close() error { return s.objects.Close() }

func syncRoot(root *os.Root) error {
	dir, err := root.Open(".")
	if err != nil {
		return err
	}
	return errors.Join(dir.Sync(), dir.Close())
}

func versionID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes[:]), nil
}

func (s *VersionStore) authorize(ctx context.Context, authorID, projectID string) error {
	var id string
	return s.db.QueryRowContext(ctx, "SELECT id FROM story_projects WHERE id = ? AND author_id = ?", projectID, authorID).Scan(&id)
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

// UploadObject verifies the declared bytes and grants only this project a
// reference. Knowing another project's digest never grants access. Failed
// uploads may leave unreferenced immutable bytes, never a published version.
func (s *VersionStore) UploadObject(ctx context.Context, authorID, projectID, hash string, size int64, reader io.Reader) error {
	if err := s.authorize(ctx, authorID, projectID); err != nil {
		return err
	}
	if !validObjectHash(hash) || size < 0 || size > s.limits.MaxObjectBytes || reader == nil {
		return fmt.Errorf("%w: invalid object declaration", ErrInvalidTree)
	}
	id, err := versionID()
	if err != nil {
		return err
	}
	temp := ".upload-" + id
	file, err := s.objects.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer s.objects.Remove(temp)
	defer file.Close()
	digest := sha256.New()
	n, err := io.Copy(io.MultiWriter(file, digest), io.LimitReader(contextReader{ctx, reader}, size+1))
	if err != nil {
		return err
	}
	if n != size || hex.EncodeToString(digest.Sum(nil)) != hash {
		return ErrObjectIntegrity
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	// Link publishes without overwriting a competing uploader's immutable file.
	// Staging lives in the same directory/filesystem; its name is never exposed.
	if err := s.objects.Link(temp, hash); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
		existing, err := s.openVerifiedObject(ctx, hash, size)
		if err != nil {
			return err
		}
		if err := existing.Close(); err != nil {
			return err
		}
	}
	if err := syncRoot(s.objects); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO project_objects(project_id,sha256,bytes)
 SELECT id,?,? FROM story_projects WHERE id=? AND author_id=?
 ON CONFLICT(project_id,sha256) DO UPDATE SET bytes=excluded.bytes WHERE project_objects.bytes=excluded.bytes`, hash, size, projectID, authorID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *VersionStore) openVerifiedObject(ctx context.Context, hash string, size int64) (*os.File, error) {
	if !validObjectHash(hash) || size < 0 {
		return nil, ErrObjectIntegrity
	}
	info, err := s.objects.Lstat(hash)
	if err != nil || !info.Mode().IsRegular() || info.Size() != size {
		return nil, fmt.Errorf("%w: %s", ErrObjectIntegrity, hash)
	}
	file, err := s.objects.Open(hash)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrObjectIntegrity, hash, err)
	}
	digest := sha256.New()
	n, err := io.Copy(digest, io.LimitReader(contextReader{ctx, file}, size+1))
	if err != nil {
		file.Close()
		return nil, err
	}
	if n != size || hex.EncodeToString(digest.Sum(nil)) != hash {
		file.Close()
		return nil, fmt.Errorf("%w: %s", ErrObjectIntegrity, hash)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}

// OpenVersionFile resolves bytes through an owned, committed version. No public
// raw-hash read exists. Integrity is verified before returning any content.
func (s *VersionStore) OpenVersionFile(ctx context.Context, authorID, projectID, version, entryID string) (io.ReadCloser, error) {
	var hash string
	var size int64
	err := s.db.QueryRowContext(ctx, `SELECT e.sha256,e.bytes FROM project_version_entries e
 JOIN story_projects p ON p.id=e.project_id
 WHERE p.author_id=? AND e.project_id=? AND e.version_id=? AND e.id=? AND e.kind='file'`, authorID, projectID, version, entryID).Scan(&hash, &size)
	if err != nil {
		return nil, err
	}
	return s.openVerifiedObject(ctx, hash, size)
}
