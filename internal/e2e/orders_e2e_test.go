package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type paginatedResponse struct {
	Items      json.RawMessage `json:"items"`
	TotalCount int64           `json:"total_count"`
}

func TestE2E_OrdersRequiresAuth(t *testing.T) {
	server, _ := newTestServer(t)
	client := server.Client()

	resp, _ := doJSONRequest(t, client, http.MethodGet, server.URL+"/orders", nil, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestE2E_OrdersListWithAuth(t *testing.T) {
	server, _ := newTestServer(t)
	client := server.Client()

	user := registerUser(t, client, server.URL, "customer-orders-1", "customer-orders-1@example.com", "password123")
	login := loginUser(t, client, server.URL, user.Email, "password123")

	resp, body := doJSONRequest(t, client, http.MethodGet, server.URL+"/orders", nil, login.SessionToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var response paginatedResponse
	require.NoError(t, json.Unmarshal(body, &response))
	require.NotNil(t, response.Items)
	assert.GreaterOrEqual(t, response.TotalCount, int64(0))
}
