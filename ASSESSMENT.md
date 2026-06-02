# Backend Developer Assessment

## Assessment Objective

Build a **Mini Deployment Platform** in Go: a production-oriented backend where users authenticate, provision machines, connect GitHub repositories, and deploy code asynchronously. Your submission will be evaluated on feature completeness, clean architecture, concurrency-safe async workflows, and test coverage—not on running real Kubernetes (you may simulate provision and deploy via background jobs).

## Stack

- **Language:** Go 1.22+
- **HTTP framework:** Fiber, Gin, or Chi (your choice)
- **Database:** PostgreSQL (required)
- **Queue:** RabbitMQ preferred; in-memory queue acceptable behind a clear abstraction
- **Data access:** Prisma Go, GORM, SQLC, pgx, or similar (your choice)
- **Local dependencies:** Docker Compose for PostgreSQL and optional RabbitMQ

## Project Tasks

Implement the features below. You choose route names, request/response shapes, and package layout; each feature lists **acceptance criteria** that must hold.

### Section 1 – Authentication & user account

Keep authentication and profile concerns logically separated in code (e.g. `/auth` vs `/user`), even if they share middleware.

#### Feature: Register and sign in

Users can create an account and obtain tokens to access protected APIs.

- Registration accepts required identity fields (e.g. email, password, name—define your schema).
- Login validates credentials and returns access (and optionally refresh) tokens.
- Protected routes require a valid token; invalid or missing tokens return consistent error responses.
- Token refresh (optional but encouraged): exchange a refresh token for a new access token without re-login.

#### Feature: Manage user profile

Authenticated users can view and update their own account details.

- Get the current user’s profile (from the token subject / user ID).
- Update allowed profile fields (e.g. name; email change optional).
- Profile endpoints are scoped to the authenticated user only—no cross-user reads or writes.

### Section 2 – Platform & infrastructure features

#### Feature: Provision and manage machines

Users can create **machines** (compute targets) they own and follow their lifecycle.

- Create a machine with name, image/sizing metadata (fields are yours to define).
- List, view, update, and soft-delete or disable machines the user owns.
- Machine status reflects provisioning progress (e.g. `creating` → `running`; allow `failed` if provision fails).
- **Provision** is asynchronous: creating a machine enqueues work; a worker (or simulated provisioner) transitions status and stores any provision result identifier you need for later deploys.
- Users cannot access another user’s machines.

#### Feature: Configure machine networking (optional / stretch)

Users can attach **networks** to their machines for connectivity configuration.

- Create and list networks scoped to a machine the user owns.
- Remove a network the user owns.

*Optional: candidates may skip this feature if time-boxed; it is not required to pass.*

#### Feature: Connect GitHub and link repositories

Users can connect GitHub and associate repos with a machine before deploying.

- OAuth or GitHub App–style connect flow (simplified mock acceptable if documented: e.g. store installation/token in DB after a callback).
- List repositories available to the connected account (can be stubbed against the GitHub API or fixture data if credentials are not available).
- **Repo connection:** link an `owner/repo` (and default branch) to a specific machine; list, update, and delete connections the user owns.

### Section 3 – Deployment pipeline features

#### Feature: Deploy from GitHub

Users can deploy a linked repository to a provisioned machine.

- Trigger deploy with machine + repo connection (and optional commit/branch).
- Validate that the machine is owned, the repo is linked to that machine, and the machine is in a deployable state.
- Enqueue a deploy job; respond immediately with a deployment record in a non-terminal state.
- **Idempotency:** duplicate in-flight deploy requests for the same intent must not spawn unbounded parallel jobs (document your rule in the README).

#### Feature: Track deployments and logs

Users can inspect deployment outcomes and history.

- List deployments for a machine (or user) with **pagination**.
- Get a single deployment including status and metadata (commit, actor, timestamps).
- Persist **deployment logs** (lines/messages) during the job; expose via a paginated HTTP API.
- Status lifecycle at minimum: `queued` → `running` → `success` | `failed` (names may vary if documented).

#### Feature: Process jobs in the background

Provisioning and deploy work run outside the HTTP request path.

- RabbitMQ-backed worker preferred; in-memory queue acceptable behind an interface.
- Worker updates deployment/machine state and appends logs atomically where needed.
- Failures set a terminal failure status and persist an error message in logs.

**Stretch goals (not required to pass):**

- Real-time log streaming (WebSocket or SSE).
- Inbound GitHub webhooks to auto-deploy on push.

### Constraints

- **Ownership:** users may only access their own machines, repo connections, and deployments.
- **Validation:** validate inputs; return consistent HTTP error shapes.
- **Concurrency:** provision, deploy, and status updates must be safe under concurrent requests (no corrupt status; honor your deploy idempotency rule).
- **Configuration:** all secrets and service URLs via environment variables (no hardcoded credentials). Provide `.env.sample`.

## Technical Requirements

- **Clean Architecture:** separate **Handlers**, **Usecases**, and **Repositories** with clear dependency direction.
- **Idiomatic Go:** consistent naming, module boundaries, and error handling.
- **Docker Compose:** PostgreSQL and optional RabbitMQ for local development.
- **Migrations:** versioned DB setup (tool of your choice).
- **Tests:**
  - **Unit tests** for usecases and domain logic.
  - **Integration tests** for HTTP + database (and queue where feasible).

## Deliverables

Submit a **GitHub repository** containing:

- Complete source code
- `README.md` with:
  - architecture overview
  - setup and run instructions
  - environment variables
  - how to exercise each **feature** (curl or API client examples)
  - trade-offs and what you would improve in production
- `.env.sample` and migration/setup scripts
- Postman/Insomnia collection **or** curl examples for required features
- Unit and integration tests

Optionally include OpenAPI/Swagger documentation.

## Evaluation Criteria

| Criterion | What we look for |
| --------- | ---------------- |
| **Code quality** | Idiomatic Go, clear packages, error handling, readability |
| **Functionality** | Required features work end-to-end; optional/stretch features are bonus only |
| **Architecture & design** | Clean layers, queue abstraction, safe concurrent provision and deploy |
| **Testing** | Meaningful coverage of auth, profile, provision, deploy, and status transitions |
