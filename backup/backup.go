// Package backup creates verified, consistent database-and-object snapshots.
// Immutable objects are never collected, so copying after a SQLite snapshot is
// safe while the single server continues to publish new versions.
package backup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/sys/unix"
)

type manifest struct {
	Schema   int              `json:"schema"`
	Database string           `json:"database"`
	Objects  map[string]int64 `json:"objects"`
}

func openRead(name string, immutable bool) (*sql.DB, error) {
	info, err := os.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("database must be a regular file")
	}
	name, err = filepath.Abs(name)
	if err != nil {
		return nil, err
	}
	u := url.URL{Scheme: "file", Path: name}
	q := u.Query()
	q.Set("mode", "ro")
	q.Set("_busy_timeout", "5000")
	if immutable {
		q.Set("immutable", "1")
	}
	u.RawQuery = q.Encode()
	return sql.Open("sqlite3", u.String())
}
func syncDir(name string) error {
	f, e := os.Open(name)
	if e != nil {
		return e
	}
	defer f.Close()
	return f.Sync()
}
func digest(name string) (string, error) {
	f, e := os.Open(name)
	if e != nil {
		return "", e
	}
	defer f.Close()
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		return "", e
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func refs(ctx context.Context, db *sql.DB) (map[string]int64, error) {
	var check string
	if err := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&check); err != nil {
		return nil, err
	}
	if check != "ok" {
		return nil, errors.New("database integrity check failed")
	}
	rows, err := db.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return nil, err
	}
	bad := rows.Next()
	rowErr := rows.Err()
	rows.Close()
	if rowErr != nil {
		return nil, rowErr
	}
	if bad {
		return nil, errors.New("database foreign key check failed")
	}
	rows, err = db.QueryContext(ctx, "SELECT DISTINCT sha256,bytes FROM project_version_entries WHERE kind='file'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]int64{}
	for rows.Next() {
		var hash string
		var size int64
		if err := rows.Scan(&hash, &size); err != nil {
			return nil, err
		}
		decoded, err := hex.DecodeString(hash)
		if err != nil || len(decoded) != 32 || strings.ToLower(hash) != hash || size < 0 {
			return nil, errors.New("invalid object reference")
		}
		if previous, ok := result[hash]; ok && previous != size {
			return nil, errors.New("inconsistent object sizes")
		}
		result[hash] = size
	}
	return result, rows.Err()
}
func copyObject(source, target, hash string, size int64) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() != size {
		return fmt.Errorf("invalid object %s", hash)
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer out.Close()
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(out, h), io.LimitReader(in, size+1))
	if err != nil {
		return err
	}
	if n != size || hex.EncodeToString(h.Sum(nil)) != hash {
		return fmt.Errorf("object checksum mismatch: %s", hash)
	}
	return out.Sync()
}
func stage(destination string) (string, error) {
	if _, err := os.Lstat(destination); !os.IsNotExist(err) {
		return "", errors.New("destination must not exist")
	}
	return os.MkdirTemp(filepath.Dir(destination), ".edda-backup-")
}
func install(stage, destination string) error {
	if err := syncDir(filepath.Join(stage, "objects")); err != nil {
		return err
	}
	if err := syncDir(stage); err != nil {
		return err
	}
	if err := unix.Renameat2(unix.AT_FDCWD, stage, unix.AT_FDCWD, destination, unix.RENAME_NOREPLACE); err != nil {
		return err
	}
	return syncDir(filepath.Dir(destination))
}

// Create uses VACUUM INTO, never a copy of a live WAL database file.
func Create(ctx context.Context, database, data, destination string) error {
	temp, err := stage(destination)
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	source, err := openRead(database, false)
	if err != nil {
		return err
	}
	defer source.Close()
	target := filepath.Join(temp, "edda.db")
	if _, err := source.ExecContext(ctx, "VACUUM INTO ?", target); err != nil {
		return err
	}
	f, err := os.Open(target)
	if err != nil {
		return err
	}
	err = f.Sync()
	f.Close()
	if err != nil {
		return err
	}
	snapshot, err := openRead(target, true)
	if err != nil {
		return err
	}
	defer snapshot.Close()
	objects, err := refs(ctx, snapshot)
	if err != nil {
		return err
	}
	if err := os.Mkdir(filepath.Join(temp, "objects"), 0700); err != nil {
		return err
	}
	for hash, size := range objects {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := copyObject(filepath.Join(data, "objects", hash), filepath.Join(temp, "objects", hash), hash, size); err != nil {
			return err
		}
	}
	dbHash, err := digest(target)
	if err != nil {
		return err
	}
	payload, err := json.MarshalIndent(manifest{1, dbHash, objects}, "", "  ")
	if err != nil {
		return err
	}
	m, err := os.OpenFile(filepath.Join(temp, "backup.json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = m.Write(payload)
	if err == nil {
		err = m.Sync()
	}
	closeErr := m.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return install(temp, destination)
}

// Verify checks the complete history's referenced bytes and the database hash.
func Verify(ctx context.Context, source string) error {
	_, err := verified(ctx, source)
	return err
}
func verified(ctx context.Context, source string) (manifest, error) {
	var m manifest
	payload, err := os.ReadFile(filepath.Join(source, "backup.json"))
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(payload, &m); err != nil {
		return m, err
	}
	if m.Schema != 1 {
		return m, errors.New("unsupported backup format")
	}
	hash, err := digest(filepath.Join(source, "edda.db"))
	if err != nil {
		return m, err
	}
	if hash != m.Database {
		return m, errors.New("backup database checksum mismatch")
	}
	db, err := openRead(filepath.Join(source, "edda.db"), true)
	if err != nil {
		return m, err
	}
	defer db.Close()
	objects, err := refs(ctx, db)
	if err != nil {
		return m, err
	}
	if len(objects) != len(m.Objects) {
		return m, errors.New("backup object inventory mismatch")
	}
	for hash, size := range objects {
		if err := ctx.Err(); err != nil {
			return m, err
		}
		stored, ok := m.Objects[hash]
		if !ok || stored != size {
			return m, errors.New("backup object inventory mismatch")
		}
		name := filepath.Join(source, "objects", hash)
		info, err := os.Lstat(name)
		if err != nil {
			return m, err
		}
		if !info.Mode().IsRegular() || info.Size() != size {
			return m, errors.New("backup object size mismatch")
		}
		actual, err := digest(name)
		if err != nil {
			return m, err
		}
		if actual != hash {
			return m, errors.New("backup object checksum mismatch")
		}
	}
	return m, nil
}

// Restore installs into a new data root only. It cannot overwrite a running DB.
func Restore(ctx context.Context, source, destination string) error {
	m, err := verified(ctx, source)
	if err != nil {
		return err
	}
	temp, err := stage(destination)
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	if err := os.Mkdir(filepath.Join(temp, "objects"), 0700); err != nil {
		return err
	}
	info, err := os.Stat(filepath.Join(source, "edda.db"))
	if err != nil {
		return err
	}
	if err := copyObject(filepath.Join(source, "edda.db"), filepath.Join(temp, "edda.db"), m.Database, info.Size()); err != nil {
		return err
	}
	for hash, size := range m.Objects {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := copyObject(filepath.Join(source, "objects", hash), filepath.Join(temp, "objects", hash), hash, size); err != nil {
			return err
		}
	}
	return install(temp, destination)
}
