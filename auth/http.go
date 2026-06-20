package auth

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"git.inkyquill.net/inky/writer/internal/httputil"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, service *Service) {
	registerRoutesWithLoginLimiter(r, service, newLoginLimiter(5, 2*time.Second, 1000))
}

func registerRoutesWithLoginLimiter(r chi.Router, service *Service, limiter *loginLimiter) {
	if service == nil {
		return
	}
	if limiter == nil {
		limiter = newLoginLimiter(5, 2*time.Second, 1000)
	}
	h := httpHandler{service: service, loginLimiter: limiter}
	r.Post("/auth/login", h.login)
}

type httpHandler struct {
	service      *Service
	loginLimiter *loginLimiter
}

func (h *httpHandler) login(w http.ResponseWriter, r *http.Request) {
	if !h.loginLimiter.allow(httputil.RemoteIP(r)) {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many login attempts"})
		return
	}

	var req LoginRequest
	if err := httputil.DecodeJSON(w, r, &req, httputil.DefaultJSONBodyLimit); err != nil {
		if httputil.IsRequestTooLarge(err) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidEmail) || errors.Is(err, ErrPasswordTooShort) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "login failed"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	httputil.WriteJSON(w, status, value)
}

type loginLimiter struct {
	mu          sync.Mutex
	burst       int
	refillEvery time.Duration
	maxEntries  int
	entries     map[string]*loginLimitEntry
}

type loginLimitEntry struct {
	tokens   int
	last     time.Time
	lastSeen time.Time
}

func newLoginLimiter(burst int, refillEvery time.Duration, maxEntries int) *loginLimiter {
	return &loginLimiter{
		burst:       burst,
		refillEvery: refillEvery,
		maxEntries:  maxEntries,
		entries:     make(map[string]*loginLimitEntry),
	}
}

func (l *loginLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	entry, ok := l.entries[key]
	if !ok {
		entry = &loginLimitEntry{tokens: l.burst, last: now, lastSeen: now}
		l.entries[key] = entry
	}
	l.refill(entry, now)
	entry.lastSeen = now

	if l.maxEntries > 0 && len(l.entries) > l.maxEntries {
		l.prune(now, key)
	}

	if entry.tokens <= 0 {
		return false
	}
	entry.tokens--
	return true
}

func (l *loginLimiter) refill(entry *loginLimitEntry, now time.Time) {
	if entry.tokens >= l.burst || l.refillEvery <= 0 {
		entry.tokens = l.burst
		entry.last = now
		return
	}

	refills := int(now.Sub(entry.last) / l.refillEvery)
	if refills <= 0 {
		return
	}

	entry.tokens += refills
	if entry.tokens >= l.burst {
		entry.tokens = l.burst
		entry.last = now
		return
	}
	entry.last = entry.last.Add(time.Duration(refills) * l.refillEvery)
}

func (l *loginLimiter) prune(now time.Time, activeKey string) {
	for key, entry := range l.entries {
		if key == activeKey {
			continue
		}
		if entry.tokens < l.burst {
			l.refill(entry, now)
		}
		if entry.tokens >= l.burst {
			delete(l.entries, key)
		}
		if len(l.entries) <= l.maxEntries {
			return
		}
	}

	for len(l.entries) > l.maxEntries {
		var oldestKey string
		var oldestSeen time.Time
		for key, entry := range l.entries {
			if key == activeKey {
				continue
			}
			if oldestKey == "" || entry.lastSeen.Before(oldestSeen) {
				oldestKey = key
				oldestSeen = entry.lastSeen
			}
		}
		if oldestKey == "" {
			return
		}
		delete(l.entries, oldestKey)
	}
}
