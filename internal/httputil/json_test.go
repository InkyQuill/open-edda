package httputil

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONRejectsOversizedBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"value":"`+strings.Repeat("x", 20)+`"}`))
	rec := httptest.NewRecorder()
	var payload struct {
		Value string `json:"value"`
	}
	err := DecodeJSON(rec, req, &payload, 8)
	if err == nil {
		t.Fatal("DecodeJSON() error = nil, want size error")
	}
	if !IsRequestTooLarge(err) {
		t.Fatalf("DecodeJSON() error = %v, want request too large", err)
	}
}

func TestDecodeJSONRejectsTrailingData(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"value":"ok"} {}`))
	rec := httptest.NewRecorder()
	var payload struct {
		Value string `json:"value"`
	}
	err := DecodeJSON(rec, req, &payload, 1024)
	if err == nil {
		t.Fatal("DecodeJSON() error = nil, want trailing data error")
	}
	if IsRequestTooLarge(err) {
		t.Fatalf("DecodeJSON() error = %v, did not expect request too large", err)
	}
}

func TestWriteJSONWritesSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusCreated, map[string]string{"status": "ok"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q, want application/json", got)
	}
	if strings.TrimSpace(rec.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestWriteJSONHandlesMarshalFailureBeforeRequestedStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusCreated, map[string]any{"bad": func() {}})
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rec.Body.String(), "internal server error") {
		t.Fatalf("body = %q, want internal server error", rec.Body.String())
	}
}

func TestRemoteIPStripsPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "192.0.2.10:54321"
	if got := RemoteIP(req); got != "192.0.2.10" {
		t.Fatalf("RemoteIP() = %q, want 192.0.2.10", got)
	}
}

func TestIsRequestTooLargeUnwrapsMaxBytesError(t *testing.T) {
	err := &http.MaxBytesError{Limit: 1}
	if !IsRequestTooLarge(err) {
		t.Fatal("IsRequestTooLarge(MaxBytesError) = false")
	}
	if !IsRequestTooLarge(errors.Join(errors.New("decode"), err)) {
		t.Fatal("IsRequestTooLarge(joined MaxBytesError) = false")
	}
}
