package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mschl/ecommerce-api/internal/cart"
)

func TestE2E_CartRequiresAuth(t *testing.T) {
	server, _ := newTestServer(t)
	client := server.Client()

	resp, body := doJSONRequest(t, client, http.MethodGet, server.URL+"/cart", nil, "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.Unmarshal(body, &errResp))
	assert.NotEmpty(t, errResp.Message)
}

func TestE2E_CartAccessAfterLogin(t *testing.T) {
	server, _ := newTestServer(t)
	client := server.Client()

	user := registerUser(t, client, server.URL, "customer-cart-1", "customer-cart-1@example.com", "password123")
	login := loginUser(t, client, server.URL, user.Email, "password123")

	resp, body := doJSONRequest(t, client, http.MethodGet, server.URL+"/cart", nil, login.SessionToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var cartResp cart.CartResponse
	require.NoError(t, json.Unmarshal(body, &cartResp))
	assert.Equal(t, user.ID, cartResp.UserID)
}
