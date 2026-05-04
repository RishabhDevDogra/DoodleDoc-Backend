# DoodleDoc Backend

A modern, real-time collaborative drawing application demonstrating Event Sourcing and CQRS architecture patterns in action.

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go) ![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react) ![WebSocket](https://img.shields.io/badge/WebSocket-gorilla-ff6b6b?style=flat) ![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker) ![License](https://img.shields.io/badge/License-MIT-green)

---

## 🎯 Overview

DoodleDocs is a full-stack web application that lets teams collaborate on drawings in real-time. Behind the scenes, it showcases professional software architecture patterns — every brush stroke is an immutable event, every collaboration update flows through a CQRS pipeline, and your entire drawing history is preserved and replayable.

> This is a production-ready reference implementation for developers learning Event Sourcing and CQRS in Go.

---

## ✨ Features

| Feature | Description |
|---|---|
| 🎨 Real-time Drawing | Freehand canvas with colour and brush size control |
| 👥 Live Collaboration | Multiple users draw simultaneously via WebSocket |
| 📜 Version History | Complete event log showing who did what and when |
| ⏮️ Time Travel | Replay drawing to any previous version instantly |
| 🔄 Undo/Redo | Full undo/redo via event replay |
| 📋 Document Management | Create, update, delete, and organise multiple drawings |
| 🔗 Share Documents | Share direct links by document ID (Google Docs style URL) |
| 💾 Immutable History | All events stored — nothing is ever overwritten |

---

## 🏗️ Architecture

```
Browser (React)
    │
    ├── REST (HTTP)   ──► Handler ──► Service ──► EventStore  (write side)
    │                                    │
    │                                    └──► EventHandler ──► ProjectionStore (read side)
    │
    └── WebSocket     ──► Hub ◄── Service (broadcasts on every command)
```

| Layer | Technology | Purpose |
|---|---|---|
| Frontend | React 18, Canvas API, WebSocket | Interactive drawing interface |
| Backend | Go 1.26, `net/http` | Business logic and event processing |
| Real-time | gorilla/websocket | Bi-directional communication |
| Storage | In-memory Event Store + Projections | Event immutability and fast reads |
| Deployment | Docker Compose, nginx | Containerised orchestration |

---

## 🚀 Quick Start

### Prerequisites

- **Docker** (easiest), OR
- **Go 1.26+** and **Node 18+** (native development)

---

### Option 1 — Docker (recommended) 🐳

```bash
docker compose up --build
```

What starts:

| Service | URL |
|---|---|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |
| Swagger UI | http://localhost:8080/swagger/index.html |

To stop:

```bash
Ctrl+C
docker compose down
```

> First build downloads Go and Node base images (~200 MB). Subsequent builds use the layer cache and are fast.

---

### Option 2 — Native (two terminals) 💻

**Terminal 1 — backend:**

```bash
go run .
# or
make run        # same thing
make dev        # auto-reloads on file save (uses Air)
```

Backend starts on `http://localhost:8080`.

**Terminal 2 — frontend:**

```bash
cd DoodleDocs.Web
npm install     # first time only
npm start
```

Frontend starts on `http://localhost:3000`.

---

### Verifying everything works

1. Open `http://localhost:3000` — the DoodleDocs editor loads
2. URL updates to `http://localhost:3000/<uuid>` as soon as a document is selected (Google Docs-style)
3. `GET http://localhost:8080/health` → `{"status":"ok"}`
4. `http://localhost:8080/swagger/index.html` → Swagger UI with all endpoints
5. Open the same document URL in two tabs — draw in one, watch it appear in the other in real time

---

## 🧪 Testing

```bash
make test
# or
go test ./...
```

17 tests covering domain aggregate behaviour, service-layer integration, and HTTP handlers. No external dependencies or mocks needed — everything runs in-process.

---

## 📁 Project Structure

```
.
├── main.go                         # Entry point, PORT env var, Swagger metadata
├── Dockerfile                      # Multi-stage Go build → alpine runtime
├── docker-compose.yml              # Full-stack local orchestration
├── .gitattributes                  # Tells GitHub to count this repo as Go
│
├── internal/
│   ├── domain/
│   │   ├── base.go                 # DomainEvent interface + BaseEvent
│   │   ├── events.go               # 6 concrete event types
│   │   └── aggregate.go            # DocumentAggregate — rules + event replay
│   │
│   ├── infrastructure/
│   │   └── eventstore.go           # InMemoryEventStore (write side)
│   │
│   ├── readmodel/
│   │   ├── projection.go           # DocumentProjection read model
│   │   ├── projectionstore.go      # InMemoryProjectionStore (read side)
│   │   └── eventhandler.go         # Builds projections from events
│   │
│   ├── service/
│   │   └── document.go             # Commands, queries, Broadcaster interface
│   │
│   ├── handler/
│   │   ├── document.go             # REST handlers + Swagger annotations
│   │   └── comment.go              # Comment REST handlers
│   │
│   ├── hub/
│   │   └── hub.go                  # WebSocket hub, fan-out broadcasts
│   │
│   └── router/
│       └── router.go               # Route registration + CORS middleware
│
├── DoodleDocs.Web/                 # React frontend
│   ├── src/
│   │   ├── App.js                  # Root component, WebSocket lifecycle
│   │   ├── config.js               # API_URL + WS_HUB_URL config
│   │   ├── components/             # DocumentEditor, Comments, VersionHistory …
│   │   ├── pages/ShareView.js      # Shared document view
│   │   └── utils/userSession.js    # Random username session
│   ├── Dockerfile                  # React build → nginx
│   └── nginx.conf                  # Serves React SPA
│
└── docs/                           # Generated Swagger assets
```

---

## ⚙️ How It Works

### No auth, no database — by design

This is a reference implementation for learning Event Sourcing. It intentionally keeps things simple:

- **No authentication** — everyone gets a random username (`Artist#4832`)
- **LocalStorage session** — your user ID persists across page refreshes
- **In-memory store** — data lives in RAM and resets on server restart
- **All documents are public** — anyone with the URL can read and edit

### Testing real-time collaboration

1. Open `http://localhost:3000` in a normal window
2. Open an incognito window (`Cmd+Shift+N`) and go to the same document URL
3. Draw in one window — it appears instantly in the other via WebSocket

Or click **Share** to copy a direct link and open it on a different device on the same network.

---

## 🎓 Architecture Deep Dive

### Event Sourcing

Instead of storing the current state of a document, we store every event that changed it:

```
DocumentCreated  (v1)
  → TitleUpdated   (v2)
  → ContentUpdated (v3)
  → ContentUpdated (v4)
  → CommentAdded   (v5)

Current state  = replay all events
State at v3    = replay events[0..3]
```

Benefits:
- Complete audit trail — nothing is ever lost
- Time travel to any point in history for free
- `GET /api/document/{id}/history` and `GET /api/document/{id}/version/{n}` require zero extra work

### CQRS

Write and read sides are separated and never touch each other's data:

```
Commands (write side)              Queries (read side)
─────────────────────              ──────────────────────
CreateDocument  ──► EventStore     ProjectionStore ◄── EventHandler
UpdateDocument                     GetDocuments    ──► O(1) map lookup
DeleteDocument                     GetByID
AddComment
```

Benefits:
- Read path is always fast — projections are plain structs in a map
- Write path enforces invariants via the aggregate without touching read models
- Storage backends are behind interfaces — swap in Postgres with one implementation change

### Why `net/http` and not a framework

Go 1.22+ supports method+path routing natively (`GET /api/document/{id}`). Zero framework overhead, zero magic, one fewer dependency to learn or upgrade.

---

## 🔀 Concurrency Model (Goroutines in Action)

This is where Go shines. Every browser tab that connects to `/hubs/document` gets its own pair of goroutines — a writer and a reader — coordinated through channels and a mutex.

### The pattern

```
REST handler goroutine
    │
    └── hub.Broadcast(msg)
            │  (sync.RWMutex — safe for concurrent callers)
            ├── client A → chan []byte ──► writer goroutine A ──► WebSocket ──► Tab A
            ├── client B → chan []byte ──► writer goroutine B ──► WebSocket ──► Tab B
            └── client C → chan []byte ──► writer goroutine C ──► WebSocket ──► Tab C
```

### Why two goroutines per connection?

WebSocket connections in Go block on read and write. You can't do both in one goroutine without either missing messages or hanging. The split is:

- **Writer goroutine** — owns all writes to the WebSocket. Blocks on the client's send channel, writes whatever arrives.
- **Reader goroutine** — owns all reads. Even though this server only pushes (never receives canvas data over WS), you *must* read continuously or the browser's ping frames go unacknowledged and the connection drops.

```go
// Writer goroutine — one per connected tab
go func() {
    defer func() {
        h.unregister(c)
        conn.Close()
    }()
    for msg := range c.send {   // blocks until a message arrives
        conn.WriteMessage(websocket.TextMessage, msg)
    }
}()

// Reader goroutine — keeps the connection alive
go func() {
    defer func() {
        h.unregister(c)
        conn.Close()
    }()
    conn.SetPongHandler(func(string) error {
        conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })
    for {
        if _, _, err := conn.ReadMessage(); err != nil {
            break   // client disconnected — triggers unregister via defer
        }
    }
}()
```

### Why a buffered channel?

```go
c := &client{send: make(chan []byte, 256)}
```

The `Broadcast` method loops over all clients and sends to each channel. If a client's channel were unbuffered, a slow or stalled browser tab would block the broadcast for *everyone*. The buffer of 256 means the broadcaster can move on immediately — if a client falls too far behind, its channel fills and it gets dropped.

### Why `sync.RWMutex` and not a channel for the client map?

Multiple goroutines read the client map simultaneously (every broadcast). Using `RWMutex` lets all of them read in parallel and only locks exclusively when a client connects or disconnects. A channel-based approach would serialize all reads unnecessarily.

```go
// Broadcast — called from service layer after every command
func (h *Hub) Broadcast(msg Message) {
    data, _ := json.Marshal(msg)
    h.mu.RLock()                          // read lock — allows concurrent broadcasts
    defer h.mu.RUnlock()
    for c := range h.clients {
        select {
        case c.send <- data:              // non-blocking send
        default:
            h.unregister(c)               // buffer full — drop the slow client
        }
    }
}
```

### Key Go concurrency concepts used here

| Concept | Where | Why |
|---|---|---|
| `goroutine` | writer + reader per connection | Lightweight — thousands can run concurrently |
| `chan []byte` | per-client send queue | Decouples broadcaster from slow writers |
| `sync.RWMutex` | client map | Multiple concurrent readers, exclusive writers |
| `select` with `default` | Broadcast loop | Non-blocking channel send — never hangs |
| `defer` + cleanup | both goroutines | Guarantees unregister even on panic or error |



## 📊 API Reference

### Documents

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/document` | List all documents |
| GET | `/api/document/{id}` | Get document |
| GET | `/api/document/{id}/history` | Full event log |
| GET | `/api/document/{id}/version/{n}` | Document state at version N |
| POST | `/api/document` | Create document |
| PUT | `/api/document/{id}` | Update document |
| DELETE | `/api/document/{id}` | Delete document |

### Comments

| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/document/{id}/comments` | List comments |
| POST | `/api/document/{id}/comments` | Add comment |
| DELETE | `/api/document/{id}/comments/{commentId}` | Delete comment |

### Infra

| Method | Endpoint | Description |
|---|---|---|
| GET | `/health` | Liveness check |
| GET | `/hubs/document` | WebSocket upgrade |
| GET | `/swagger/` | Swagger UI |

### Example requests

```bash
# Create a document
curl -X POST http://localhost:8080/api/document \
  -H "Content-Type: application/json" \
  -d '{"title": "My Drawing", "userId": "abc", "userName": "Artist#4832"}'

# Get event history
curl http://localhost:8080/api/document/{id}/history

# Restore to version 3
curl http://localhost:8080/api/document/{id}/version/3
```

### WebSocket messages (server → client)

```json
{ "type": "DocumentCreated", "payload": { "documentId": "...", "title": "..." } }
{ "type": "DocumentUpdated", "payload": { "documentId": "..." } }
{ "type": "DocumentDeleted", "payload": { "documentId": "..." } }
{ "type": "CommentAdded",    "payload": { "documentId": "..." } }
{ "type": "EventAdded",      "payload": { "documentId": "...", "eventType": "...", "description": "...", "timestamp": "..." } }
{ "type": "Connected",       "payload": null }
```

---

## 🔧 Configuration

### Backend environment variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Set automatically by Render and most hosts |
| `FRONTEND_ORIGIN` | — | Deployed frontend URL, enables CORS for that origin |

### Frontend environment variables

| Variable | Default | Description |
|---|---|---|
| `REACT_APP_API_URL` | `http://localhost:8080` | Backend base URL for API and WebSocket |

---

## 🚢 Deployment (Render)

Two services, each deploying from the same repo.

### Frontend service

- Runtime: **Docker**
- Root Directory: `DoodleDocs.Web`
- Dockerfile Path: `Dockerfile`
- Docker Build Context: `.`
- Env: `REACT_APP_API_URL=https://doodledocs-backend.onrender.com`

### Backend service

- Runtime: **Go**
- Root Directory: *(leave blank)*
- Build Command: `go build -o app .`
- Start Command: `./app`
- Health Check Path: `/health`
- Env: `FRONTEND_ORIGIN=https://doodledocs.onrender.com`

### Production checklist

- [ ] Set `FRONTEND_ORIGIN` on the backend service
- [ ] Set `REACT_APP_API_URL` on the frontend service
- [ ] Redeploy frontend after backend is live
- [ ] Replace in-memory stores with a persistent event store for production data

---

## 💡 Key Design Decisions

| Decision | Rationale |
|---|---|
| Event Sourcing | Complete audit trail and time travel come for free |
| CQRS | Read and write models optimised independently |
| In-memory store | Fast for demos; interfaces make swapping to Postgres one file change |
| `net/http` only | Go 1.22 standard library covers all routing needs, no framework overhead |
| gorilla/websocket | Battle-tested, handles ping/pong and graceful close correctly |
| Plain WebSocket | No third-party client library needed; all browsers support WebSocket natively |

---

## 🐛 Troubleshooting

**Port already in use**

```bash
lsof -ti :8080 | xargs kill -9
lsof -ti :3000 | xargs kill -9
docker compose up
```

**Docker build fails first time**

Ensure Docker Desktop is running (`open -a Docker`), then retry.

**Frontend shows blank page after deploy**

Check that `REACT_APP_API_URL` is set on the frontend Render service and that the backend service is live before redeploying the frontend.

**Real-time not working after deploy**

Ensure `FRONTEND_ORIGIN` on the backend matches your frontend URL exactly (including `https://`).

---

## 📚 Learning Resources

- [Event Sourcing — Martin Fowler](https://martinfowler.com/eaaDev/EventSourcing.html)
- [CQRS — Microsoft](https://learn.microsoft.com/en-us/azure/architecture/patterns/cqrs)
- [gorilla/websocket](https://github.com/gorilla/websocket)
- [Effective Go](https://go.dev/doc/effective_go)

---

## 📄 License

MIT License © 2026

Built with Go, Event Sourcing, CQRS, and real-time WebSocket collaboration.
