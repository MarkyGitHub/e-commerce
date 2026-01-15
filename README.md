# E-Commerce REST API

Eine vollständige E-Commerce REST API, gebaut in Go und umgesetzt nach einem Layered-Architecture-Pattern (Handler → Service → Repository).

## Features

- **Benutzerauthentifizierung**: Registrierung, Login, Logout mit sitzungsbasierter Authentifizierung
- **Produktverwaltung**: CRUD-Operationen für Produkte (nur Admin: erstellen/aktualisieren/löschen)
- **Bestandsverwaltung (Inventory)**: Lagerbestände verfolgen und aktualisieren (nur Admin)
- **Warenkorb**: Artikel hinzufügen, aktualisieren, entfernen – mit Bestandsvalidierung
- **Bestellabwicklung**: Bestellungen aus dem Warenkorb erstellen, inkl. transaktionalem Bestandsabbau
- **Benachrichtigungen**: Bestellbenachrichtigungen für Kund:innen

## Architektur

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

## Projektstruktur

```
├── cmd/
│   └── server/
│       └── main.go           # Anwendungseinstiegspunkt
├── configs/
│   ├── local.env             # Einstellungen für lokale Entwicklung
│   ├── test.env              # Einstellungen für Testumgebung
│   └── prod.env              # Produktionseinstellungen
├── internal/
│   ├── auth/                 # Authentifizierungsmodul
│   ├── cart/                 # Warenkorb-Modul
│   ├── config/               # Zentrale Konfiguration
│   ├── db/                   # Datenbankverbindung und Migrationen
│   ├── inventory/            # Bestandsverwaltungsmodul (Stocks)
│   ├── middleware/           # HTTP-Middleware (Auth, Rollen)
│   ├── notifications/        # Benachrichtigungsmodul
│   ├── orders/               # Bestellabwicklungsmodul
│   ├── products/             # Produktkatalog-Modul
│   └── shared/               # Gemeinsame Utilities
├── docs/
│   └── swagger/              # Swagger-Dokumentation
├── go.mod
├── go.sum
└── readme.md
└── installation.md

```

## Voraussetzungen

- Go 1.21 oder höher
- PostgreSQL 14 oder höher

## Konfiguration

Die Anwendung verwendet Umgebungsvariablen zur Konfiguration. Konfigurationsdateien befinden sich im Verzeichnis `configs/`.

| Variable | Default | Beschreibung |
|----------|---------|--------------|
| `APP_ENV` | local | Umgebung (`local`, `test`, `prod`) |
| `DB_HOST` | localhost | PostgreSQL-Host |
| `DB_PORT` | 5434 | PostgreSQL-Port |
| `DB_USER` | postgres | PostgreSQL-Benutzername |
| `DB_PASSWORD` | example | PostgreSQL-Passwort |
| `DB_NAME` | postgres | Datenbankname |
| `DB_SSLMODE` | disable | SSL-Modus (disable/require) |
| `SERVER_PORT` | 8080 | Server-Port |
| `SESSION_TTL_MINUTES` | 60 | Gültigkeitsdauer des Session-Tokens in Minuten |

## Schnellstart

1. **Projekt öffnen**:
   ```bash
   cd E_Commerce_Plattform
   ```

2. **Docker starten**:
   ```bash
   ocker-compose up -d
   ```

3. **Abhängigkeiten installieren**:
   ```bash
   go mod tidy
   ```

4. **Server starten**:
   ```bash
   go run cmd/server/main.go
   ```

5. **Die API ist nun erreichbar unter** `http://localhost:8080`

## API-Endpunkte

### Authentifizierung
| Method | Endpoint | Beschreibung | Auth |
|--------|----------|--------------|------|
| POST | `/auth/register` | Neuen Benutzer registrieren | Public |
| POST | `/auth/login` | Login und Session-Token erhalten | Public |
| POST | `/auth/logout` | Logout und Session ungültig machen | Session |

