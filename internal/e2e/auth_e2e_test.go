package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/MarkyGitHub/e-commerce/internal/auth"
	"github.com/MarkyGitHub/e-commerce/internal/db"
	"github.com/MarkyGitHub/e-commerce/internal/inventory"
	"github.com/MarkyGitHub/e-commerce/internal/products"
	"github.com/MarkyGitHub/e-commerce/internal/server"
	"github.com/MarkyGitHub/e-commerce/internal/testutil"
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

func createAdminLogin(t *testing.T, client *http.Client, baseURL string, database *gorm.DB, username, email, password string) loginResponse {
	t.Helper()

	admin := registerUser(t, client, baseURL, username, email, password)
	require.NoError(t, database.Model(&auth.User{}).Where("email = ?", admin.Email).Update("role", auth.RoleAdmin).Error)

	return loginUser(t, client, baseURL, admin.Email, password)
}

func createProduct(t *testing.T, client *http.Client, baseURL, token, name string, price float64) products.Product {
	t.Helper()

	createReq := products.CreateProductRequest{
		Name:        name,
		Description: "E2E product",
		Price:       price,
	}

	resp, body := doJSONRequest(t, client, http.MethodPost, baseURL+"/products", createReq, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created products.Product
	require.NoError(t, json.Unmarshal(body, &created))
	return created
}

func seedInventory(t *testing.T, database *gorm.DB, productID string, quantity int, location string) inventory.Inventory {
	t.Helper()

	inv := inventory.Inventory{
		ID:        uuid.New(),
		ProductID: mustParseUUID(t, productID),
		Quantity:  quantity,
		Location:  location,
	}
	require.NoError(t, database.Create(&inv).Error)
	return inv
}

func mustParseUUID(t *testing.T, id string) uuid.UUID {
	t.Helper()
	parsed, err := uuid.Parse(id)
	require.NoError(t, err)
	return parsed
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

func TestE2E_RegisterCreatesUserWithHashedPassword(t *testing.T) {
	server, database := newTestServer(t)
	client := server.Client()

	user := registerUser(t, client, server.URL, "max_mustermann", "max@example.com", "sicheres_passwort_123")
	assert.Equal(t, "max_mustermann", user.Username)
	assert.Equal(t, "max@example.com", user.Email)

	var stored auth.User
	require.NoError(t, database.Where("email = ?", "max@example.com").First(&stored).Error)
	require.NotEmpty(t, stored.PasswordHash)
	require.NotEqual(t, "sicheres_passwort_123", stored.PasswordHash)

	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("sicheres_passwort_123")))
}

func TestE2E_LoginCreatesSession(t *testing.T) {
	server, database := newTestServer(t)
	client := server.Client()

	registerUser(t, client, server.URL, "login-user", "login-user@example.com", "password123")
	login := loginUser(t, client, server.URL, "login-user@example.com", "password123")

	assert.Equal(t, auth.RoleCustomer, login.User.Role)
	require.NotEmpty(t, login.SessionToken)

	var session auth.Session
	require.NoError(t, database.Where("token = ?", login.SessionToken).First(&session).Error)
	assert.Equal(t, login.User.ID, session.UserID)
}

func TestE2E_CreateProductAsAdminPersists(t *testing.T) {
	server, database := newTestServer(t)
	client := server.Client()

	adminLogin := createAdminLogin(t, client, server.URL, database, "admin-product-1", "admin-product-1@example.com", "password123")
	created := createProduct(t, client, server.URL, adminLogin.SessionToken, "New Product", 89.90)

	var stored products.Product
	require.NoError(t, database.Where("id = ?", created.ID).First(&stored).Error)
	assert.Equal(t, created.Name, stored.Name)
	assert.Equal(t, created.Price, stored.Price)
}

func TestE2E_GetProductsPagination(t *testing.T) {
	server, database := newTestServer(t)
	client := server.Client()

	adminLogin := createAdminLogin(t, client, server.URL, database, "admin-product-2", "admin-product-2@example.com", "password123")
	createProduct(t, client, server.URL, adminLogin.SessionToken, "Product A", 10.00)
	createProduct(t, client, server.URL, adminLogin.SessionToken, "Product B", 20.00)

	resp, body := doJSONRequest(t, client, http.MethodGet, server.URL+"/products?page=1&limit=1", nil, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Items      []products.Product `json:"items"`
		TotalCount int64              `json:"total_count"`
	}
	require.NoError(t, json.Unmarshal(body, &response))
	require.Len(t, response.Items, 1)
	assert.GreaterOrEqual(t, response.TotalCount, int64(2))
}

func TestE2E_LogoutInvalidatesSession(t *testing.T) {
	server, database := newTestServer(t)
	client := server.Client()

	user := registerUser(t, client, server.URL, "customer-auth-2", "customer-auth-2@example.com", "password123")
	login := loginUser(t, client, server.URL, user.Email, "password123")

	resp, _ := doJSONRequest(t, client, http.MethodPost, server.URL+"/auth/logout", nil, login.SessionToken)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	var sessionCount int64
	require.NoError(t, database.Model(&auth.Session{}).Where("token = ?", login.SessionToken).Count(&sessionCount).Error)
	assert.Equal(t, int64(0), sessionCount)

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

	adminLogin := createAdminLogin(t, client, server.URL, database, "admin-auth-1", "admin-auth-1@example.com", "password123")
	resp, body := doJSONRequest(t, client, http.MethodPost, server.URL+"/products", createReq, adminLogin.SessionToken)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created products.Product
	require.NoError(t, json.Unmarshal(body, &created))
	assert.Equal(t, createReq.Name, created.Name)
	assert.Equal(t, createReq.Price, created.Price)
}
