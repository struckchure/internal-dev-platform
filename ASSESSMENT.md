# Backend Technical Assessment Spec

## Goal

Design a take-home assessment that evaluates whether a candidate can build a production-oriented Go backend with authenticated APIs, business-service orchestration, database persistence, and one asynchronous workflow.

## Candidate Prompt (What To Build)

Build a Go backend service for a **Mini Deployment Platform**.

### Required Capabilities (High Level)

- **Authentication & authorization**
  - User registration/login with token-based access for protected resources.
  - Basic ownership/permission checks for user-scoped data.
- **Core domain management**
  - CRUD-like management for a primary resource (e.g., machines/projects/services).
  - Resource lifecycle states (e.g., draft/active/disabled).
- **Asynchronous workflow**
  - Trigger an async operation from the API (e.g., deployment/build/provision task).
  - Background worker processes jobs and persists status/log history.

### Architecture & Organization Expectations

Use clear backend engineering best practices for project structure and separation of concerns.

- Organize the codebase in a way that is idiomatic, maintainable, and easy to navigate.
- Keep transport, business logic, data access, and infrastructure concerns logically separated.
- Apply consistent naming, module boundaries, and dependency direction.
- Document major architectural decisions and trade-offs in the README.

### Data Model (minimum)

- `users`: credentials/profile and ownership context
- primary resource table (e.g., `machines`) linked to user
- async job table (e.g., `deployments`) with status lifecycle
- job log/history table for traceability

### Behavior Expectations

- Owner-only access to machine and deployment resources.
- Input validation and consistent error responses.
- Idempotent-ish deploy trigger handling (reasonable safeguards against duplicate in-flight jobs).
- At least one traceable async state transition: `queued -> running -> success|failed`.

## Technical Constraints

- Language: **Go 1.22+**
- HTTP framework: Fiber / Gin / Chi (candidate choice)
- DB: PostgreSQL (required)
- Queue: RabbitMQ preferred; in-memory queue acceptable with clear abstraction
- ORM/query layer: Prisma Go / GORM / SQLC / pgx (candidate choice)
- Provide Docker Compose for local dependencies (DB, optional queue)

## Deliverables

- Source code in a public/private Git repo
- `README.md` with:
  - architecture overview
  - setup/run instructions
  - env variables
  - API usage examples
  - trade-offs and what would be improved in production
- API collection (Postman/Insomnia) or curl examples
- DB migration/setup scripts
- Tests
