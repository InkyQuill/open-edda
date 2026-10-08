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
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/InkyQuill/open-edda/fileproject"
	"github.com/InkyQuill/open-edda/project"
)

type importExclusions []string

func (e *importExclusions) String() string         { return strings.Join(*e, ",") }
func (e *importExclusions) Set(value string) error { *e = append(*e, value); return nil }

func runImport(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("import", flag.ContinueOnError)
	flags.SetOutput(stderr)
	root, flagArgs := splitOptionalPath(args)
	server := flags.String("server", os.Getenv("OPEN_EDDA_URL"), "Open Edda base URL (or OPEN_EDDA_URL)")
	projectID := flags.String("project", "", "empty project ID")
	dryRun := flags.Bool("dry-run", false, "preview inventory; no network or source changes")
	verbose := flags.Bool("verbose", false, "list all included and excluded paths")
	machine := flags.Bool("json", false, "output JSON for scripts")
	var excludes importExclusions
	flags.Var(&excludes, "exclude", "exact relative file or directory to exclude (repeatable)")
	if err := flags.Parse(flagArgs); err != nil {
		return err
	}
	if flags.NArg() > 1 || (flags.NArg() > 0 && len(args) > 0 && !strings.HasPrefix(args[0], "-")) {
		return errors.New("import accepts one folder")
	}
	if flags.NArg() == 1 {
		root = flags.Arg(0)
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	root = absolute
	if !*dryRun {
		info, err := os.Lstat(root)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return errors.New("import requires a real directory")
		}
		if _, err := os.Lstat(filepath.Join(root, ".edda")); !os.IsNotExist(err) {
			return errors.New("folder already has .edda metadata; use edda send for a connected project; import never replaces an existing binding")
		}
	}
	var tracked []project.TreeEntry
	if *dryRun {
		if _, err := os.Lstat(checkoutPath(root)); err == nil {
			state, err := readCheckout(root)
			if err != nil {
				return err
			}
			excludes = append(excludes, state.localExclusions()...)
			tracked = append(tracked, state.Base.Entries...)
			tracked = append(tracked, state.Identity...)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	inventory, err := fileproject.ScanInventory(context.Background(), root, excludes, tracked)
	if err != nil {
		return err
	}
	if !*machine {
		printInventory(inventory, stdout)
		if *verbose {
			for _, notice := range inventory.Excluded {
				fmt.Fprintf(stdout, "  excluded %q (%s)\n", notice.Path, notice.Reason)
			}
			for _, entry := range inventory.Entries {
				fmt.Fprintf(stdout, "  %s %q (%d bytes)\n", entry.Kind, entry.Path, entry.Bytes)
			}
		}
	} else if *dryRun || len(inventory.Problems) > 0 {
		if err := json.NewEncoder(stdout).Encode(inventory); err != nil {
			return err
		}
	}
	if len(inventory.Problems) > 0 {
		return errors.New("inventory has unsupported entries; resolve them or use --exclude")
	}
	if *dryRun {
		if *machine {
			return nil
		}
		fmt.Fprintln(stdout, "Preview only. Upload and connect this folder with edda send FOLDER.")
		return nil
	}
	if *machine && *projectID == "" {
		return errors.New("--project is required with --json")
	}
	if err := askProject(projectID, *server, stderr); err != nil {
		return err
	}
	connection, err := resolveConnection(*server)
	if err != nil {
		return err
	}
	client, err := newImportClient(connection.Server, *projectID, connection.Token)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	operation, err := fileproject.ImportOperationID(inventory.Entries)
	if err != nil {
		return err
	}
	receipt, err := client.version(ctx, "operations/"+operation)
	if err == nil {
		if receipt.OperationID != operation || !reflect.DeepEqual(receipt.Entries, inventory.Entries) {
			return errors.New("server import receipt does not match the inventory")
		}
		if err := installBinding(root, connection.Server, receipt, excludes); err != nil {
			return fmt.Errorf("files are already imported, but saving the local binding failed; retry import or use edda attach: %w", err)
		}
		if *machine {
			return json.NewEncoder(stdout).Encode(receipt)
		}
		fmt.Fprintf(stdout, "Already imported as version %s; folder connected. Use edda send or edda take from inside the project.\n", receipt.ID)
		return nil
	}
	var status *importHTTPError
	if !errors.As(err, &status) || status.code != http.StatusNotFound {
		return err
	}
	head, err := client.version(ctx, "versions/current")
	if err != nil {
		return err
	}
	if len(head.Entries) > 0 {
		return errors.New("import requires an empty project; existing files were not changed")
	}
	publication := struct {
		ExpectedVersion string              `json:"expectedVersion"`
		OperationID     string              `json:"operationId"`
		Entries         []project.TreeEntry `json:"entries"`
	}{head.ID, operation, inventory.Entries}
	body, err := json.Marshal(publication)
	if err != nil {
		return err
	}
	if len(body) > 8<<20 {
		return errors.New("manifest exceeds the server's 8 MiB limit")
	}
	staging, err := os.MkdirTemp("", "edda-import-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	if err := fileproject.StageInventory(ctx, inventory, excludes, staging); err != nil {
		return err
	}
	for _, entry := range inventory.Entries {
		if entry.Kind != "file" {
			continue
		}
		file, err := os.Open(filepath.Join(staging, entry.ID))
		if err != nil {
			return err
		}
		response, uploadErr := client.request(ctx, "PUT", "objects/"+entry.SHA256, struct{ io.Reader }{file}, entry.Bytes)
		closeErr := file.Close()
		if uploadErr != nil {
			return uploadErr
		}
		response.Body.Close()
		if closeErr != nil {
			return closeErr
		}
	}
	response, err := client.request(ctx, "POST", "versions", bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var published project.ProjectVersion
	if err := json.NewDecoder(io.LimitReader(response.Body, (8<<20)+1)).Decode(&published); err != nil {
		return fmt.Errorf("read publication receipt (retry this import safely): %w", err)
	}
	if published.ProjectID != *projectID || published.OperationID != operation || !reflect.DeepEqual(published.Entries, inventory.Entries) {
		return errors.New("server publication receipt does not match the inventory; retry to check receipt")
	}
	if err := installBinding(root, connection.Server, published, excludes); err != nil {
		return fmt.Errorf("files were imported, but saving the local binding failed; retry import or use edda attach: %w", err)
	}
	if *machine {
		return json.NewEncoder(stdout).Encode(published)
	}
	fmt.Fprintf(stdout, "Imported %d entries (%d bytes) as version %s. Folder connected. Use edda send or edda take from inside the project.\n", len(inventory.Entries), inventory.Bytes, published.ID)
	return nil
}

type importHTTPError struct{ code int }

func (e *importHTTPError) Error() string {
	switch e.code {
	case 401:
		return "server rejected authentication; log in again or set OPEN_EDDA_TOKEN to a current token"
	case 403, 404:
		return "project or operation not found or inaccessible"
	case 409:
		return "project changed or is not a file project; nothing was overwritten"
	default:
		return fmt.Sprintf("server returned HTTP %d; retry the same import to check its receipt", e.code)
	}
}

type importClient struct {
	server    string
	root      string
	token     string
	projectID string
	http      *http.Client
}

func newImportClient(server, projectID, token string) (*importClient, error) {
	u, err := url.Parse(server)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.RawPath != "" {
		return nil, errors.New("--server or OPEN_EDDA_URL must be an HTTP(S) base URL without credentials, query or fragment")
	}
	if token == "" {
		return nil, errors.New("OPEN_EDDA_TOKEN is required; credentials are never written into project files")
	}
	if strings.TrimSpace(token) != token || strings.ContainsAny(token, "\r\n") {
		return nil, errors.New("OPEN_EDDA_TOKEN is malformed")
	}
	return &importClient{server: strings.TrimRight(server, "/"), root: strings.TrimRight(server, "/") + "/api/projects/" + url.PathEscape(projectID) + "/files/", token: token, projectID: projectID, http: &http.Client{Timeout: 2 * time.Minute, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func (c *importClient) request(ctx context.Context, method, path string, body io.Reader, length int64) (*http.Response, error) {
	if !strings.HasPrefix(path, "auth/") {
		token, err := refreshSavedAccess(ctx, c.server, c.token)
		if err != nil {
			return nil, err
		}
		c.token = token
	}
	request, err := http.NewRequestWithContext(ctx, method, c.root+path, body)
	if err != nil {
		return nil, err
	}
	request.ContentLength = length
	if body != nil && length == 0 {
		request.Body = http.NoBody
	}
	if c.token != "" {
		request.Header.Set("Authorization", "Bearer "+c.token)
	}
	if method == "POST" {
		request.Header.Set("Content-Type", "application/json")
	} else if method == "PUT" {
		request.Header.Set("Content-Type", "application/octet-stream")
	}
	response, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("contact server (retry this import safely): %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		response.Body.Close()
		return nil, &importHTTPError{response.StatusCode}
	}
	return response, nil
}
func (c *importClient) version(ctx context.Context, path string) (project.ProjectVersion, error) {
	var v project.ProjectVersion
	response, err := c.request(ctx, "GET", path, nil, 0)
	if err != nil {
		return v, err
	}
	defer response.Body.Close()
	err = json.NewDecoder(io.LimitReader(response.Body, (8<<20)+1)).Decode(&v)
	if err != nil {
		return v, err
	}
	if v.ID == "" || v.ProjectID != c.projectID {
		return v, errors.New("server returned an invalid project version")
	}
	return v, nil
}
