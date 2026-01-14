# E-Commerce REST API

A complete E-Commerce REST API built in Go following a layered architecture pattern (Handler → Service → Repository).

## Features

- **User Authentication**: Registration, login, logout with session-based authentication
- **Product Management**: CRUD operations for products (admin only for create/update/delete)
- **Inventory Management**: Stock tracking and updates (admin only)
- **Shopping Cart**: Add, update, remove items with inventory validation
- **Order Processing**: Create orders from cart with transactional inventory decrement
- **Notifications**: Order notifications for customers

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Go HTTP Server (net/http)                   │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Middleware Chain                      │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │    │
│  │  │   Logging    │→ │  Auth Check  │→ │ Role Check   │   │    │
│  │  └──────────────┘  └──────────────┘  └──────────────┘   │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                │                                 │
│                                ▼                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                      Handlers Layer                      │    │
│  │  ┌────────┐ ┌──────────┐ ┌───────────┐ ┌──────┐ ┌─────┐│    │
│  │  │  Auth  │ │ Products │ │ Inventory │ │ Cart │ │Order││    │
│  │  └────────┘ └──────────┘ └───────────┘ └──────┘ └─────┘│    │
│  └─────────────────────────────────────────────────────────┘    │
│                                │                                 │
│                                ▼                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                      Services Layer                      │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                │                                 │
│                                ▼                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Repositories Layer                    │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     PostgreSQL Database                          │
└─────────────────────────────────────────────────────────────────┘
```

## Project Structure

```
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── auth/                 # Authentication module
│   ├── cart/                 # Shopping cart module
│   ├── db/                   # Database connection and migrations
│   ├── inventory/            # Inventory management module
│   ├── middleware/           # HTTP middleware (auth, roles)
│   ├── notifications/        # Notifications module
│   ├── orders/               # Order processing module
│   ├── products/             # Product catalog module
│   └── shared/               # Shared utilities
├── docs/
│   └── swagger/              # Swagger documentation
├── go.mod
├── go.sum
└── README.md
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 14 or higher

## Configuration

The application uses environment variables for configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_USER` | postgres | PostgreSQL username |
| `DB_PASSWORD` | postgres | PostgreSQL password |
| `DB_NAME` | ecommerce | Database name |
| `DB_SSLMODE` | disable | SSL mode (disable/require) |
| `PORT` | 8080 | Server port |

## Getting Started

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd E_Commerce_Plattform
   ```

2. **Set up PostgreSQL database**:
   ```bash
   createdb ecommerce
   ```

3. **Install dependencies**:
   ```bash
   go mod tidy
   ```

4. **Run the server**:
   ```bash
   go run cmd/server/main.go
   ```

5. **The API is now available at** `http://localhost:8080`

## API Endpoints

### Authentication
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/auth/register` | Register new user | Public |
| POST | `/auth/login` | Login and get session token | Public |
| POST | `/auth/logout` | Logout and invalidate session | Session |

### Products
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/products` | List all products (paginated) | Public |
| GET | `/products/{id}` | Get product by ID | Public |
| POST | `/products` | Create product | Admin |
| PUT | `/products/{id}` | Update product | Admin |
| DELETE | `/products/{id}` | Delete product | Admin |

### Inventory
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/inventory/{product_id}` | Get inventory for product | Public |
| PUT | `/inventory/{product_id}` | Update inventory quantity | Admin |

### Cart
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/cart` | Get current user's cart | Customer |
| POST | `/cart/items` | Add item to cart | Customer |
| PUT | `/cart/items/{item_id}` | Update cart item quantity | Customer |
| DELETE | `/cart/items/{item_id}` | Remove item from cart | Customer |
| DELETE | `/cart` | Clear entire cart | Customer |

### Orders
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/orders` | Create order from cart | Customer |
| GET | `/orders` | List orders (user's own or all for admin) | Customer/Admin |
| GET | `/orders/{id}` | Get order details | Customer/Admin |

### Notifications
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/notifications` | List user's notifications | Customer |
| PUT | `/notifications/{id}/read` | Mark notification as read | Customer |

### Utility
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/health` | Health check | Public |

## Authentication

The API uses session-based authentication. After logging in, include the session token in all authenticated requests:

```
X-Session-Token: <your-session-token>
```

## Example Usage

### Register a new user
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","name":"John Doe"}'
```

### Login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### List products
```bash
curl http://localhost:8080/products
```

### Add item to cart
```bash
curl -X POST http://localhost:8080/cart/items \
  -H "Content-Type: application/json" \
  -H "X-Session-Token: <your-token>" \
  -d '{"product_id":"<product-uuid>","quantity":2}'
```

### Create order
```bash
curl -X POST http://localhost:8080/orders \
  -H "X-Session-Token: <your-token>"
```

## Testing

Run all tests:
```bash
go test ./... -v
```

Run tests with coverage:
```bash
go test ./... -cover
```

Generate coverage report:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Error Responses

All errors follow a consistent format:
```json
{
  "error": "HTTP Status Text",
  "message": "Detailed error message"
}
```

### HTTP Status Codes
| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request (validation error) |
| 401 | Unauthorized (missing/invalid token) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Not Found |
| 409 | Conflict (duplicate resource) |
| 500 | Internal Server Error |

## License

This project is part of a coursework assignment.
