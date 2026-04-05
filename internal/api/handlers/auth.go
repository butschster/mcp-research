package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/butschster/mcp-research/internal/auth"
	"github.com/butschster/mcp-research/internal/domain"
	"github.com/butschster/mcp-research/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
	log     *slog.Logger
}

func NewAuthHandler(authSvc *service.AuthService, log *slog.Logger) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, log: log}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	user, token, err := h.authSvc.Register(r.Context(), input.Email, input.Password, input.Name)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailTaken):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, service.ErrRegistrationClosed):
			writeError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrPasswordTooShort),
			errors.Is(err, service.ErrInvalidEmail):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			h.log.Error("register failed", "error", err)
			writeError(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	user, token, err := h.authSvc.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, err.Error())
		} else {
			h.log.Error("login failed", "error", err)
			writeError(w, http.StatusInternalServerError, "login failed")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input struct {
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}

	plain, key, err := h.authSvc.CreateAPIKey(r.Context(), user.ID, input.Name)
	if err != nil {
		h.log.Error("create api key failed", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create API key")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"key":       plain,
		"id":        key.ID,
		"name":      key.Name,
		"prefix":    key.KeyPrefix,
		"created_at": key.CreatedAt,
	})
}

func (h *AuthHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	keys, err := h.authSvc.ListAPIKeys(r.Context(), user.ID)
	if err != nil {
		h.log.Error("list api keys failed", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list API keys")
		return
	}

	if keys == nil {
		keys = []*domain.APIKey{}
	}
	writeJSON(w, http.StatusOK, keys)
}

func (h *AuthHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	keyID := r.PathValue("id")
	if err := h.authSvc.DeleteAPIKey(r.Context(), keyID, user.ID); err != nil {
		writeError(w, http.StatusNotFound, "API key not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}

func (h *AuthHandler) AuthInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"auth_enabled":       true,
		"allow_registration": h.authSvc.AllowRegistration(),
	})
}
