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

Interactive terminal client (Charm [Bubble Tea](https://github.com/charmbracelet/bubbletea)) for exercising the API like a frontend:

```bash
go run ./cmd/tui
```

Defaults: HTTP `http://localhost:3000`, WebSocket `ws://localhost:9090/ws`.

### Quick start

1. Start the API and dependencies (see above).
2. In the TUI, stay on **Auth+User** and select **Login**.
3. **A** = email, **B** = password, press **Enter**.
4. Tokens are stored in memory automatically; all protected routes send `Authorization: Bearer …`.
5. After login, the TUI auto-subscribes to `deployment-log-stream-event` and streams messages in the output panel.

### Controls

| Key | Action |
|-----|--------|
| `shift+←` / `shift+→` | Previous / next section tab |
| `↑` / `↓` | Select action in the current section |
| `tab` / `shift+tab` | Next / previous input (A → B → JSON → HTTP → WS) |
| `enter` | Run selected action |
| `?` | Toggle in-app usage guide |
| `ctrl+l` | Clear output |
| `q` | Quit |

Press `?` inside the TUI for the full guide, including WebSocket event names.

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

Run service tests:

```bash
go test ./services -v
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
- `POST /machine/`
- `GET /machine/:machineId`
- `PATCH /machine/:machineId`
- `DELETE /machine/:machineId`

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

- `GET /deployments/`
- `POST /deployments/deploy`
- `GET /deployments/:deploymentId`
- `GET /deployments/:deploymentId/logs`

### Callback

- `GET /callback/github/`

### Webhook

- `POST /webhook/github/`

## WebSocket

Connect to:

```text
ws://localhost:${SOCKET_PORT}/ws?event=<event-name>
```

Current event channels:

- `deployment-notification-event/<machineId>`
- `deployment-log-stream-event`

## Production Docker Build

```bash
docker build -t idp . \
  --target production \
  --build-arg infisical_token=<infisical_token> \
  --build-arg infisical_project_id=<infisical_project_id> \
  --build-arg infisical_env=<infisical_env>
```
