package httputil

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
)

const DefaultJSONBodyLimit int64 = 1 << 20

func DecodeJSON(w http.ResponseWriter, r *http.Request, value any, limit int64) error {
	if limit <= 0 {
		limit = DefaultJSONBodyLimit
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, limit))
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if IsRequestTooLarge(err) {
			return err
		}
		return errors.New("trailing JSON data")
	}
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, value any) {
	body, err := json.Marshal(value)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal server error"}` + "\n"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(append(body, '\n'))
}

func IsRequestTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}

// RemoteIP returns the peer IP from RemoteAddr.
//
// It intentionally ignores X-Forwarded-For and X-Real-IP because those headers
// are client-spoofable unless the server has an explicit trusted-proxy model.
// Deployments behind a reverse proxy must rewrite RemoteAddr to the real client
// IP before requests reach Open Edda, or add trusted-proxy handling here.
func RemoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
