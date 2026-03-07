# Status Platform

A production-ready, self-hosted status page and monitoring platform — similar to Atlassian Statuspage, BetterStack, and UptimeRobot. Single-tenant, fully open-source.

![Status Platform](https://img.shields.io/badge/Go-1.21-00ADD8?logo=go) ![React](https://img.shields.io/badge/React-18-61DAFB?logo=react) ![MongoDB](https://img.shields.io/badge/MongoDB-7-47A248?logo=mongodb) ![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker)

## Features

- **Public Status Page** — Atlassian-style status page with 90-day uptime history bars, active incidents, and scheduled maintenance
- **Admin CMS Dashboard** — Manage components, incidents, maintenance windows, monitors, and subscribers
- **Automated Monitoring** — HTTP, TCP, DNS, and ICMP ping checks with configurable intervals
- **Auto Incident Management** — Automatically creates incidents after 3 consecutive failures; auto-resolves when healthy
- **Real-time Updates** — WebSocket push for instant status changes without page refresh
- **Incident Timeline** — Full update history with status transitions (investigating → identified → monitoring → resolved)
- **Email Subscribers** — Collect and manage subscriber emails for status notifications
- **JWT Authentication** — Secure admin-only routes with bcrypt password hashing
- **90-Day Uptime History** — Daily uptime aggregation with color-coded bars per component
- **Docker Compose** — One-command deployment

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend API | Go 1.21, Gin, JWT, bcrypt |
| Monitoring Worker | Go, goroutines, ICMP/TCP/DNS/HTTP |
| Frontend | React 18, Vite, TypeScript, Tailwind CSS |
| Charts | Recharts |
| Database | MongoDB 7 |
| Cache / Pub-Sub | Redis |
| Real-time | WebSocket (gorilla/websocket) |
| Deployment | Docker Compose |

## Project Structure

```
status-platform/
├── apps/
│   ├── api/           # Gin HTTP API server
│   ├── worker/        # Background monitoring worker
│   └── web/           # React + Vite frontend
├── internal/
│   ├── database/      # MongoDB + Redis connection helpers
│   ├── handlers/      # HTTP handlers (auth, components, incidents, etc.)
│   ├── middleware/     # JWT auth middleware
│   └── models/        # MongoDB document models
├── configs/           # Config loader (env vars)
├── docker/            # Dockerfiles for api, worker, web
├── scripts/           # seed.go — seed sample data
├── docker-compose.yml
└── .env.example
```

## Quick Start

### Prerequisites

- Docker and Docker Compose

### 1. Clone and configure

```bash
git clone <repo-url>
cd status-platform

cp .env.example .env
# Edit .env if you want to change defaults (optional for local dev)
```

### 2. Start all services

```bash
docker compose up --build
```

This starts:
- **API** on port `8080`
- **Worker** (monitoring daemon, no HTTP port)
- **Web** on port `3000`
- **MongoDB** on port `27017`
- **Redis** on port `6379`

### 3. (Optional) Seed sample data

```bash
# After services are running
go run scripts/seed.go
```

This creates sample components, subcomponents, and a resolved incident with full timeline.

### 4. Access the platform

| URL | Description |
|-----|-------------|
| http://localhost:3000 | Public status page |
| http://localhost:3000/admin/login | Admin login |
| http://localhost:8080/api | API base |

**Default admin credentials:**
```
Email:    admin@statusplatform.com
Password: admin123
```

> Change these in `.env` before deploying to production.

## Environment Variables

Copy `.env.example` to `.env` and configure:

```env
# MongoDB
MONGO_URI=mongodb://mongo:27017
MONGO_DB_NAME=statusplatform

# Redis
REDIS_ADDR=redis:6379

# API
PORT=8080
JWT_SECRET=change-me-in-production

# Default admin account (created on first startup)
ADMIN_EMAIL=admin@statusplatform.com
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
```

## API Endpoints

### Public (no auth required)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/status/summary` | Overall system status |
| GET | `/api/status/components` | All components with subcomponents |
| GET | `/api/status/incidents` | Active and recent incidents |
| POST | `/api/subscribe` | Subscribe an email for notifications |

### Admin (JWT required)

#### Auth
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/login` | Login, returns JWT token |
| GET | `/api/auth/me` | Current admin info |

#### Components
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/components` | List all components |
| POST | `/api/components` | Create component |
| PATCH | `/api/components/:id` | Update component |
| DELETE | `/api/components/:id` | Delete component |
| GET | `/api/components/:id/subcomponents` | List subcomponents by parent |

#### Subcomponents
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/subcomponents` | Create subcomponent |
| PATCH | `/api/subcomponents/:id` | Update subcomponent |

#### Incidents
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/incidents` | List incidents |
| POST | `/api/incidents` | Create incident |
| PATCH | `/api/incidents/:id` | Update incident |
| POST | `/api/incidents/:id/update` | Add timeline update |
| GET | `/api/incidents/:id/updates` | Get timeline updates |

#### Maintenance
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/maintenance` | List maintenance windows |
| POST | `/api/maintenance` | Schedule maintenance |
| PATCH | `/api/maintenance/:id` | Update maintenance |

#### Monitors
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/monitors` | List monitors |
| POST | `/api/monitors` | Create monitor |
| DELETE | `/api/monitors/:id` | Delete monitor |
| GET | `/api/monitors/:id/logs` | Monitor check logs |
| GET | `/api/monitors/:id/uptime` | 90-day uptime data |

#### Subscribers
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/subscribers` | List subscribers |
| DELETE | `/api/subscribers/:id` | Remove subscriber |

## Monitoring Worker

The worker runs continuous health checks on all configured monitors:

- **HTTP** — GET request, checks status code (default: 200)
- **TCP** — TCP dial to host:port
- **DNS** — Resolves a hostname, checks for expected IP
- **Ping (ICMP)** — ICMP echo request (requires root/capabilities in Docker)

**Auto-incident behavior:**
- After **3 consecutive failures** → automatically creates a new incident and sets component status to `major_outage`
- When monitor recovers → auto-resolves the incident and restores component status to `operational`

Check interval is configurable per monitor (default: 60 seconds).

## WebSocket

Connect to `ws://localhost:8080/ws` for real-time events:

```json
{ "type": "incident_created", "data": { ... } }
{ "type": "incident_resolved", "data": { ... } }
{ "type": "component_updated", "data": { ... } }
```

## Component Status Values

| Status | Description |
|--------|-------------|
| `operational` | Fully operational |
| `degraded_performance` | Slower than normal |
| `partial_outage` | Some requests failing |
| `major_outage` | Service unavailable |
| `maintenance` | Scheduled maintenance |

## Production Deployment

1. **Change secrets** in `.env`:
   - `JWT_SECRET` — use a long random string
   - `ADMIN_PASSWORD` — use a strong password

2. **Set MongoDB credentials** — update `MONGO_URI` with auth

3. **Reverse proxy** — put Nginx or Caddy in front, terminate TLS

4. **ICMP monitoring** — the worker container needs `NET_RAW` capability (already configured in `docker-compose.yml`)

5. **Backups** — back up the MongoDB `statusplatform` database

## License

MIT
