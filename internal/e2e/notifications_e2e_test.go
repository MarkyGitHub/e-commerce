package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mschl/ecommerce-api/internal/notifications"
)

func TestE2E_NotificationsRequireAuth(t *testing.T) {
	server, _ := newTestServer(t)
	client := server.Client()

	resp, _ := doJSONRequest(t, client, http.MethodGet, server.URL+"/notifications", nil, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestE2E_NotificationsListWithAuth(t *testing.T) {
	server, _ := newTestServer(t)
	client := server.Client()

	user := registerUser(t, client, server.URL, "customer-notify-1", "customer-notify-1@example.com", "password123")
	login := loginUser(t, client, server.URL, user.Email, "password123")

	resp, body := doJSONRequest(t, client, http.MethodGet, server.URL+"/notifications", nil, login.SessionToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var items []notifications.NotificationResponse
	require.NoError(t, json.Unmarshal(body, &items))
	assert.NotEmpty(t, items)
}

func TestE2E_MarkNotificationAsRead(t *testing.T) {
	server, database := newTestServer(t)
	client := server.Client()

	user := registerUser(t, client, server.URL, "customer-notify-2", "customer-notify-2@example.com", "password123")
	login := loginUser(t, client, server.URL, user.Email, "password123")

	resp, body := doJSONRequest(t, client, http.MethodGet, server.URL+"/notifications", nil, login.SessionToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var items []notifications.NotificationResponse
	require.NoError(t, json.Unmarshal(body, &items))
	require.NotEmpty(t, items)

	first := items[0]
	resp, _ = doJSONRequest(t, client, http.MethodPatch, server.URL+"/notifications/"+first.ID.String()+"/read", nil, login.SessionToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var stored notifications.Notification
	require.NoError(t, database.Where("id = ?", first.ID).First(&stored).Error)
	assert.True(t, stored.IsRead)
}
