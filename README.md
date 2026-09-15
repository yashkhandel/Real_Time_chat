# Wire — Realtime Chat & Notification System

A full-stack realtime messaging platform built on WebSockets, with a
Redis-backed pub-sub layer so the backend can scale to multiple instances
without losing message delivery — the same pattern used by production chat
systems (Slack, Discord-style architectures) at a smaller scale.

## Architecture

```
┌─────────────┐   WebSocket (JWT)   ┌──────────────┐
│   React     │ ◀─────────────────▶ │  Go Backend  │
│  (Vite)     │      REST (auth,    │ (Gin + Hub)  │
│             │   rooms, history)   │              │
└─────────────┘                     └──────┬───────┘
                                            │
                              ┌─────────────┼─────────────┐
                              ▼                            ▼
                       ┌──────────────┐            ┌──────────────┐
                       │  PostgreSQL  │            │    Redis     │
                       │ (persistence)│            │  (pub/sub —  │
                       └──────────────┘            │ cross-instance
                                                    │  broadcast)  │
                                                    └──────────────┘
```

**Why Redis pub-sub?** Each backend instance only knows about WebSocket
connections made directly to it. When a message needs to reach a user
connected to a *different* instance (common once you run more than one
backend process behind a load balancer), the Hub publishes the event to a
shared Redis channel; every instance subscribes and forwards to its own
locally-connected clients. This is what makes the chat system horizontally
scalable instead of single-process only.

## Features

- JWT authentication (register/login), token passed via query param for the
  WebSocket handshake (headers aren't available post-upgrade)
- Real-time message delivery over WebSockets, persisted to PostgreSQL
- Multi-instance-safe broadcast via Redis pub-sub
- Live typing indicators and online/offline presence
- Room-based conversations (1:1 or group), with membership stored relationally
- Frontend auto-reconnect with exponential backoff on socket drop
- Multi-tab support — a user can have several active connections at once

## Running locally

```bash
git clone <your-repo-url>
cd realtime-chat
docker-compose up --build
```

- Frontend: http://localhost:3001
- Backend API: http://localhost:8081/api/v1
- WebSocket endpoint: ws://localhost:8081/api/v1/ws?token=<jwt>

## Running services individually (dev mode)

**Backend**
```bash
cd backend
cp .env.example .env
go mod tidy
go run main.go
```

**Frontend**
```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Redis and PostgreSQL still need to be running locally (or via
`docker-compose up postgres redis`).

## API Reference

| Method | Endpoint                          | Auth | Description                          |
|--------|-------------------------------------|------|----------------------------------------|
| POST   | `/api/v1/auth/register`            | No   | Create a new user                      |
| POST   | `/api/v1/auth/login`               | No   | Login, returns JWT                     |
| GET    | `/api/v1/users`                    | Yes  | List users (to start conversations)    |
| POST   | `/api/v1/rooms`                    | Yes  | Create a room with members             |
| GET    | `/api/v1/rooms`                    | Yes  | List rooms the user belongs to         |
| GET    | `/api/v1/rooms/:id/messages`       | Yes  | Paginated message history              |
| GET    | `/api/v1/ws?token=<jwt>`           | Yes  | WebSocket upgrade                      |

### WebSocket message types

Client → Server: `{"type": "message", "room_id": 1, "body": "hello"}`,
`{"type": "typing", "room_id": 1}`

Server → Client: `{"type": "message" | "typing" | "presence", "room_id": ..., "data": {...}}`

## Why this project

Demonstrates real-time system design: connection lifecycle management,
pub-sub for horizontal scaling, presence/typing ephemeral state vs.
persisted data, and a reconnect-resilient frontend — core concerns in any
chat, collaboration, or live-notification product.

## License

MIT
