package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MarkyGitHub/e-commerce/internal/shared"
)

// Handler verarbeitet Authentifizierungs-HTTP-Anfragen
type Handler struct {
	service AuthService
}

// NewHandler erstellt einen neuen Authentifizierungs-Handler
func NewHandler(service AuthService) *Handler {
	return &Handler{service: service}
}

// Register handles POST /auth/register
// @Summary Register a new user
// @Description Register a new customer account. Email must be valid format, password must be at least 8 characters, and name is required.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} UserResponse "User successfully registered"
// @Failure 400 {object} shared.ErrorResponse "Invalid request body or validation error"
// @Failure 409 {object} shared.ErrorResponse "Email already registered"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusCreated, user.ToResponse())
}

// Login handles POST /auth/login
// @Summary Login
// @Description Authenticate with email and password to receive a session token. The token must be included in subsequent requests via X-Session-Token header.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Login successful, returns session token"
// @Failure 400 {object} shared.ErrorResponse "Invalid request body"
// @Failure 401 {object} shared.ErrorResponse "Invalid credentials"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.service.Login(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, response)
}

// Logout handles POST /auth/logout
// @Summary Logout
// @Description Invalidate the current session token. After logout, the token can no longer be used for authentication.
// @Tags auth
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Success 204 "Logout successful"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get token from header (middleware already validated it)
	token := r.Header.Get("X-Session-Token")

	if err := h.service.Logout(r.Context(), token); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleError konvertiert Service-Fehler zu HTTP-Antworten
func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shared.ErrInvalidEmail):
		shared.WriteError(w, http.StatusBadRequest, "invalid email format")
	case errors.Is(err, shared.ErrPasswordTooShort):
		shared.WriteError(w, http.StatusBadRequest, "password must be at least 8 characters")
	case errors.Is(err, shared.ErrConflict):
		shared.WriteError(w, http.StatusConflict, "email already registered")
	case errors.Is(err, shared.ErrInvalidCredentials):
		shared.WriteError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, shared.ErrUnauthorized):
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, shared.ErrBadRequest):
		shared.WriteError(w, http.StatusBadRequest, "invalid request")
	default:
		shared.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
