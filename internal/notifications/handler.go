package notifications

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/mschl/ecommerce-api/internal/shared"
)

// Handler verarbeitet Benachrichtigungs-HTTP-Anfragen
type Handler struct {
	service NotificationService
}

// NewHandler erstellt einen neuen Benachrichtigungs-Handler
func NewHandler(service NotificationService) *Handler {
	return &Handler{service: service}
}

// List handles GET /notifications
// @Summary List notifications
// @Description Get all notifications for the current authenticated user. Notifications are ordered by creation date (newest first).
// @Tags notifications
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Success 200 {array} NotificationResponse "List of user notifications"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /notifications [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userVal := r.Context().Value(shared.UserContextKey)
	if userVal == nil {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// Type Assertion nutzen um User ID zu bekommen ohne auth-Paket zu importieren
	// Dies vermeidet Import-Zyklus: auth -> notifications -> auth
	type userWithID interface {
		GetID() uuid.UUID
	}
	user, ok := userVal.(userWithID)
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notifications, err := h.service.GetNotifications(r.Context(), user.GetID())
	if err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, ToResponseList(notifications))
}

// MarkAsRead handles PUT /notifications/{id}/read
// @Summary Mark notification as read
// @Description Mark a notification as read. Users can only mark their own notifications as read.
// @Tags notifications
// @Produce json
// @Security SessionToken
// @Param X-Session-Token header string true "Session token"
// @Param id path string true "Notification ID" format(uuid)
// @Success 200 {object} map[string]string "Notification marked as read"
// @Failure 400 {object} shared.ErrorResponse "Invalid notification ID format"
// @Failure 401 {object} shared.ErrorResponse "Missing or invalid session token"
// @Failure 403 {object} shared.ErrorResponse "Access denied (not your notification)"
// @Failure 404 {object} shared.ErrorResponse "Notification not found"
// @Failure 500 {object} shared.ErrorResponse "Internal server error"
// @Router /notifications/{id}/read [put]
func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userVal := r.Context().Value(shared.UserContextKey)
	if userVal == nil {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// Type Assertion nutzen um User ID zu bekommen ohne auth-Paket zu importieren
	// Dies vermeidet Import-Zyklus: auth -> notifications -> auth
	type userWithID interface {
		GetID() uuid.UUID
	}
	user, ok := userVal.(userWithID)
	if !ok {
		shared.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notificationID, err := h.parseUUID(r, "id")
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid notification ID")
		return
	}

	if err := h.service.MarkAsRead(r.Context(), user.GetID(), notificationID); err != nil {
		h.handleError(w, err)
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]string{"message": "notification marked as read"})
}

// parseUUID extrahiert eine UUID aus dem URL-Pfad
func (h *Handler) parseUUID(r *http.Request, key string) (uuid.UUID, error) {
	idStr := r.PathValue(key)
	if idStr == "" {
		return uuid.Nil, errors.New("missing id")
	}
	return uuid.Parse(idStr)
}

// handleError konvertiert Service-Fehler zu HTTP-Antworten
func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		shared.WriteError(w, http.StatusNotFound, "notification not found")
	case errors.Is(err, shared.ErrForbidden):
		shared.WriteError(w, http.StatusForbidden, "access denied")
	default:
		shared.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
