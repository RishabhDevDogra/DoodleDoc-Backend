# DoodleDoc Backend

A Go REST API server.

## Running

```bash
go run .
```

The server starts on `http://localhost:8080`.

Or use Make targets:

```bash
make run
```

## Development Workflow

Use auto-reload during development:

```bash
make dev
```

Useful commands:

```bash
make test
make build
make docs
make tidy
make help
```

## API Docs (Swagger UI)

Open Swagger UI in your browser:

```text
http://localhost:8080/swagger/index.html
```

Regenerate API docs after changing annotations:

```bash
make docs
```

## Endpoints

| Method | Path      | Description        |
|--------|-----------|--------------------|
| GET    | /health   | Liveness check     |
| GET    | /cities   | List city names    |
| GET    | /swagger/ | Swagger UI assets  |

## Project Structure

```
.
├── main.go
└── internal/
    ├── handler/     # Controller-like HTTP handlers
    ├── model/       # Domain models/entities
    ├── repository/  # Data access layer
    ├── service/     # Business logic layer
    └── router/      # Route registration
```

## Request Flow (Mature)

```text
Router -> Handler (Controller) -> Service -> Repository -> Response
```
