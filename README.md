# StatusForge

StatusForge is a self-hosted status page and monitoring platform with a Go/Gin backend, a React/Vite frontend, MongoDB persistence, Redis-backed realtime support, and a built-in monitoring worker.

It is designed to run as a single unified server binary that serves both the API and the embedded web app.

## Features

- Public status page with realtime updates over WebSocket
- Category-level status detail pages at `/status/:categoryPrefix`
- Component and subcomponent status tracking
- Incident and maintenance management
- HTTP, TCP, DNS, ping, and SSL monitoring
- SSL certificate and domain expiry warning support
- Admin console with RBAC and MFA-aware authentication flow
- Subscriber and webhook channel management
- Themeable public status page with branding and layout controls
- Unified Docker deployment for API, worker, and frontend assets

## Architecture

### Backend

- **Language/runtime:** Go 1.26
- **HTTP framework:** Gin
- **Entrypoint:** `cmd/server/main.go`
- **Server bootstrap:** `internal/server/server.go`
- **Route registration:** `internal/server/api_routes.go`
- **Mongo connection:** `internal/database/mongo.go`
- **Redis connection:** `internal/database/redis.go`
- **Embedded frontend assets:** `internal/embed/embed.go`

The backend starts from `server.RunServer()`, loads `.env`, connects to MongoDB and Redis, creates the Gin engine, registers API routes, starts the monitoring worker when enabled, seeds the initial admin user, and serves the embedded frontend for non-API routes.

### Frontend

- **Framework:** React 18
- **Build tool:** Vite 5
- **Language:** TypeScript
- **Router:** `react-router-dom`
- **HTTP client:** Axios
- **Icons/UI utilities:** `lucide-react`
- **Frontend root:** `apps/web/src/App.tsx`

The frontend is built into static assets and copied into `internal/embed/dist` during Docker builds so the Go server can serve the UI directly.

### Data stores

- **MongoDB:** primary application database
- **Redis:** connectivity check plus realtime/pub-sub support used by the unified server and worker paths

## Repository Layout

```text
.
├── cmd/server/                  # Go application entrypoint
├── configs/                     # Environment-driven config loading
├── internal/
│   ├── database/                # MongoDB and Redis connection helpers
│   ├── embed/                   # Embedded frontend assets
│   ├── handlers/                # HTTP handlers and websocket hub
│   ├── middleware/              # Auth, MFA, and role enforcement
│   ├── models/                  # Mongo-backed domain models
│   ├── repository/              # Data access layer
│   ├── server/                  # Unified server + worker bootstrap
│   ├── services/                # Business logic services
│   └── utils/                   # Monitor protocol utilities
├── apps/web/                    # React/Vite frontend
├── Dockerfile                   # Multi-stage unified image build
├── docker-compose.yml           # Local stack: app + Mongo + Redis
├── Makefile                     # Docker compose convenience commands
└── .env.example                 # Local configuration template
```

## Public Routes

Registered from `internal/server/api_routes.go`:

- `GET /health`
- `GET /ws`
- `GET /api/status/summary`
- `GET /api/status/components`
- `GET /api/status/incidents`
- `GET /api/status/category/:prefix`
- `GET /api/v1/status/category/:prefix`
- `GET /api/status/settings`
- `POST /api/subscribe`

The public web UI includes:

- `/` — main status page
- `/status/:categoryPrefix` — category detail page
- `/history` — incident history page

## Admin API Surface

The admin API is protected in layers:

1. JWT authentication via `AuthMiddleware`
2. MFA verification via `RequireMFAVerified`
3. Role checks via `RequireRoles`

Major admin route groups include:

- Auth/profile/MFA endpoints
- Incidents and incident updates
- Maintenance windows
- Components and subcomponents
- Monitors, logs, uptime, history, and outage data
- Subscribers
- Status page settings
- Webhook channels
- Users and user invitations

## Monitoring Worker

The unified server can start an internal monitoring worker from `internal/server/worker.go` when `ENABLE_WORKER=true`.

The worker:

- polls monitor definitions from MongoDB
- executes monitor checks on a ticker
- stores monitor logs
- updates component/subcomponent status
- updates daily uptime records
- detects outages
- applies SSL/domain warning degradation logic
- updates maintenance status

