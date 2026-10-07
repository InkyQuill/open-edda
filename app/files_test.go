package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/InkyQuill/open-edda/auth"
	"github.com/InkyQuill/open-edda/project"
)

func TestFileProjectHTTPRoundTrip(t *testing.T) {
	db := openMigratedTestDB(t)
	versions, err := project.NewVersionStore(db, t.TempDir(), project.VersionLimits{MaxObjectBytes: 1024})
	if err != nil {
		t.Fatal(err)
	}
	defer versions.Close()
	secret := "file-test-secret-at-least-32-bytes-long"
	token, err := auth.GenerateToken("author-1", "author@example.com", secret)
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := auth.GenerateToken("other-author", "other@example.com", secret)
	if err != nil {
		t.Fatal(err)
	}
	handler := New(&Dependencies{AuthService: auth.NewService(db, secret), ProjectService: project.NewService(db), VersionStore: versions})
	call := func(method, path, body, credential string, status int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if credential != "" {
			r.Header.Set("Authorization", "Bearer "+credential)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s %s: %d want %d: %s", method, path, w.Code, status, w.Body.String())
		}
		return w
	}
	decodeVersion := func(w *httptest.ResponseRecorder) project.ProjectVersion {
		t.Helper()
		var v project.ProjectVersion
		if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	call("POST", "/api/projects", `{"title":"Book","storageMode":"files"}`, "", 401)
	call("POST", "/api/projects", `{"title":"Book","storageMode":"unknown"}`, token, 400)
	w := call("POST", "/api/projects", `{"title":"Book","language":"ru","storageMode":"files"}`, token, 201)
	var p project.StoryProject
	if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if p.StorageMode != "files" {
		t.Fatal(p)
	}
	root := "/api/projects/" + p.ID + "/files"
	initial := decodeVersion(call("GET", root+"/versions/current", "", token, 200))
	if initial.ID == "" || len(initial.Entries) != 0 {
		t.Fatal(initial)
	}
	list := call("GET", "/api/projects", "", token, 200)
	if !bytes.Contains(list.Body.Bytes(), []byte(`"storageMode":"files"`)) {
		t.Fatal(list.Body.String())
	}
	call("GET", root+"/versions/current", "", foreign, 404)
	call("GET", "/api/projects/project-1/files/versions/current", "", token, 409)
	call("POST", "/api/projects/"+p.ID+"/content", `{"kind":"chapter","title":"Hidden second tree"}`, token, 409)
	text := "# Глава\n\nТекст книги.\n"
	sum := sha256.Sum256([]byte(text))
	hash := hex.EncodeToString(sum[:])
	call("PUT", root+"/objects/"+hash, text, foreign, 404)
	call("PUT", root+"/objects/"+hash, "wrong bytes", token, 422)
	call("PUT", root+"/objects/"+hash, strings.Repeat("a", 1025), token, 413)
	call("PUT", root+"/objects/"+hash, text, token, 204)
	body := map[string]any{"expectedVersion": initial.ID, "operationId": "write-1", "entries": []project.TreeEntry{{ID: "chapter", Path: "глава.md", Kind: "file", SHA256: hash, Bytes: int64(len(text))}}}
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	published := decodeVersion(call("POST", root+"/versions", string(encoded), token, 201))
	retry := decodeVersion(call("POST", root+"/versions", string(encoded), token, 201))
	if retry.ID != published.ID {
		t.Fatal("retry created another version")
	}
	body["operationId"] = "write-2"
	encoded, _ = json.Marshal(body)
	call("POST", root+"/versions", string(encoded), token, 409)
	file := call("GET", root+"/versions/"+published.ID+"/entries/chapter", "", token, 200)
	if file.Body.String() != text || file.Header().Get("Content-Disposition") != "attachment" || file.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("file bytes or headers changed")
	}
	call("GET", root+"/versions/"+initial.ID+"/entries/chapter", "", token, 404)
	call("GET", root+"/versions/"+published.ID+"/entries/chapter", "", foreign, 404)
	call("POST", root+"/versions", `{"operationId":"bad","entries":[]} {}`, token, 400)
	call("POST", root+"/versions", `{"message":"`+strings.Repeat("x", 8<<20)+`"}`, token, 413)
	// Failure while making the initial manifest must not leave an orphan project.
	if _, err := db.Exec(`CREATE TRIGGER fail_new_tree BEFORE INSERT ON project_version_heads BEGIN SELECT RAISE(ABORT,'failed head'); END`); err != nil {
		t.Fatal(err)
	}
	call("POST", "/api/projects", `{"title":"Rollback","storageMode":"files"}`, token, 500)
	var count int
	if err := db.QueryRow("SELECT count(*) FROM story_projects WHERE title='Rollback'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("orphan project: %d %v", count, err)
	}
}
