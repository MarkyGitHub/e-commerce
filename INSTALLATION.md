# Installations- & Konfigurationsleitfaden

Dieses Dokument enthält Anleitungen zum Installieren, Konfigurieren und Ausführen der E-Commerce REST API.

## Inhaltsverzeichnis

- [Voraussetzungen](#voraussetzungen)
- [Schnellstart mit Docker](#schnellstart-mit-docker)
- [Konfigurationsreferenz](#konfigurationsreferenz)
- [Anwendung ausführen](#anwendung-ausführen)
- [Installation überprüfen](#installation-überprüfen)
- [Swagger API-Dokumentation](#swagger-api-dokumentation)
- [Tests ausführen](#tests-ausführen)
- [Fehlerbehebung](#fehlerbehebung)

---

## Voraussetzungen

### Benötigte Software

| Software | Mindestversion | Download-Link |
|----------|-----------------|---------------|
| Go | 1.21+ | [golang.org/dl](https://golang.org/dl/) |
| Docker | 20.10+ | [docker.com](https://www.docker.com/get-started) |
| Docker Compose | 2.x | In Docker Desktop enthalten |

### Voraussetzungen prüfen

```bash
# Check version
go version

# Check version
docker --version
docker-compose --version
```

---

## Schnellstart mit Docker

Der schnellste Weg, die Datenbank zu starten, ist Docker Compose.

### Schritt 1: PostgreSQL-Container starten

```bash
# Navigate to the project
cd E_Commerce_Plattform

# Start docker container
docker-compose up -d
```

Dadurch wird ein PostgreSQL Container mit folgender Konfiguration gestartet:
- **Container-Name:** `e-commerce-db`
- **User:** `postgres`
- **Passwort:** `example`
- **Datenbank:** `postgres`
- **Port:** `5434` (extern) → `5432` (intern)
- **Adminer (DB-Verwaltung):** `http://localhost:8083`

### Schritt 2: Prüfen, ob die Datenbank läuft

```bash
# Check status
docker ps

# View logs
docker logs e-commerce-db
```

### Schritt 3: Anwendung starten

```bash
# Install dependencies
go mod tidy

# Run the server
go run cmd/server/main.go
```

### Datenbank stoppen

```bash
# Stop and remove container (preserves data)
docker-compose down

# Stop and remove container AND data
docker-compose down -v
```

---

## Konfigurationsreferenz

### Umgebungsvariablen

| Variable | Default | Beschreibung |
|----------|---------|--------------|
| `APP_ENV` | `local` | Anwendungskontext (`local`, `test`, `prod`) |
| `DB_HOST` | `localhost` | Hostname des PostgreSQL-Servers |
| `DB_PORT` | `5434` | Port des PostgreSQL-Servers |
| `DB_USER` | `postgres` | Datenbank-Benutzername |
| `DB_PASSWORD` | `example` | Datenbank-Passwort |
| `DB_NAME` | `postgres` | Datenbankname |
| `DB_SSLMODE` | `disable` | SSL-Modus (`disable`, `require`, `verify-full`) |
| `SERVER_PORT` | `8080` | HTTP-Server-Port |
| `SESSION_TTL_MINUTES` | `60` | Gültigkeitsdauer des Session-Tokens in Minuten |

### Konfigurationsdateien

Im Verzeichnis `configs/` liegen vorkonfigurierte Environment-Dateien:

| Datei | Zweck |
|------|---------|
| `configs/local.env` | Einstellungen für lokale Entwicklung |
| `configs/test.env` | Einstellungen für Testumgebung |
| `configs/prod.env` | Produktionseinstellungen (bei Bedarf anpassen) |

### Konfigurationsdateien laden

```bash
# Linux/macOS (bash/zsh) - Config-Datei laden
source configs/local.env

# Windows PowerShell - Config-Datei laden
Get-Content configs\local.env | ForEach-Object {
    if ($_ -match '^([^=]+)=(.*)$') {
        [System.Environment]::SetEnvironmentVariable($matches[1], $matches[2])
    }
}

# Windows CMD - Config-Datei laden (aktuelle cmd-Session)
for /f "usebackq tokens=1,* delims==" %A in ("configs\local.env") do set %A=%B
```

---

## Anwendung ausführen

### Entwicklungsmodus

```bash
# run
go run cmd/server/main.go

# With custom port (Linux/macOS)
PORT=3000 go run cmd/server/main.go

# With custom port (Windows PowerShell)
$env:PORT="3000"; go run cmd/server/main.go

# With custom port (Windows CMD)
set PORT=3000 && go run cmd/server/main.go
```

### Binary bauen und ausführen

```bash
# Build the application
go build -o server.exe cmd/server/main.go

# Run the binary (Windows)
.\server.exe

# Run the binary (Linux/macOS)
./server.exe
```

### Vorkonfigurierte Umgebungen verwenden

```bash
# Local environment (Linux/macOS)
source configs/local.env && go run cmd/server/main.go

# Test environment (Linux/macOS)
source configs/test.env && go run cmd/server/main.go

# Windows PowerShell (local)
Get-Content configs\local.env | ForEach-Object { if ($_ -match '^([^=]+)=(.*)$') { [System.Environment]::SetEnvironmentVariable($matches[1], $matches[2]) } }; go run cmd/server/main.go

# Windows CMD (local)
for /f "usebackq tokens=1,* delims==" %A in ("configs\local.env") do set %A=%B
go run cmd/server/main.go
```

---

## Installation überprüfen

### Schritt 1: Health-Check

Wenn der Server läuft, kannst geprüft werden, ob alles funktioniert:

```bash
# Using curl
curl http://localhost:8080/health

# Using PowerShell
Invoke-RestMethod -Uri "http://localhost:8080/health"
```

Erwartete Antwort:
```json
{"status":"healthy"}
```

### Schritt 2: Registrierung testen

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

### Schritt 3: Login testen

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

## Swagger API-Dokumentation

Die API enthält eine interaktive Swagger-Dokumentation.

### Swagger UI öffnen

Wenn der Server läuft, öffne im Browser:

```
http://localhost:8080/swagger/index.html
```

### Swagger-Funktionen

- **Interaktiver API-Explorer** – Endpunkte direkt im Browser testen
- **Request/Response-Schemas** – Erwartete Datenformate einsehen
- **Authentifizierung** – Über "Authorize" den Session-Token setzen
- **Try it out** – API-Calls ausführen und Responses ansehen

### Swagger-Dokumentation neu generieren

```bash
# Install swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
swag init -g cmd/server/main.go -o docs/swagger
```

---

## Tests ausführen

### Alle Tests ausführen

```bash
go test ./... -v
```

E2E-Tests ausführen:
```bash
go test .\internal\e2e\...
```

E2E-Tests ausführen (mit Test-Environment):
```bash
set APP_ENV=test & go test .\internal\e2e\...
set APP_ENV=test & go test -v .\internal\e2e\...
```

### Tests mit Coverage ausführen

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage in terminal
go tool cover -func=coverage.out

# View coverage in browser (HTML report)
go tool cover -html=coverage.out
```

### Tests für ein bestimmtes Package ausführen

```bash
# Test auth package
go test ./internal/auth/... -v

# Test products package
go test ./internal/products/... -v

# Test with race detection
go test ./... -race -v
```

### Testdatenbank einrichten

Für Tests empfiehlt sich eine separate Testdatenbank via Docker:

```bash
# Connect to the Docker container and create test database
docker exec -it e-commerce-db psql -U postgres -c "CREATE DATABASE postgres_test;"
```

---

## Fehlerbehebung

### Häufige Probleme

#### 1. Datenbankverbindung fehlgeschlagen

**Fehler:** `failed to connect to database: connection refused`

**Lösungen:**
- Prüfe, ob der PostgreSQL-Container läuft: `docker ps`
- Prüfe die Datenbank-Credentials in den Umgebungsvariablen
- Stelle sicher, dass der DB-Port nicht durch eine Firewall blockiert wird
- Container neu starten: `docker-compose restart`

```bash
# Check if PostgreSQL is running
docker ps | grep e-commerce-db

# Test connection via Docker
docker exec -it e-commerce-db psql -U postgres -c "SELECT 1;"

# Open Adminer (browser-based DB admin)
# Navigate to http://localhost:8083
# System: PostgreSQL, Server: db, User: postgres, Password: example
```

#### 2. Port bereits belegt

**Fehler:** `listen tcp :8080: bind: address already in use`

**Lösungen:**
- Anderen Port verwenden: `PORT=3000 go run cmd/server/main.go`
- Prozess finden und stoppen, der den Port belegt:

```bash
# Linux/macOS
lsof -i :8080
kill -9 <PID>

# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

#### 3. Fehlende Go-Module

**Fehler:** `cannot find package`

**Lösung:**
```bash
go mod tidy
go mod download
```

#### 4. Permission Denied (Linux/macOS)

**Fehler:** `permission denied`

**Lösung:**
```bash
chmod +x server.exe
```

#### 5. Docker-Container startet nicht

**Lösungen:**
```bash
# Check container logs
docker logs e-commerce-db

# Remove and recreate container
docker-compose down -v
docker-compose up -d
```

### Logging und Debugging

Die Anwendung loggt nach stdout. Um Logs zu speichern:

```bash
# Redirect to file
go run cmd/server/main.go > server.log 2>&1

# View logs in real-time
go run cmd/server/main.go 2>&1 | tee server.log
```

### Hilfe bekommen

Wenn du auf Probleme stößt, die hier nicht beschrieben sind:

1. Sieh in der [README.md](README.md) nach weiteren Informationen
2. Prüfe die API-Spezifikation in `Projektarbeit_ E-Commerce REST-API.pdf`
3. Öffne die Swagger-Dokumentation unter `/swagger/index.html`

---

