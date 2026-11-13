package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MarkyGitHub/e-commerce/internal/auth"
	"github.com/MarkyGitHub/e-commerce/internal/cart"
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

func TestE2E_AddItemToCart(t *testing.T) {
	server, database := newTestServer(t)
	client := server.Client()

	adminLogin := createAdminLogin(t, client, server.URL, database, "admin-cart-1", "admin-cart-1@example.com", "password123")
	product := createProduct(t, client, server.URL, adminLogin.SessionToken, "Cart Item Product", 24.50)
	seedInventory(t, database, product.ID.String(), 10, "Main Warehouse")

	customer := registerUser(t, client, server.URL, "customer-cart-2", "customer-cart-2@example.com", "password123")
	customerLogin := loginUser(t, client, server.URL, customer.Email, "password123")

	addReq := cart.AddItemRequest{
		ProductID: product.ID,
		Quantity:  2,
	}

	resp, body := doJSONRequest(t, client, http.MethodPost, server.URL+"/cart/items", addReq, customerLogin.SessionToken)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var item cart.CartItemResponse
	require.NoError(t, json.Unmarshal(body, &item))
	assert.Equal(t, addReq.ProductID, item.ProductID)
	assert.Equal(t, addReq.Quantity, item.Quantity)

	var storedCart cart.Cart
	require.NoError(t, database.Where("user_id = ?", customer.ID).First(&storedCart).Error)

	var storedItem cart.CartItem
	require.NoError(t, database.Where("cart_id = ?", storedCart.ID).First(&storedItem).Error)
	assert.Equal(t, addReq.ProductID, storedItem.ProductID)
	assert.Equal(t, addReq.Quantity, storedItem.Quantity)

	var adminUser auth.User
	require.NoError(t, database.Where("email = ?", "admin-cart-1@example.com").First(&adminUser).Error)
	assert.Equal(t, auth.RoleAdmin, adminUser.Role)
}
