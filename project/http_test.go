package project

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

type restoreRevisionFake struct {
	item  ContentItem
	err   error
	input RestoreRevisionInput
}

func (fake *restoreRevisionFake) RestoreRevision(_ context.Context, input RestoreRevisionInput) (ContentItem, error) {
	fake.input = input
	if fake.err != nil {
		return ContentItem{}, fake.err
	}
	return fake.item, nil
}

func TestRestoreRevisionHTTPRejectsInvalidRevisionNumber(t *testing.T) {
	response := httptest.NewRecorder()
	request := restoreRevisionRequestWithParams("0", `{"expectedRevision":1}`)

	restoreRevisionHTTP(response, request, &restoreRevisionFake{})

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
	if !strings.Contains(response.Body.String(), "invalid revision number") {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestRestoreRevisionHTTPRejectsMalformedJSON(t *testing.T) {
	response := httptest.NewRecorder()
	request := restoreRevisionRequestWithParams("1", `{`)

	restoreRevisionHTTP(response, request, &restoreRevisionFake{})

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
	if !strings.Contains(response.Body.String(), "malformed JSON") {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestRestoreRevisionHTTPMapsServiceErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		err    error
		status int
		body   string
	}{
		"conflict":  {err: ErrConflict, status: http.StatusConflict, body: "conflict"},
		"not found": {err: sql.ErrNoRows, status: http.StatusNotFound, body: "not found"},
	} {
		t.Run(name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := restoreRevisionRequestWithParams("2", `{"expectedRevision":4,"reason":"restore"}`)
			fake := &restoreRevisionFake{err: tc.err}

			restoreRevisionHTTP(response, request, fake)

			if response.Code != tc.status {
				t.Fatalf("status = %d, want %d", response.Code, tc.status)
			}
			if !strings.Contains(response.Body.String(), tc.body) {
				t.Fatalf("body = %s", response.Body.String())
			}
			if fake.input.Reason != "restore" {
				t.Fatalf("reason = %q", fake.input.Reason)
			}
		})
	}
}

func TestRestoreRevisionHTTPSuccess(t *testing.T) {
	fake := &restoreRevisionFake{
		item: ContentItem{
			ID:              "content-1",
			ProjectID:       "project-1",
			Kind:            KindChapter,
			Title:           "Opening",
			Slug:            "opening",
			BodyMarkdown:    "Earlier body",
			MetadataJSON:    "{}",
			SortOrder:       1,
			CurrentRevision: 5,
		},
	}
	response := httptest.NewRecorder()
	request := restoreRevisionRequestWithParams("2", `{"expectedRevision":4,"reason":"restore"}`)

	restoreRevisionHTTP(response, request, fake)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if fake.input.ProjectID != "project-1" || fake.input.ContentID != "content-1" {
		t.Fatalf("input ids = %#v", fake.input)
	}
	if fake.input.RevisionNumber != 2 || fake.input.ExpectedRevision != 4 || fake.input.CreatedBy != "author" || fake.input.Reason != "restore" {
		t.Fatalf("input = %#v", fake.input)
	}
	if !strings.Contains(response.Body.String(), `"currentRevision":5`) {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func restoreRevisionRequestWithParams(revisionNumber string, body string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/restore", strings.NewReader(body))
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("projectID", "project-1")
	routeContext.URLParams.Add("contentID", "content-1")
	routeContext.URLParams.Add("revisionNumber", revisionNumber)
	return request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))
}
