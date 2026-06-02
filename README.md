# idp API

idp is a Go backend service for managing machines, networks, repository connections, and deployments with authentication and async job processing.

## Architecture and Tooling

- **Dependency Injection:** `go.uber.org/fx` (Uber Fx) is used to wire and manage dependencies/modules across the application lifecycle.
- **HTTP API:** `gofiber/fiber` powers the REST API and middleware stack.
- **Database Access:** `prisma-client-go` is used for schema-driven data access and migrations.
- **Async Processing:** `rabbitmq/amqp091-go` is used for queue-based background tasks.
- **Caching/State:** `go-redis/redis/v9` is used for Redis integration.
- **Scheduling:** `robfig/cron/v3` is used for scheduled/background jobs.
- **API Docs:** `swaggo/swag` is used to generate Swagger/OpenAPI docs.
- **WebSockets:** native WebSocket communication is handled via `nhooyr.io/websocket`.
- **Configuration/Secrets:** `spf13/viper` is used for local env loading, with secrets sourced via the project secrets factory (`local` / `aws`).

## Requirements

- Go `1.25.1+`
- Docker and Docker Compose

## Environment Setup

1. Copy env template:

```bash
cp .env.sample .env
```

2. Fill required values in `.env`:

- `DATABASE_URL`
- `APP_PORT`, `SOCKET_PORT`
- `JWT_ACCESS_KEY`, `JWT_REFRESH_KEY`
- `RABBITMQ_URL`
- `REDIS_URL`
- `GH_APP_SLUG`, `GH_APP_ID`, `GH_APP_CLIENT_ID`, `GH_APP_CLIENT_SECRET`, `GH_PRIVATE_KEY`
- `GH_APP_REDIRECT_URL` (must match your GitHub App callback URL exactly, including trailing slash; default `http://localhost:3000/api/v1/callback/github/`)
- `DEFAULT_ADMIN_EMAIL`, `DEFAULT_ADMIN_PASS`
- `K8S_CLUSTER_CONFIG`
- `INGRESS_ROOT_DOMAIN`
- optional: `SECRET_FROM` (`local` by default, `aws` to load from AWS Secrets Manager)

## Running Locally

### 1) Start dependencies

```bash
docker compose up -d
```

This starts:
- PostgreSQL on `5432`
- RabbitMQ on `5672` (management UI on `15672`)
- Redis on `6379`

### 2) Install dependencies

```bash
go mod download
```

### 3) Run database migrations and Prisma client generation

```bash
go run github.com/steebchen/prisma-client-go migrate deploy
go run github.com/steebchen/prisma-client-go generate
```

### 4) Run the API

```bash
go run .
```

The app listens on `0.0.0.0:${APP_PORT}`.
The WebSocket server listens on `0.0.0.0:${SOCKET_PORT}`.

## API Monitor TUI

