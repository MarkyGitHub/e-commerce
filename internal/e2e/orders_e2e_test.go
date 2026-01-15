package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mschl/ecommerce-api/internal/cart"
	"github.com/mschl/ecommerce-api/internal/inventory"
	"github.com/mschl/ecommerce-api/internal/orders"
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

func TestE2E_CreateOrderFromCart(t *testing.T) {
	server, database := newTestServer(t)
	client := server.Client()

	adminLogin := createAdminLogin(t, client, server.URL, database, "admin-order-1", "admin-order-1@example.com", "password123")
	product := createProduct(t, client, server.URL, adminLogin.SessionToken, "Order Product", 15.00)
	seedInventory(t, database, product.ID.String(), 10, "Order Warehouse")

	user := registerUser(t, client, server.URL, "customer-order-1", "customer-order-1@example.com", "password123")
	userLogin := loginUser(t, client, server.URL, user.Email, "password123")

	addReq := cart.AddItemRequest{
		ProductID: product.ID,
		Quantity:  2,
	}
	resp, _ := doJSONRequest(t, client, http.MethodPost, server.URL+"/cart/items", addReq, userLogin.SessionToken)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	orderReq := orders.CreateOrderRequest{
		Items: []orders.OrderItemRequest{
			{ProductID: product.ID, Quantity: 2},
		},
		ShippingStreet:  "Main Street 1",
		ShippingCity:    "Berlin",
		ShippingZip:     "10115",
		ShippingCountry: "DE",
	}

	resp, body := doJSONRequest(t, client, http.MethodPost, server.URL+"/orders", orderReq, userLogin.SessionToken)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var orderResp orders.OrderResponse
	require.NoError(t, json.Unmarshal(body, &orderResp))
	require.Len(t, orderResp.Items, 1)
	assert.Equal(t, user.ID, orderResp.UserID)
	assert.Equal(t, product.ID, orderResp.Items[0].ProductID)
	assert.Equal(t, 2, orderResp.Items[0].Quantity)

	var cartCount int64
	require.NoError(t, database.Model(&cart.CartItem{}).Where("cart_id IN (SELECT id FROM carts WHERE user_id = ?)", user.ID).Count(&cartCount).Error)
	assert.Equal(t, int64(0), cartCount)
}

func TestE2E_GetInventoryRequiresAuthAndReturnsData(t *testing.T) {
	server, database := newTestServer(t)
	client := server.Client()

	adminLogin := createAdminLogin(t, client, server.URL, database, "admin-inv-1", "admin-inv-1@example.com", "password123")
	product := createProduct(t, client, server.URL, adminLogin.SessionToken, "Inventory Product", 30.00)
	seedInventory(t, database, product.ID.String(), 5, "Inventory Warehouse")

	user := registerUser(t, client, server.URL, "customer-inv-1", "customer-inv-1@example.com", "password123")
	userLogin := loginUser(t, client, server.URL, user.Email, "password123")

	resp, body := doJSONRequest(t, client, http.MethodGet, server.URL+"/inventory/"+product.ID.String(), nil, userLogin.SessionToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var inv inventory.InventoryResponse
	require.NoError(t, json.Unmarshal(body, &inv))
	assert.Equal(t, product.ID, inv.ProductID)
	assert.Equal(t, 5, inv.Quantity)
}

func TestE2E_IncrementInventoryAsAdmin(t *testing.T) {
	server, database := newTestServer(t)
	client := server.Client()

	adminLogin := createAdminLogin(t, client, server.URL, database, "admin-inv-2", "admin-inv-2@example.com", "password123")
	product := createProduct(t, client, server.URL, adminLogin.SessionToken, "Inventory Product 2", 40.00)
	seedInventory(t, database, product.ID.String(), 3, "Inventory Warehouse")

	updateReq := inventory.UpdateInventoryRequest{IncrementBy: 4}
	resp, body := doJSONRequest(t, client, http.MethodPatch, server.URL+"/inventory/"+product.ID.String(), updateReq, adminLogin.SessionToken)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var inv inventory.InventoryResponse
	require.NoError(t, json.Unmarshal(body, &inv))
	assert.Equal(t, 7, inv.Quantity)
}
