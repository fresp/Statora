# StatusForge: Self-Hosted Status Page & Monitoring Platform

StatusForge is a production-ready, self-hosted status page and monitoring platform, inspired by leading services like Atlassian Statuspage, Better Stack, and UptimeRobot. It provides a robust solution for businesses to maintain transparency and communicate service health to their users effectively.

## ✨ Key Features

- **Real-time Monitoring**: Track the uptime and performance of your services with instant updates.
- **Advanced Certificate & Domain Expiry Monitoring**: Track SSL certificate expiry and domain expiration windows with configurable alert thresholds.
- **Incident & Maintenance Management**: Efficiently create, update, and resolve incidents and schedule maintenance events.
- **Admin Console UX**: Navigate the admin area more efficiently with grouped, collapsible sidebar sections and clearer visual hierarchy.
- **Customizable Public Status Page**: Brand your status page with theme presets, light/dark/system modes, typography controls, background and hero images, layout variants, custom metadata, and custom CSS.
- **Role-Based Access Control (RBAC)**: Secure admin workflows with distinct `admin` and `operator` roles.
- **Multi-Factor Authentication (MFA)**: Enhance security for administrative access.
- **Subscriber Management**: Allow users to subscribe to updates for incidents and maintenance.
- **Webhook Integrations**: Configure webhooks for automated notifications on service status changes.
- **Flexible HTTPS Checks**: For HTTPS-based HTTP monitors, optionally ignore TLS certificate errors during availability checks when certificate expiry monitoring is not enabled.
- **Self-Hosted**: Full control over your data and infrastructure.

## 🚀 Screenshots

Run `npm run docs:screenshots --prefix apps/web` to regenerate these images after UI changes.

| Admin Dashboard | Public Status Page | Incident History |
|-----------------|--------------------|------------------|
| ![Admin Dashboard](docs/screenshots/admin-dashboard.png) | ![Public Status Page](docs/screenshots/status-page.png) | ![Incident History](docs/screenshots/incident-history.png) |

| Monitoring Table | Theme Settings |
|------------------|----------------|
| ![Monitoring Table](docs/screenshots/monitoring-table.png) | ![Theme Settings](docs/screenshots/theme-settings.png) |

## 🛠️ Tech Stack

StatusForge is built with a modern, efficient, and scalable technology stack.

- **Backend**: Go (GoLang 1.26+) with Gin HTTP web framework
- **Frontend**: React 18 with Vite for a fast development experience and TypeScript for type safety
- **Database**: MongoDB for flexible data storage
- **Caching & Pub/Sub**: Redis for high-performance caching and real-time updates
- **Styling**: Tailwind CSS for utility-first styling

## ⚡ Quick Start with Docker Compose

The fastest way to get StatusForge up and running locally is using Docker Compose.

### Prerequisites

Ensure you have Docker and Docker Compose installed on your system.

### Steps

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/fresp/StatusForge.git
    cd StatusForge
    ```

2.  **Configure Environment Variables:**
    Copy the example environment file:
    ```bash
    cp .env.example .env
    ```
    You can customize these variables in the `.env` file. Important variables include:
    -   `MONGODB_URI`: MongoDB connection string. The database name is expected to be part of this URI.
    -   `REDIS_URI`: Redis connection address. The app accepts either `host:port` or a `redis://` URI.
    -   `JWT_SECRET`: Secret key for JWT authentication ( **change this in production!** ).
    -   `MFA_SECRET_KEY`: Secret key for MFA ( **change this in production!** ).
    -   `PORT`: The port StatusForge will run on (default: `8080`).
    -   `ADMIN_EMAIL`, `ADMIN_USERNAME`, `ADMIN_PASSWORD`: Credentials for the initial admin user (used for bootstrap only).
    -   `ENABLE_WORKER`: Enables the background worker that executes monitor checks and writes monitor status updates.
    -   `GRACEFUL_SHUTDOWN`: Enables signal-aware graceful shutdown handling (default: `true`).
    -   `SHUTDOWN_TIMEOUT`: Timeout in seconds for graceful shutdown (default: `30`).

3.  **Start the Services:**
    Build and start all services using Docker Compose:
    ```bash
    docker compose up --build
    ```
    Alternatively, if you have `make` installed:
    ```bash
    make up-build
    ```

