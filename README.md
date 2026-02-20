# GoApp

A modern web application built with Go, Templ, HTMX, and Tailwind CSS featuring server-side rendering and RESTful API.

## Tech Stack

* **Backend**: Go 1.25.x with Gin framework
* **Templates**: Templ (type-safe HTML templates)
* **Styling**: Tailwind CSS v4
* **Interactivity**: HTMX
* **Database**: PostgreSQL
* **Hot Reload**: Air

## Prerequisites

* Go 1.25.x or higher
* Node.js 24.x or higher
* Docker & Docker Compose
* Make

## Quick Start

```bash
# Clone repository
git clone https://github.com/PhuNguyenPT/GoApp.git
cd GoApp

# Install dependencies
go mod download

# Setup environment variables
cp .env.example .env
# Edit .env with your configuration

# Run database + app in dev mode (Docker)
make docker-watch

# Or run with hot reload locally
make watch
```

Visit http://localhost:8080

## Makefile Commands

### Build

| Command | Description |
|---|---|
| `make all` | Build and test |
| `make build` | Build application binary |
| `make templ-generate` | Generate templ files |
| `make sqlc-generate` | Generate sqlc database files |
| `make tailwind-build` | Build Tailwind CSS |

### Development

| Command | Description |
|---|---|
| `make watch` | Hot reload with Air (recommended) |
| `make run` | Run SSR server + SPA frontend |

### Docker

| Command | Description |
|---|---|
| `make docker-watch` | Start dev environment with hot reload |
| `make docker-watch-down` | Stop dev environment |
| `make docker-prod` | Start production environment |
| `make docker-prod-down` | Stop production environment |

### Database Migrations

| Command | Description |
|---|---|
| `make migrate-up` | Run pending migrations |
| `make migrate-down` | Roll back last migration |

### Testing

| Command | Description |
|---|---|
| `make test` | Run all tests |
| `make itest` | Run integration tests |

### Code Quality

| Command | Description |
|---|---|
| `make lint` | Run linter |
| `make lint-fix` | Run linter with auto-fix |
| `make vet` | Run static analysis |
| `make fmt` | Format code |

### Cleanup

| Command | Description |
|---|---|
| `make clean` | Remove binary and generated templ files |

## Project Structure

```
GoApp/
├── cmd/api/
│   └── main.go                     # Application entry point
├── internal/
│   ├── server/
│   │   ├── server.go               # HTTP server setup
│   │   ├── routes.go               # Route definitions
│   │   ├── middleware.go           # Auth & session middleware
│   │   ├── auth.go                 # Login, register, logout handlers
│   │   ├── pages.go                # Page handlers (home, contact, dashboard)
│   │   ├── api.go                  # REST API handlers
│   │   ├── config.go               # Server configuration
│   │   └── session_cleanup.go      # Background session cleanup
│   ├── views/
│   │   ├── layout.templ            # Base layout
│   │   ├── home.templ              # Home page
│   │   ├── contact.templ           # Contact page
│   │   ├── login.templ             # Login page
│   │   ├── register.templ          # Register page
│   │   ├── dashboard.templ         # Dashboard page
│   │   └── seo.go                  # SEO helpers
│   └── database/
│       ├── database.go             # DB connection
│       ├── db.go                   # sqlc DB interface
│       ├── models.go               # Generated models
│       ├── users.sql.go            # Generated user queries
│       ├── sessions.sql.go         # Generated session queries
│       ├── queries/
│       │   ├── users.sql           # User SQL queries
│       │   └── sessions.sql        # Session SQL queries
│       └── migrations/             # Goose migration files
├── frontend-template/              # SSR static assets
├── .air.toml                       # Hot reload configuration
├── .goreleaser.yaml                # Release configuration
└── Makefile
```

## API Endpoints

### Pages (SSR)

* `GET /` - Home page
* `GET /contact` - Contact page
* `POST /contact` - Submit contact form
* `GET /login` - Login page
* `POST /login` - Authenticate user
* `GET /register` - Register page
* `POST /register` - Create new account
* `GET /logout` - Logout and clear session
* `GET /dashboard` - Dashboard (requires authentication)
* `GET /sitemap.xml` - Sitemap
* `GET /robots.txt` - Robots file
* `GET /favicon.ico` - Favicon

### API

* `GET /api/` - API information
* `GET /api/health` - Health check
* `GET /api/websocket` - WebSocket connection

## Deployment

### Download Binary

Get pre-built binaries from [releases page](https://github.com/PhuNguyenPT/GoApp/releases).

### Run Binary

**Linux/macOS:**
```bash
./GoApp
```

**Windows:**
```bash
GoApp.exe
```

## Environment Variables

Copy the example file and configure:

```bash
cp .env.example .env
```

Edit `.env` with your settings:

```env
# Application Configuration
PORT=8080
APP_ENV=dev

GIN_MODE=debug

# PostgreSQL Database Configuration
POSTGRES_VERSION=18
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DATABASE=goapp
POSTGRES_USERNAME=your_username
POSTGRES_PASSWORD=your_password
POSTGRES_SCHEMA=public
```

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/name`)
3. Commit changes (`git commit -m 'Add feature'`)
4. Push to branch (`git push origin feature/name`)
5. Open Pull Request

## License

AGPL-3.0 License - see LICENSE file for details.
