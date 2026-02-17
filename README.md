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

# Start database
make docker-run

# Run with hot reload
make watch
```

Visit http://localhost:8080

## Makefile Commands

### Build

* `make all` - Build and test
* `make templ-generate` - Generate templ files
* `make tailwind-build` - Build Tailwind CSS
* `make build` - Build application binary

### Development

* `make watch` - Hot reload (recommended)
* `make run` - Run SSR + SPA frontend

### Database

* `make docker-run` - Start PostgreSQL container
* `make docker-down` - Stop PostgreSQL container

### Testing

* `make test` - Run all tests
* `make itest` - Run integration tests

### Cleanup

* `make clean` - Remove binary and generated files

## Project Structure

```
GoApp/
├── cmd/api/              # Application entry point
├── internal/
│   ├── server/           # HTTP server and routes
│   ├── views/            # Templ templates
│   └── database/         # Database logic
├── frontend/             # SPA frontend
├── frontend-template/    # SSR static assets
├── .air.toml             # Hot reload configuration
├── .goreleaser.yml       # Release configuration
└── Makefile
```

## API Endpoints

### Pages (SSR)

* `GET /` - Home page
* `GET /contact` - Contact page
* `POST /contact` - Submit contact form

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
PORT=8080
APP_ENV=local

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