4.  **Access StatusForge:**
    Once the services are up, access StatusForge in your web browser:
    -   **Public Status Page**: `http://localhost:8080`
    -   **Admin Panel**: `http://localhost:8080/admin`
    -   **Health Check**: `http://localhost:8080/health`

    The default admin credentials are `admin@statusplatform.com` with password `admin123`. **It is crucial to change these credentials immediately after the first login in a production environment.**

### Runtime Notes

-   The Docker Compose stack provisions `server`, `mongo`, and `redis` services and wires them together through `.env`.
-   The backend defaults to `MONGODB_URI=mongodb://localhost:27017` and `REDIS_URI=localhost:6379` when environment variables are not provided.
-   The checked-in `.env` file overrides those defaults for the Compose network, so containerized development uses the in-stack MongoDB and Redis services by default.
-   The MongoDB and Redis clients both apply basic connection-pool tuning in the current backend implementation.

## ⚙️ Local Development

For developers who want to contribute or customize StatusForge, you can run the backend and frontend separately.

### Prerequisites

-   Go 1.26+
-   Node.js 20+
-   MongoDB instance (local or remote)
-   Redis instance (local or remote)

### Backend

1.  **Configure Environment Variables**: Copy `.env.example` to `.env` as described in the Docker Quick Start.
2.  **Install Go dependencies**:
    ```bash
    go mod download
    ```
3.  **Run the Go server**:
    ```bash
    go run cmd/server/main.go
    ```
    By default the backend listens on `http://localhost:8080`.

### Frontend

1.  **Navigate to the frontend directory**:
    ```bash
    cd apps/web
    ```
2.  **Install dependencies**:
    ```bash
    npm install
    ```
3.  **Start the Vite development server**:
    ```bash
    npm run dev
    ```
    The Vite development server will provide live reloading for frontend changes.

## 📦 Build & Deployment

### Building for Production

To build the frontend and backend for production:

1.  **Frontend Build**:
    ```bash
    cd apps/web
    npm run build
    ```
2.  **Backend Build**:
    ```bash
    go build -o server cmd/server/main.go
    ```
    The `Dockerfile` handles this multi-stage build process automatically for Docker deployments.

### Docker Deployment

The provided `Dockerfile` creates a minimal Alpine-based image embedding the built frontend assets into the Go binary. The `docker-compose.yml` orchestrates the `server`, `mongo`, and `redis` services, exposes the application on port `8080`, and includes health checks for the backing data stores.

For detailed Docker operations, refer to the `Makefile` for convenient commands like `make up`, `make up-build`, `make down`, `make down-v`, `make logs`, `make logs-server`, `make ps`, and `make shell-server`.

## 🎨 Status Page Branding & Theme Controls

Administrators can customize the public status page from the admin settings screen without rebuilding the application.

- **Branding assets**: Site name, logo URL, background image URL, and hero image URL.
- **Theme presets**: Built-in `default`, `ocean`, and `graphite` presets.
- **Color modes**: `light`, `dark`, or `system` mode selection.
- **Palette editing**: Separate light and dark palettes for primary, background, text, and accent colors.
- **Typography**: Configurable font family and font scale (`sm`, `md`, `lg`).
- **Layout variants**: `classic`, `compact`, `minimal`, and `cards` layouts for the public page.
- **Incident history UX**: The public status page surfaces a rolling 7-day incident snapshot by default, and `/history` exposes the archive through quarter navigation with month-grouped incident lists and empty-month states.
- **Preview tooling**: The admin console includes a live preview before saving changes.
- **Validation**: Backend validation enforces `http(s)` URLs for branding assets and `#RRGGBB` hex colors for theme values.

These settings are served through the status page settings API and pushed to connected clients through the existing realtime update flow.

## 🤝 Contributing

We welcome contributions to StatusForge! Whether it's bug reports, feature requests, documentation improvements, or code contributions, your help is valuable.

-   Fork the repository.
-   Create a new branch for your feature or bug fix.
-   Submit a pull request with a clear description of your changes.

Please ensure your code adheres to existing style guides and passes all tests.

## 📄 License

StatusForge is open-source software licensed under the [MIT License](LICENSE).
