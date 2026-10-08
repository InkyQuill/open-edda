package auth

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/InkyQuill/open-edda/internal/httputil"
)

const refreshCookie = "edda_refresh"

// Cookie mode uses a custom header, requiring a successful CORS preflight for
// cross-origin callers. The API does not enable cross-origin credentials.
func cookieMode(r *http.Request) bool { return r.Header.Get("X-Edda-Session") == "cookie" }

// HTTPS is required for browser sessions except loopback development. Proxy
// chains may replace X-Forwarded-Proto, so a public cookie is secure by default.
func secureSessionCookie(r *http.Request) bool {
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}
	host := r.Host
	if value, _, err := net.SplitHostPort(host); err == nil {
		host = value
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return false
	}
	ip := net.ParseIP(host)
	return ip == nil || !ip.IsLoopback()
}

func setRefreshCookie(w http.ResponseWriter, r *http.Request, token string, expires int64) {
	maxAge := int(expires - time.Now().Unix())
	if token == "" {
		maxAge = -1
	}
	http.SetCookie(w, &http.Cookie{Name: refreshCookie, Value: token, Path: "/api/auth", HttpOnly: true, Secure: secureSessionCookie(r), SameSite: http.SameSiteStrictMode, Expires: time.Unix(expires, 0), MaxAge: maxAge})
}
func writeSession(w http.ResponseWriter, r *http.Request, response AuthResponse) {
	w.Header().Set("Cache-Control", "no-store")
	if cookieMode(r) {
		setRefreshCookie(w, r, response.RefreshToken, response.RefreshExpiresAt)
		response.RefreshToken = ""
	}
	writeJSON(w, http.StatusOK, response)
}
func requestRefresh(w http.ResponseWriter, r *http.Request) (string, error) {
	if cookieMode(r) {
		cookie, err := r.Cookie(refreshCookie)
		if err != nil {
			return "", ErrInvalidRefresh
		}
		return cookie.Value, nil
	}
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := httputil.DecodeJSON(w, r, &body, 4096); err != nil {
		return "", err
	}
	return body.RefreshToken, nil
}
func (h httpHandler) refresh(w http.ResponseWriter, r *http.Request) {
	token, err := requestRefresh(w, r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh request"})
		return
	}
	response, err := h.service.Refresh(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrInvalidRefresh) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "session refresh failed"})
		}
		return
	}
	writeSession(w, r, response)
}
func (h httpHandler) logout(w http.ResponseWriter, r *http.Request) {
	token, err := requestRefresh(w, r)
	if err != nil && !errors.Is(err, ErrInvalidRefresh) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid logout request"})
		return
	}
	if err := h.service.RevokeRefresh(r.Context(), token); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "logout failed"})
		return
	}
	if cookieMode(r) {
		setRefreshCookie(w, r, "", 0)
	}
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusNoContent)
}
