package project

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

var (
	ErrInvalidTree       = errors.New("invalid project tree")
	ErrVersionConflict   = errors.New("project version conflict")
	ErrOperationConflict = errors.New("operation ID already used for a different request")
	ErrObjectIntegrity   = errors.New("project object missing or corrupt")
)

type TreeEntry struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Kind   string `json:"kind"` // file or directory; no content-role restrictions
	SHA256 string `json:"sha256,omitempty"`
	Bytes  int64  `json:"bytes"`
}

type ProjectVersion struct {
	ID          string      `json:"id"`
	ProjectID   string      `json:"projectId"`
	ParentID    string      `json:"parentId"`
	OperationID string      `json:"operationId"`
	Message     string      `json:"message"`
	CreatedAt   string      `json:"createdAt"`
	Entries     []TreeEntry `json:"entries"`
}

// PublishVersionInput describes a complete replacement tree, not a patch.
// ExpectedVersion is empty only when publishing a project's first version.
// Callers retain entry IDs for renames; copies receive new IDs.
type PublishVersionInput struct {
	AuthorID        string
	ProjectID       string
	ExpectedVersion string
	OperationID     string
	Message         string
	Entries         []TreeEntry
}

// VersionLimits bound staging and manifest work. Zero fields use defaults.
type VersionLimits struct {
	MaxObjectBytes  int64
	MaxProjectBytes int64
	MaxEntries      int
}

func validTreeID(id string) bool {
	if len(id) == 0 || len(id) > 128 {
		return false
	}
	for _, r := range id {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

func validObjectHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	for _, r := range hash {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}

func validateTreePath(name string) error {
	if name == "." || !utf8.ValidString(name) || !fs.ValidPath(name) || len(name) > 4096 || strings.ContainsAny(name, "\\:") {
		return fmt.Errorf("%w: invalid relative path %q", ErrInvalidTree, name)
	}
	for _, part := range strings.Split(name, "/") {
		switch part {
		case ".edda", ".git", "node_modules", "__pycache__", ".DS_Store":
			return fmt.Errorf("%w: reserved local path %q", ErrInvalidTree, name)
		}
		if part == ".env" || strings.HasPrefix(part, ".env.") {
			return fmt.Errorf("%w: local environment file %q is excluded", ErrInvalidTree, name)
		}

		if len(part) > 255 || strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			return fmt.Errorf("%w: unsupported filename %q", ErrInvalidTree, name)
		}
		for _, r := range part {
			if unicode.IsControl(r) {
				return fmt.Errorf("%w: control character in path", ErrInvalidTree)
			}
		}
		if strings.ContainsAny(part, `<>"|?*`) {
			return fmt.Errorf("%w: unsupported filename %q", ErrInvalidTree, name)
		}
		stem := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		switch stem {
		case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
			return fmt.Errorf("%w: reserved filename %q", ErrInvalidTree, name)
		}
	}
	return nil
}

func canonicalTree(input PublishVersionInput, limits VersionLimits, restoredFrom string) ([]TreeEntry, string, error) {
	if !validTreeID(input.OperationID) || len(input.Message) > 4096 || !utf8.ValidString(input.Message) || len(input.Entries) > limits.MaxEntries {
		return nil, "", fmt.Errorf("%w: operation, message or entry limit", ErrInvalidTree)
	}
	entries := slices.Clone(input.Entries)
	if entries == nil {
		entries = []TreeEntry{}
	}
	paths := make(map[string]string, len(entries))
	ids := make(map[string]bool, len(entries))
	folded := make(map[string]bool, len(entries))
	var total int64
	for _, entry := range entries {
		if err := validateTreePath(entry.Path); err != nil {
			return nil, "", err
		}
		if !validTreeID(entry.ID) || ids[entry.ID] || folded[strings.ToLower(norm.NFC.String(entry.Path))] {
			return nil, "", fmt.Errorf("%w: duplicate path or invalid/duplicate ID", ErrInvalidTree)
		}
		ids[entry.ID] = true
		folded[strings.ToLower(norm.NFC.String(entry.Path))] = true
		paths[entry.Path] = entry.Kind
		switch entry.Kind {
		case "directory":
			if entry.SHA256 != "" || entry.Bytes != 0 {
				return nil, "", fmt.Errorf("%w: directory has file data", ErrInvalidTree)
			}
		case "file":
			if !validObjectHash(entry.SHA256) || entry.Bytes < 0 || entry.Bytes > limits.MaxObjectBytes || entry.Bytes > limits.MaxProjectBytes-total {
				return nil, "", fmt.Errorf("%w: invalid object or byte limit", ErrInvalidTree)
			}
			total += entry.Bytes
		default:
			return nil, "", fmt.Errorf("%w: unsupported entry kind", ErrInvalidTree)
		}
	}
	for _, entry := range entries {
		if parent := path.Dir(entry.Path); parent != "." && paths[parent] != "directory" {
			return nil, "", fmt.Errorf("%w: missing parent directory for %q", ErrInvalidTree, entry.Path)
		}
	}
	slices.SortFunc(entries, func(a, b TreeEntry) int { return strings.Compare(a.Path, b.Path) })
	data, err := json.Marshal(struct {
		Base         string
		Message      string
		RestoredFrom string
		Entries      []TreeEntry
	}{input.ExpectedVersion, input.Message, restoredFrom, entries})
	if err != nil {
		return nil, "", err
	}
	if len(data) > (8<<20)-8192 {
		return nil, "", fmt.Errorf("%w: manifest exceeds transfer limit", ErrInvalidTree)
	}
	sum := sha256.Sum256(data)
	return entries, hex.EncodeToString(sum[:]), nil
}

// ValidateTree applies the public server's default portability and size limits
// before clients upload any bytes. Optional content-role mapping is irrelevant.
func ValidateTree(entries []TreeEntry) error {
	_, _, err := canonicalTree(PublishVersionInput{OperationID: "validate", Entries: entries}, VersionLimits{MaxObjectBytes: 64 << 20, MaxProjectBytes: 1 << 30, MaxEntries: 10000}, "")
	return err
}