Interactive terminal client ([Charm Bubble Tea](https://github.com/charmbracelet/bubbletea) + [huh](https://github.com/charmbracelet/huh)) under `cmd/tui`:

```bash
go run ./cmd/tui
```

Defaults: HTTP `http://localhost:3000`, WebSocket `ws://localhost:9090/ws`.

**The TUI is a complete, end-to-end integration reference for how clients should use the idp API.** It exercises the same REST routes, auth flow, GitHub linking, repo connections, deployments, and WebSocket channels that a production frontend would—use it as the canonical map of API usage when building your own client.

### What it covers

| Section | Operations |
|---------|------------|
| **Auth+User** | Register, login, refresh, logout, profile read/update |
| **Machine+Network** | Machines CRUD (create uses an interactive form), networks list/create/delete |
| **Repo+Deploy** | Repo connections CRUD, list deployments, deploy, deployment detail/logs |
| **GitHub** | List repos, authorize account, update app access, list account connections |
| **WebSocket** | Subscribe / disconnect from event channels |

Interactive forms (no raw JSON for common flows) include machine create, network create, repo connection create/update, and deploy repo—with scrollable selects for machines and GitHub repos where applicable.

### Auth and tokens

- After login or register, access and refresh tokens are saved to `~/.idp-tui-auth.json` and sent as `Authorization: Bearer …` on protected routes.
- On login, the TUI auto-subscribes to `deployment-log-stream-event`.
- **Deploy Repo** also auto-subscribes and streams deployment log **messages** in the output panel (log text only, not raw WebSocket JSON).

Press **`o`** to open links returned by the API (for example GitHub authorize URLs).

### Typical workflow

1. Start the API and dependencies (see above).
2. **Auth+User → Login** (default admin from `.env`: `admin@idp.local` / `admin123`).
3. **GitHub → Authorize Account Link**, then press **`o`** to complete OAuth in the browser.
4. **Machine+Network → Create Machine** — pick `struckchure/alpine` or `struckchure/ubuntu` (required base images with git/SSH for deploy).
5. **Repo+Deploy → Create Repo Connection** — select machine + GitHub repo.
6. Ensure the target repo has a workflow file at `.formatio/action.yaml` or `.idp/action.yaml` (Storm/GitHub Actions-style `jobs` are supported).
7. **Repo+Deploy → Deploy Repo** — select machine, repo connection, optional git ref; watch logs stream in the output panel.

Machines must finish provisioning (non-empty `containerId` in the database / K8s deployment exists) before deploy will succeed.

### Controls

| Key | Action |
|-----|--------|
| `shift+←` / `shift+→` | Previous / next section tab |
| `↑` / `↓` | Select action in the current section |
| `tab` / `shift+tab` | Next / previous config field (HTTP ↔ WS) or form field |
| `↑` / `↓` (in selects) | Browse dropdown options |
| `enter` | Run selected action / submit interactive form |
| `ctrl+s` | Force-submit current form |
| `o` | Open latest link from API output |
| `?` | Toggle in-app usage guide |
| `ctrl+l` | Clear output |
| `q` / `ctrl+c` | Quit |

Press `?` inside the TUI for the full guide, including WebSocket event names.

### Machine images

Only these container images are accepted when creating a machine:

- `struckchure/alpine` (default in the TUI)
- `struckchure/ubuntu`

Other images are rejected by the API with `400`.

## Helpful Task Commands

If you use `task`:

```bash
task dev                 # run app
task db:diff             # create new migration and regenerate client
task db:migrate          # apply migrations
task db:generate         # regenerate Prisma client
task test                # run service tests
task doc:generate        # regenerate swagger docs
```

## Testing

Run service and internals tests:

```bash
go test ./services/... ./internals/... -v
```

Run TUI UI tests:

```bash
go test ./cmd/tui/ui/... -v
```

Run unit tests directory:

```bash
go test -v ./tests/unit/...
```

## API Endpoints

Base path: `/api/v1`

### Auth

- `POST /auth/register/`
- `POST /auth/login/`
- `POST /auth/refresh-access-token/`

### User

- `GET /user/profile/`
- `PATCH /user/profile/`

### Machines

- `GET /machine/`
- `POST /machine/` — body includes `machineName`, `cpu`, `memory`, `machineImage` (`struckchure/alpine` or `struckchure/ubuntu`)
- `GET /machine/:machineId`
- `PATCH /machine/:machineId`
- `DELETE /machine/:machineId`

Machine create is async (RabbitMQ `create-machine-queue`); provisioning sets `containerId` to the K8s deployment name when ready.

### Networks

- `GET /network/`
- `POST /network/`
- `DELETE /network/:networkId`

### GitHub

- `GET /gh/repos`
- `GET /gh/authorize`
- `GET /gh/update-app-access`
- `GET /gh/account-connections`

### Repo Connections

- `POST /repo-connection/`
- `GET /repo-connection/`
- `GET /repo-connection:connectionId`
- `PATCH /repo-connection:connectionId`
- `DELETE /repo-connection:connectionId`

Note: repo connection detail/update/delete routes currently do not include a `/` before `:connectionId`.

### Deployments

Deploy enqueues async work on `deployment-deploy-repo-queue`. The worker clones the repo, reads `.formatio/action.yaml` or `.idp/action.yaml`, runs Storm on the machine’s K8s pods, and publishes logs to `deployment-log-stream-event`.

- `GET /deployments/`
- `POST /deployments/deploy` — body: `{ "connectionId": "...", "ref": "main" }` (optional `ref`)
- `GET /deployments/:deploymentId`
- `GET /deployments/:deploymentId/logs`

### Callback

- `GET /callback/github/` — GitHub OAuth callback (302 redirect); `state` and `code` query params

Set `GH_APP_REDIRECT_URL` in `.env` to this URL and register the same URL on your GitHub App.

### Webhook

- `POST /webhook/github/`

## WebSocket

Connect to:

```text
ws://localhost:${SOCKET_PORT}/ws?event=<event-name>
```

Current event channels:

- `deployment-log-stream-event` — deployment log lines (WebSocket payload: `{ "event", "content" }` where `content.message` is the log text)
- `deployment-notification-event/<machineId>` — deployment lifecycle notifications

Log messages are also available via `GET /deployments/:deploymentId/logs`.

## Production Docker Build

Build the image (no secrets baked in at build time):

```bash
docker build -t idp .
```

Run with your own environment (for example from `.env`):

```bash
docker run --rm -p 3000:3000 -p 9090:9090 --env-file .env idp
```

Apply database migrations before or alongside the first deploy (requires `DATABASE_URL` from the same env file):

```bash
go run github.com/steebchen/prisma-client-go migrate deploy
```

Or run migrations from a one-off container that mounts the same `--env-file` if you add a migrate entrypoint later.
