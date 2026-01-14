# Installation & Configuration Guide

This document provides detailed instructions for installing, configuring, and running the E-Commerce REST API application.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start with Docker](#quick-start-with-docker)
- [Manual Installation](#manual-installation)
- [Configuration Reference](#configuration-reference)
- [Running the Application](#running-the-application)
- [Verifying the Installation](#verifying-the-installation)
- [Swagger API Documentation](#swagger-api-documentation)
- [Running Tests](#running-tests)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Software

| Software | Minimum Version | Download Link |
|----------|-----------------|---------------|
| Go | 1.21+ | [golang.org/dl](https://golang.org/dl/) |
| Docker | 20.10+ | [docker.com](https://www.docker.com/get-started) |
| Docker Compose | 2.x | Included with Docker Desktop |
| Git | 2.x | [git-scm.com](https://git-scm.com/) |

### Verify Prerequisites

```bash
# Check Go version
go version

# Check Docker version
docker --version
docker-compose --version
```

---

## Quick Start with Docker

The fastest way to get the database running is using Docker Compose.

### Step 1: Start PostgreSQL Container

```bash
# Navigate to the project directory
cd E_Commerce_Plattform

# Start PostgreSQL container
docker-compose up -d
```

This will start a PostgreSQL 16 container with the following configuration:
- **Container Name:** `ecommerce-postgres`
- **User:** `ecommerce`
- **Password:** `ecommerce`
- **Database:** `ecommerce`
- **Port:** `5432`

### Step 2: Verify Database is Running

```bash
# Check container status
docker ps

# View container logs
docker logs ecommerce-postgres
```

### Step 3: Run the Application

```bash
# Install dependencies
go mod tidy

# Run the server
go run cmd/server/main.go
```

### Stop the Database

```bash
# Stop and remove container (preserves data)
docker-compose down

# Stop and remove container AND data
docker-compose down -v
```

---

## Manual Installation

### Step 1: Clone and Setup Project

```bash
# Clone the repository
git clone <repository-url>
cd E_Commerce_Plattform

# Download Go dependencies
go mod tidy
```

### Step 2: Configure Environment

Create a `.env` file in the project root or set environment variables:

```bash
# Option 1: Create .env file
cp configs/local.env .env

# Option 2: Export environment variables (Linux/macOS)
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=ecommerce
export DB_PASSWORD=ecommerce
export DB_NAME=ecommerce
export PORT=8080

# Option 2: Set environment variables (Windows PowerShell)
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="ecommerce"
$env:DB_PASSWORD="ecommerce"
$env:DB_NAME="ecommerce"
$env:PORT="8080"
```

---

## Configuration Reference

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL server hostname |
| `DB_PORT` | `5432` | PostgreSQL server port |
| `DB_USER` | `postgres` | Database username |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `ecommerce` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode (`disable`, `require`, `verify-full`) |
| `PORT` | `8080` | HTTP server port |
| `APP_ENV` | - | Application environment (`local`, `test`, `prod`) |
| `SESSION_TTL_MINUTES` | `60` | Session token validity duration in minutes |

### Configuration Files

The project includes pre-configured environment files in the `configs/` directory:

| File | Purpose |
|------|---------|
| `configs/local.env` | Local development settings |
| `configs/test.env` | Test environment settings |
| `configs/prod.env` | Production settings (customize as needed) |

### Loading Configuration Files

```bash
# Linux/macOS - Load from config file
source configs/local.env

# Windows PowerShell - Load from config file
Get-Content configs\local.env | ForEach-Object {
    if ($_ -match '^([^=]+)=(.*)$') {
        [System.Environment]::SetEnvironmentVariable($matches[1], $matches[2])
    }
}
```

---

## Running the Application

### Development Mode

```bash
# Standard run
go run cmd/server/main.go

# With custom port
PORT=3000 go run cmd/server/main.go
```

### Building and Running Binary

```bash
# Build the application
go build -o server.exe cmd/server/main.go

# Run the binary (Windows)
.\server.exe

# Run the binary (Linux/macOS)
./server.exe
```

### Using Pre-configured Environments

```bash
# Local development
source configs/local.env && go run cmd/server/main.go

# Test environment
source configs/test.env && go run cmd/server/main.go
```

---

## Verifying the Installation

### Step 1: Health Check

Once the server is running, verify it's working:

```bash
# Using curl
curl http://localhost:8080/health

# Using PowerShell
Invoke-RestMethod -Uri "http://localhost:8080/health"
```

Expected response:
```json
{"status":"healthy"}
```

### Step 2: Test Registration

```bash
# Using curl
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'

# Using PowerShell
$body = @{
    username = "testuser"
    email = "test@example.com"
    password = "password123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/auth/register" `
  -Method POST `
  -ContentType "application/json" `
  -Body $body
```

### Step 3: Test Login

```bash
# Using curl
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Using PowerShell
$body = @{
    email = "test@example.com"
    password = "password123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/auth/login" `
  -Method POST `
  -ContentType "application/json" `
  -Body $body
```

---

## Swagger API Documentation

The API includes interactive Swagger documentation.

### Accessing Swagger UI

Once the server is running, open your browser and navigate to:

```
http://localhost:8080/swagger/index.html
```

### Swagger Features

- **Interactive API explorer** - Test endpoints directly in the browser
- **Request/Response schemas** - View expected data formats
- **Authentication** - Use the "Authorize" button to add your session token
- **Try it out** - Execute API calls and see responses

### Regenerating Swagger Docs

If you modify API endpoints, regenerate the documentation:

```bash
# Install swag CLI (if not installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
swag init -g cmd/server/main.go -o docs/swagger
```

---

## Running Tests

### Run All Tests

```bash
go test ./... -v
```

### Run Tests with Coverage

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage in terminal
go tool cover -func=coverage.out

# View coverage in browser (HTML report)
go tool cover -html=coverage.out
```

### Run Tests for Specific Package

```bash
# Test auth package
go test ./internal/auth/... -v

# Test products package
go test ./internal/products/... -v

# Test with race detection
go test ./... -race -v
```

### Test Database Setup

For running tests, create a separate test database using Docker:

```bash
# Connect to the Docker PostgreSQL container and create test database
docker exec -it ecommerce-postgres psql -U ecommerce -c "CREATE DATABASE ecommerce_test;"
```

---

## Troubleshooting

### Common Issues

#### 1. Database Connection Failed

**Error:** `failed to connect to database: connection refused`

**Solutions:**
- Verify PostgreSQL container is running: `docker ps`
- Check database credentials in environment variables
- Ensure the database port is not blocked by firewall
- Restart the container: `docker-compose restart`

```bash
# Check if PostgreSQL container is running
docker ps | grep ecommerce-postgres

# Test connection via Docker
docker exec -it ecommerce-postgres psql -U ecommerce -d ecommerce -c "SELECT 1;"
```

#### 2. Port Already in Use

**Error:** `listen tcp :8080: bind: address already in use`

**Solutions:**
- Use a different port: `PORT=3000 go run cmd/server/main.go`
- Find and stop the process using the port:

```bash
# Linux/macOS
lsof -i :8080
kill -9 <PID>

# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

#### 3. Missing Go Modules

**Error:** `cannot find package`

**Solution:**
```bash
go mod tidy
go mod download
```

#### 4. Permission Denied (Linux/macOS)

**Error:** `permission denied`

**Solution:**
```bash
chmod +x server.exe
```

#### 5. Docker Container Won't Start

**Solutions:**
```bash
# Check container logs
docker logs ecommerce-postgres

# Remove and recreate container
docker-compose down -v
docker-compose up -d
```

### Logging and Debugging

The application logs to stdout. To capture logs:

```bash
# Redirect to file
go run cmd/server/main.go > server.log 2>&1

# View logs in real-time
go run cmd/server/main.go 2>&1 | tee server.log
```

### Getting Help

If you encounter issues not covered here:

1. Check the [README.md](README.md) for additional information
2. Review the API specification in `Projektarbeit_ E-Commerce REST-API.pdf`
3. Examine the Swagger documentation at `/swagger/index.html`

---

## Next Steps

After successful installation:

1. **Explore the API** - Visit Swagger UI at `http://localhost:8080/swagger/index.html`
2. **Create an Admin User** - Register a user and manually update their role to `admin` in the database
3. **Add Products** - Use admin credentials to create products via `POST /products`
4. **Test the Flow** - Register as customer, add items to cart, and create an order

### Creating an Admin User

```bash
# Connect to the Docker PostgreSQL container
docker exec -it ecommerce-postgres psql -U ecommerce -d ecommerce

# Run the SQL command to update user role
UPDATE users SET role = 'admin' WHERE email = 'your-email@example.com';

# Exit psql
\q
```

---

*Last updated: January 2026*
