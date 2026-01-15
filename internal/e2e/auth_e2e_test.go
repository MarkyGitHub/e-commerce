package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	server "github.com/mschl/ecommerce-api/cmd/server"
	"github.com/mschl/ecommerce-api/internal/auth"
	"github.com/mschl/ecommerce-api/internal/db"
	"github.com/mschl/ecommerce-api/internal/products"
	"github.com/mschl/ecommerce-api/internal/testutil"
)

type loginResponse struct {
	SessionToken string    `json:"session_token"`
	ExpiresAt    string    `json:"expires_at"`
	User         auth.User `json:"user"`
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func newTestServer(t *testing.T) (*httptest.Server, *gorm.DB) {
	t.Helper()
	t.Setenv("APP_ENV", "test")

	database := testutil.SetupTestDB(t)
	require.NoError(t, db.Migrate(database))

	httpServer := httptest.NewServer(server.BuildMux(database))
	t.Cleanup(httpServer.Close)

	return httpServer, database
}

func doJSONRequest(t *testing.T, client *http.Client, method, url string, body any, token string) (*http.Response, []byte) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, reader)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("X-Session-Token", token)
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, respBody
}

func registerUser(t *testing.T, client *http.Client, baseURL, username, email, password string) auth.UserResponse {
	t.Helper()

	payload := auth.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	}

	resp, body := doJSONRequest(t, client, http.MethodPost, baseURL+"/auth/register", payload, "")
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var user auth.UserResponse
	require.NoError(t, json.Unmarshal(body, &user))
	require.NotEmpty(t, user.ID)
	return user
}

func loginUser(t *testing.T, client *http.Client, baseURL, email, password string) loginResponse {
	t.Helper()

	payload := auth.LoginRequest{Email: email, Password: password}

	resp, body := doJSONRequest(t, client, http.MethodPost, baseURL+"/auth/login", payload, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var login loginResponse
	require.NoError(t, json.Unmarshal(body, &login))
	require.NotEmpty(t, login.SessionToken)
	return login
}

func TestE2E_InvalidCredentialsRejected(t *testing.T) {
	server, _ := newTestServer(t)
	client := server.Client()

	registerUser(t, client, server.URL, "customer-auth-1", "customer-auth-1@example.com", "password123")

	payload := auth.LoginRequest{Email: "customer-auth-1@example.com", Password: "wrongpass"}
	resp, body := doJSONRequest(t, client, http.MethodPost, server.URL+"/auth/login", payload, "")

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var errResp errorResponse
	require.NoError(t, json.Unmarshal(body, &errResp))
	assert.NotEmpty(t, errResp.Message)
}

func TestE2E_LogoutInvalidatesSession(t *testing.T) {
	server, _ := newTestServer(t)
	client := server.Client()

	user := registerUser(t, client, server.URL, "customer-auth-2", "customer-auth-2@example.com", "password123")
	login := loginUser(t, client, server.URL, user.Email, "password123")

	resp, _ := doJSONRequest(t, client, http.MethodPost, server.URL+"/auth/logout", nil, login.SessionToken)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	resp, _ = doJSONRequest(t, client, http.MethodGet, server.URL+"/cart", nil, login.SessionToken)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestE2E_AdminEndpointAuthorization(t *testing.T) {
	server, database := newTestServer(t)
	client := server.Client()

	registerUser(t, client, server.URL, "customer-auth-3", "customer-auth-3@example.com", "password123")
	customerLogin := loginUser(t, client, server.URL, "customer-auth-3@example.com", "password123")

	createReq := products.CreateProductRequest{
		Name:        "Demo Product",
		Description: "E2E product",
		Price:       19.99,
	}

	resp, _ := doJSONRequest(t, client, http.MethodPost, server.URL+"/products", createReq, customerLogin.SessionToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	admin := registerUser(t, client, server.URL, "admin-auth-1", "admin-auth-1@example.com", "password123")
	require.NoError(t, database.Model(&auth.User{}).Where("email = ?", admin.Email).Update("role", auth.RoleAdmin).Error)

	adminLogin := loginUser(t, client, server.URL, admin.Email, "password123")
	resp, body := doJSONRequest(t, client, http.MethodPost, server.URL+"/products", createReq, adminLogin.SessionToken)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created products.Product
	require.NoError(t, json.Unmarshal(body, &created))
	assert.Equal(t, createReq.Name, created.Name)
	assert.Equal(t, createReq.Price, created.Price)
}
