package auth

import (
	"errors"
	"net/http"

	"git.inkyquill.net/inky/writer/internal/httputil"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, service *Service) {
	if service == nil {
		return
	}
	h := httpHandler{service: service}
	r.Post("/auth/login", h.login)
}

type httpHandler struct {
	service *Service
}

func (h *httpHandler) login(w http.ResponseWriter, r *http.Request) {
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