### Produkte
| Method | Endpoint | Beschreibung | Auth |
|--------|----------|--------------|------|
| GET | `/products` | Alle Produkte auflisten (paginiert) | Public |
| GET | `/products/{id}` | Produkt per ID abrufen | Public |
| POST | `/products` | Produkt erstellen | Admin |
| PUT | `/products/{id}` | Produkt aktualisieren | Admin |
| DELETE | `/products/{id}` | Produkt löschen | Admin |

### Inventory (Bestand)
| Method | Endpoint | Beschreibung | Auth |
|--------|----------|--------------|------|
| GET | `/inventory/{product_id}` | Bestand für ein Produkt abrufen | Public |
| PUT | `/inventory/{product_id}` | Bestandsmenge aktualisieren | Admin |

### Warenkorb
| Method | Endpoint | Beschreibung | Auth |
|--------|----------|--------------|------|
| GET | `/cart` | Warenkorb des aktuellen Benutzers abrufen | Customer |
| POST | `/cart/items` | Artikel zum Warenkorb hinzufügen | Customer |
| PUT | `/cart/items/{item_id}` | Menge eines Warenkorb-Artikels aktualisieren | Customer |
| DELETE | `/cart/items/{item_id}` | Artikel aus dem Warenkorb entfernen | Customer |
| DELETE | `/cart` | Gesamten Warenkorb leeren | Customer |

### Bestellungen
| Method | Endpoint | Beschreibung | Auth |
|--------|----------|--------------|------|
| POST | `/orders` | Bestellung aus dem Warenkorb erstellen | Customer |
| GET | `/orders` | Bestellungen auflisten (eigene oder alle für Admin) | Customer/Admin |
| GET | `/orders/{id}` | Bestelldetails abrufen | Customer/Admin |

### Benachrichtigungen
| Method | Endpoint | Beschreibung | Auth |
|--------|----------|--------------|------|
| GET | `/notifications` | Benachrichtigungen des Benutzers auflisten | Customer |
| PUT | `/notifications/{id}/read` | Benachrichtigung als gelesen markieren | Customer |

### Utility
| Method | Endpoint | Beschreibung | Auth |
|--------|----------|--------------|------|
| GET | `/health` | Health-Check | Public |

## Authentifizierung

Die API verwendet sitzungsbasierte Authentifizierung. Nach dem Login muss der Session-Token bei allen authentifizierten Requests mitgesendet werden:

```
X-Session-Token: <your-session-token>
```

## Beispielnutzung

### Neuen Benutzer registrieren
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

### Produkte auflisten
```bash
curl http://localhost:8080/products
```

### Artikel zum Warenkorb hinzufügen
```bash
curl -X POST http://localhost:8080/cart/items \
  -H "Content-Type: application/json" \
  -H "X-Session-Token: <your-token>" \
  -d '{"product_id":"<product-uuid>","quantity":2}'
```

### Bestellung erstellen
```bash
curl -X POST http://localhost:8080/orders \
  -H "X-Session-Token: <your-token>"
```

## Tests

Alle Tests ausführen:
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

Tests mit Coverage ausführen:
```bash
go test ./... -cover
```

Coverage-Report erzeugen:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Fehlerantworten

Alle Fehler folgen einem konsistenten Format:
```json
{
  "error": "HTTP Status Text",
  "message": "Detaillierte Fehlermeldung"
}
```

### HTTP-Statuscodes
| Code | Bedeutung |
|------|-----------|
| 200 | Erfolgreich |
| 201 | Erstellt |
| 400 | Bad Request (Validierungsfehler) |
| 401 | Unauthorized (fehlender/ungültiger Token) |
| 403 | Forbidden (unzureichende Berechtigungen) |
| 404 | Not Found |
| 409 | Conflict (doppelte Ressource) |
| 500 | Internal Server Error |

## Lizenz

Dieses Projekt ist Teil der Kurs-Projektarbeit, Programmieren mit Go.