Supported monitor types visible in the worker path include:

- HTTP
- TCP
- DNS
- Ping
- SSL

## Environment Variables

From `configs/config.go` and `.env.example`:

| Variable | Required | Default | Purpose |
|---|---|---:|---|
| `MONGO_URI` | Yes | `mongodb://localhost:27017` | MongoDB connection string |
| `MONGO_DB_NAME` | Yes | `statusplatform` | Mongo database name |
| `REDIS_ADDR` | Yes | `localhost:6379` | Redis address |
| `JWT_SECRET` | Yes | fallback present, override in real deployments | JWT signing secret |
| `MFA_SECRET_KEY` | Yes | empty | MFA-related secret material |
| `PORT` | No | `8080` | HTTP server port |
| `ADMIN_EMAIL` | Bootstrap only | `admin@statusplatform.com` | Seeded admin email |
| `ADMIN_USERNAME` | Bootstrap only | `admin` | Seeded admin username |
| `ADMIN_PASSWORD` | Bootstrap only | `admin123` | Seeded admin password |
| `ENABLE_WORKER` | No | `true` | Enables the internal monitoring worker |

## Local Development

### Prerequisites

- Go 1.26+
- Node.js 20+
- MongoDB
- Redis

### 1. Configure environment

```bash
cp .env.example .env
```

Adjust values for your local MongoDB and Redis instances as needed.

### 2. Run the backend

```bash
go mod download
go run cmd/server/main.go
```

This starts the unified server on `http://localhost:8080` by default.

### 3. Run the frontend in Vite dev mode

In another terminal:

```bash
cd apps/web
npm install
npm run dev
```

The Vite dev server runs on `http://localhost:3000` and proxies:

- `/api` → `http://localhost:8080`
- `/ws` → `ws://localhost:8080`

## Docker Compose Quick Start

```bash
cp .env.example .env
docker compose up --build
```

Or with `make`:

```bash
make up-build
```

Services defined in `docker-compose.yml`:

- `server` — unified StatusForge app on port `8080`
- `mongo` — MongoDB 7 on port `27017`
- `redis` — Redis 7 on port `6379`

The compose stack also grants `NET_RAW` to the server container for ICMP ping monitor support.

## Build

### Frontend

```bash
cd apps/web
npm run build
```

### Backend

```bash
go build -o server cmd/server/main.go
```

### Docker image

The root `Dockerfile` uses a multi-stage build:

1. build the React frontend with Node 20
2. build the Go backend with Go 1.26
3. copy frontend assets into `internal/embed/dist`
4. produce a final minimal Alpine image running `/app/server`

## Make Targets

Common targets from `Makefile`:

- `make up`
- `make up-build`
- `make down`
- `make down-v`
- `make restart`
- `make logs`
- `make logs-server`
- `make logs-mongo`
- `make logs-redis`
- `make ps`
- `make shell-server`
- `make shell-mongo`
- `make shell-redis`

## Notable Recent Status Feature Additions

Current changes in this branch add category-level public status support:

- backend category detail endpoint via `GetStatusCategory`
- repository/service layering for category summary generation
- frontend route for `/status/:categoryPrefix`
- clickable category cards from the main status page

This functionality is implemented across:

- `internal/repository/status_repository.go`
- `internal/services/status/service.go`
- `internal/handlers/status.go`
- `internal/server/api_routes.go`
- `apps/web/src/pages/StatusCategoryPage.tsx`
- `apps/web/src/lib/api.ts`
- `apps/web/src/hooks/useApi.ts`
- `apps/web/src/types/index.ts`
- `apps/web/src/pages/StatusPage.tsx`
- `apps/web/src/App.tsx`

## Screenshots

Run the screenshot generator after UI updates:

```bash
npm run docs:screenshots --prefix apps/web
```

Available screenshots are stored in `docs/screenshots/`.

## Testing and Verification

Useful commands for local verification:

```bash
go test ./...
cd apps/web && npm run build
```

## Security Notes

- Change bootstrap admin credentials before real deployment.
- Replace default JWT and MFA secrets in any non-local environment.
- Review CORS settings in `internal/server/api_routes.go` before internet-facing deployment.

## License

This repository is licensed under the MIT License. See `LICENSE`.
